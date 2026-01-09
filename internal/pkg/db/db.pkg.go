package database

import (
	"api-jellyfish/internal/pkg/helper"
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // driver pgx untuk database/sql
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	SSLMode   string
	URL       string
	Driver    DriverEnum
	Cache     bool
	CacheTime time.Duration
}

type Database struct {
	DB           DBInterface
	Config       *Config
	CursorCrypto *helper.CursorCrypto
}

func Setup(cfg *Config) (*Database, error) {
	var db *sqlx.DB
	var err error

	crypto, err := helper.NewCursorCrypto(cfg.User + cfg.Password + cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("crypto init error: %w", err)
	}

	dsn := cfg.URL
	if dsn == "" {
		switch cfg.Driver {
		case POSTGRES:
			dsn = fmt.Sprintf(
				"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
				cfg.User,
				cfg.Password,
				cfg.Host,
				cfg.Port,
				cfg.Database,
				cfg.SSLMode,
			)
		case MYSQL:
			dsn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&loc=Local",
				cfg.User,
				cfg.Password,
				cfg.Host,
				cfg.Port,
				cfg.Database,
			)
		default:
			return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
		}
	}

	// Connect to DB
	switch cfg.Driver {
	case POSTGRES:
		db, err = sqlx.Connect("pgx", dsn)
	case MYSQL:
		db, err = sqlx.Connect("mysql", dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	// Set connection pool config
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour)

	// NOTE: Caching layer not available directly in sqlx.
	// You must implement manual caching via Redis or middleware if needed.

	return &Database{
		DB:           DBInterface(db),
		Config:       cfg,
		CursorCrypto: crypto,
	}, nil
}

func (db *Database) Close() error {
	return db.DB.Close()
}

// GetSqlxDB returns the underlying *sqlx.DB instance for advanced operations like pagination
// This should be used sparingly and only when the interface doesn't provide the needed functionality
func (db *Database) GetSqlxDB() (*sqlx.DB, error) {
	if sqlxDB, ok := db.DB.(*sqlx.DB); ok {
		return sqlxDB, nil
	}
	return nil, fmt.Errorf("database interface is not *sqlx.DB, got %T", db.DB)
}

func SelectInContext[T any](ctx context.Context, db DBInterface, dest *[]T, baseQuery string, args ...any) error {
	// Gunakan sqlx.In untuk ekspansi slice di IN clause
	query, expandedArgs, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return err
	}

	// Rebind untuk PostgreSQL (ubah ? ke $1, $2, dst)
	query = db.Rebind(query)

	// Jalankan query
	return db.SelectContext(ctx, dest, query, expandedArgs...)
}
