// internal/repository/goods_receipt.repository.go
package repository

import (
	"api-jellyfish/internal/entity"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IGoodsReceiptRepository interface {
	Create(ctx context.Context, receipt *entity.GoodsReceipt) error
	CreateItem(ctx context.Context, item *entity.GoodsReceiptItem) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error)
	FindByGRNumber(ctx context.Context, grNumber string) (*entity.GoodsReceipt, error)
	FindByPurchaseOrderID(ctx context.Context, purchaseOrderID uuid.UUID) ([]entity.GoodsReceipt, error)
	Update(ctx context.Context, id uuid.UUID, receipt *entity.GoodsReceipt) error
	GetItemsByReceiptID(ctx context.Context, receiptID uuid.UUID) ([]entity.GoodsReceiptItem, error)
	GetNextGRNumber(ctx context.Context) (string, error)
	BeginTx(ctx context.Context) (interface{}, error)
	CreateTx(ctx context.Context, tx interface{}, receipt *entity.GoodsReceipt) error
	CreateItemTx(ctx context.Context, tx interface{}, item *entity.GoodsReceiptItem) error
	UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, receipt *entity.GoodsReceipt) error
}

type GoodsReceiptRepository struct {
	db *sqlx.DB
}

func NewGoodsReceiptRepository(db *sqlx.DB) IGoodsReceiptRepository {
	return &GoodsReceiptRepository{db: db}
}

func (r *GoodsReceiptRepository) Create(ctx context.Context, receipt *entity.GoodsReceipt) error {
	query := `
		INSERT INTO goods_receipts (
			id, gr_number, purchase_order_id, warehouse_id, receipt_date, status,
			notes, received_by, approved_by, created_at, updated_at
		) VALUES (
			:id, :gr_number, :purchase_order_id, :warehouse_id, :receipt_date, :status,
			:notes, :received_by, :approved_by, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, receipt)
	return err
}

func (r *GoodsReceiptRepository) CreateItem(ctx context.Context, item *entity.GoodsReceiptItem) error {
	query := `
		INSERT INTO goods_receipt_items (
			id, goods_receipt_id, purchase_order_item_id, received_qty, location_id,
			batch_number, expiry_date, notes, created_at, updated_at
		) VALUES (
			:id, :goods_receipt_id, :purchase_order_item_id, :received_qty, :location_id,
			:batch_number, :expiry_date, :notes, :created_at, :updated_at
		)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *GoodsReceiptRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.GoodsReceipt, error) {
	var receipt entity.GoodsReceipt
	query := `SELECT * FROM goods_receipts WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &receipt, query, id)
	return &receipt, err
}

func (r *GoodsReceiptRepository) FindByGRNumber(ctx context.Context, grNumber string) (*entity.GoodsReceipt, error) {
	var receipt entity.GoodsReceipt
	query := `SELECT * FROM goods_receipts WHERE gr_number = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &receipt, query, grNumber)
	return &receipt, err
}

func (r *GoodsReceiptRepository) FindByPurchaseOrderID(ctx context.Context, purchaseOrderID uuid.UUID) ([]entity.GoodsReceipt, error) {
	var receipts []entity.GoodsReceipt
	query := `SELECT * FROM goods_receipts WHERE purchase_order_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &receipts, query, purchaseOrderID)
	return receipts, err
}

func (r *GoodsReceiptRepository) Update(ctx context.Context, id uuid.UUID, receipt *entity.GoodsReceipt) error {
	query := `
		UPDATE goods_receipts
		SET status = :status, notes = :notes, updated_at = :updated_at
		WHERE id = :id`
	receipt.ID = id
	_, err := r.db.NamedExecContext(ctx, query, receipt)
	return err
}

func (r *GoodsReceiptRepository) GetItemsByReceiptID(ctx context.Context, receiptID uuid.UUID) ([]entity.GoodsReceiptItem, error) {
	var items []entity.GoodsReceiptItem
	query := `SELECT * FROM goods_receipt_items WHERE goods_receipt_id = $1 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &items, query, receiptID)
	return items, err
}

func (r *GoodsReceiptRepository) GetNextGRNumber(ctx context.Context) (string, error) {
	var lastNumber string
	query := `SELECT gr_number FROM goods_receipts 
	          WHERE gr_number LIKE 'GR-' || to_char(CURRENT_DATE, 'YYYYMMDD') || '-%' 
	          ORDER BY gr_number DESC LIMIT 1`

	err := r.db.GetContext(ctx, &lastNumber, query)
	if err != nil {
		return "GR-" + time.Now().Format("20060102") + "-0001", nil
	}

	// Extract the sequence number and increment
	parts := strings.Split(lastNumber, "-")
	if len(parts) == 3 {
		seq, err := strconv.Atoi(parts[2])
		if err == nil {
			return fmt.Sprintf("GR-%s-%04d", time.Now().Format("20060102"), seq+1), nil
		}
	}

	return "GR-" + time.Now().Format("20060102") + "-0001", nil
}

func (r *GoodsReceiptRepository) BeginTx(ctx context.Context) (interface{}, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *GoodsReceiptRepository) CreateTx(ctx context.Context, tx interface{}, receipt *entity.GoodsReceipt) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO goods_receipts (
			id, gr_number, purchase_order_id, warehouse_id, receipt_date, status,
			notes, received_by, approved_by, created_at, updated_at
		) VALUES (
			:id, :gr_number, :purchase_order_id, :warehouse_id, :receipt_date, :status,
			:notes, :received_by, :approved_by, :created_at, :updated_at
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, receipt)
	return err
}

func (r *GoodsReceiptRepository) CreateItemTx(ctx context.Context, tx interface{}, item *entity.GoodsReceiptItem) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		INSERT INTO goods_receipt_items (
			id, goods_receipt_id, purchase_order_item_id, received_qty, location_id,
			batch_number, expiry_date, notes, created_at, updated_at
		) VALUES (
			:id, :goods_receipt_id, :purchase_order_item_id, :received_qty, :location_id,
			:batch_number, :expiry_date, :notes, :created_at, :updated_at
		)`
	_, err := sqlxTx.NamedExecContext(ctx, query, item)
	return err
}

func (r *GoodsReceiptRepository) UpdateTx(ctx context.Context, tx interface{}, id uuid.UUID, receipt *entity.GoodsReceipt) error {
	sqlxTx, ok := tx.(*sqlx.Tx)
	if !ok {
		return errors.New("invalid transaction")
	}
	query := `
		UPDATE goods_receipts
		SET status = :status, notes = :notes, updated_at = :updated_at
		WHERE id = :id`
	receipt.ID = id
	_, err := sqlxTx.NamedExecContext(ctx, query, receipt)
	return err
}
