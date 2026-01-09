package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBInterface defines the interface for database operations
type DBInterface interface {
	// Query methods with context
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row

	// Query methods without context (for backward compatibility)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row

	// Execution methods with context
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)

	// Execution methods without context
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)

	// Prepare methods
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)

	// Transaction methods
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
	Beginx() (*sqlx.Tx, error)

	// Named query methods
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)

	// Utility methods
	Rebind(query string) string
	BindNamed(query string, arg interface{}) (string, []interface{}, error)

	// Connection management
	Close() error
	Ping() error
	PingContext(ctx context.Context) error

	// Database info
	DriverName() string

	// Unsafe methods
	Unsafe() *sqlx.DB
}

// Ensure sqlx.DB implements DBInterface
var _ DBInterface = (*sqlx.DB)(nil)
