package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func (db *Database) WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := db.DB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("transaction panic: %v", r)
			// Consider logging the panic stack trace here
		}
	}()

	if err = fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		// Attempt to rollback on commit failure, although it may not always succeed.
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("commit err: %v, rb err: %v", err, rbErr)
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
