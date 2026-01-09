package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("record not found")
	ErrDuplicate       = errors.New("duplicate record")
	ErrForeignKey      = errors.New("foreign key violation")
	ErrCheckViolation  = errors.New("check constraint violation")
	ErrInvalidInput    = errors.New("invalid input")
	ErrNotNull         = errors.New("null value in column that requires not null")
	ErrUndefinedColumn = errors.New("undefined column")
	ErrUnexpected      = errors.New("unexpected database error")
)

// MapPGError normalizes PostgreSQL errors into friendly errors
func MapPGError(err error) error {
	if err == nil {
		return nil
	}

	// sql.ErrNoRows → Not Found
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	// PostgreSQL specific errors (pgx driver)
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		switch pgxErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", ErrDuplicate, pgxErr.Detail)
		case "23503":
			return fmt.Errorf("%w: %s", ErrForeignKey, pgxErr.Detail)
		case "23514":
			return fmt.Errorf("%w: %s", ErrCheckViolation, pgxErr.Message)
		case "22P02":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgxErr.Message)
		case "23502":
			return fmt.Errorf("%w: column %s", ErrNotNull, pgxErr.ColumnName)
		case "42703":
			return fmt.Errorf("%w: %s", ErrUndefinedColumn, pgxErr.Message)
		default:
			return fmt.Errorf("postgres error [%s]: %s", pgxErr.Code, pgxErr.Message)
		}
	}

	return fmt.Errorf("%w: %s", ErrUnexpected, err.Error())
}
