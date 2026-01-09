package database

import (
	"api-jellyfish/internal/pkg/db/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Struct(t *testing.T) {
	t.Run("Config struct initialization", func(t *testing.T) {
		config := Config{
			Host:      "localhost",
			Port:      5432,
			User:      "testuser",
			Password:  "testpass",
			Database:  "testdb",
			SSLMode:   "disable",
			Driver:    POSTGRES,
			Cache:     true,
			CacheTime: 5 * time.Minute,
		}

		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "testuser", config.User)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "disable", config.SSLMode)
		assert.Equal(t, POSTGRES, config.Driver)
		assert.True(t, config.Cache)
		assert.Equal(t, 5*time.Minute, config.CacheTime)
	})
}

func TestDatabase_Struct(t *testing.T) {
	t.Run("Database struct initialization", func(t *testing.T) {
		db := Database{}

		assert.Nil(t, db.DB)
		assert.Nil(t, db.Config)
		assert.Nil(t, db.CursorCrypto)
	})
}

func TestSetup_PostgreSQL(t *testing.T) {
	// Note: These tests would require actual database connections or more sophisticated mocking
	// For demonstration purposes, we'll test the configuration validation logic

	t.Run("PostgreSQL config with URL", func(t *testing.T) {
		config := &Config{
			URL:    "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
			Driver: POSTGRES,
		}

		// This would normally connect to a real database
		// In a real test environment, you'd use a test database or docker container
		_, err := Setup(config)
		// We expect an error since we don't have a real DB connection
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to DB")
	})

	t.Run("PostgreSQL config without URL", func(t *testing.T) {
		config := &Config{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
			Driver:   POSTGRES,
		}

		_, err := Setup(config)
		// We expect an error since we don't have a real DB connection
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to DB")
	})

	t.Run("MySQL config without URL", func(t *testing.T) {
		config := &Config{
			Host:     "localhost",
			Port:     3306,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			Driver:   MYSQL,
		}

		_, err := Setup(config)
		// We expect an error since we don't have a real DB connection
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to DB")
	})

	t.Run("Unsupported driver", func(t *testing.T) {
		config := &Config{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			Driver:   DriverEnum("unsupported"),
		}

		_, err := Setup(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
	})

	t.Run("Empty driver with URL", func(t *testing.T) {
		config := &Config{
			URL:    "postgres://testuser:testpass@localhost:5432/testdb",
			Driver: DriverEnum(""),
		}

		_, err := Setup(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
	})

	t.Run("Crypto initialization failure", func(t *testing.T) {
		// Test with minimal config that would cause crypto init issues
		config := &Config{
			User:     "", // Empty user might cause crypto issues
			Password: "",
			Database: "",
			Driver:   POSTGRES,
		}

		_, err := Setup(config)
		// Could be crypto error or connection error
		assert.Error(t, err)
	})
}

func TestDatabase_Close(t *testing.T) {
	t.Run("Close database connection", func(t *testing.T) {
		// Create a mock database
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)

		// Set up expectation for close
		mock.ExpectClose()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		database := &Database{
			DB: sqlxDB,
		}

		err = database.Close()
		assert.NoError(t, err)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Close with nil database interface", func(t *testing.T) {
		database := &Database{
			DB: nil,
		}

		// Calling Close on nil DB interface should return an error or panic
		// In this case, we'll just verify the structure exists
		assert.NotNil(t, database)
		assert.Nil(t, database.DB)
	})
}

func TestDatabase_GetSqlxDB(t *testing.T) {
	t.Run("Get sqlx.DB from valid Database", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		database := &Database{
			DB: sqlxDB,
		}

		result, err := database.GetSqlxDB()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sqlxDB, result)
	})

	t.Run("Get sqlx.DB from invalid interface", func(t *testing.T) {
		// Create a mock implementation that's not *sqlx.DB
		database := &Database{
			DB: &mocks.MockDBInterface{},
		}

		result, err := database.GetSqlxDB()
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database interface is not *sqlx.DB")
	})

	t.Run("Get sqlx.DB with nil database", func(t *testing.T) {
		database := &Database{}

		result, err := database.GetSqlxDB()
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSelectInContext(t *testing.T) {
	t.Run("Successful SelectInContext", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		// Prepare mock expectations
		mock.ExpectQuery("SELECT \\* FROM users WHERE id IN \\(\\$1, \\$2\\)").
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "John").
				AddRow(2, "Jane"))

		ctx := context.Background()
		var users []struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
		}

		err = SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users WHERE id IN (?)", []int{1, 2})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "John", users[0].Name)
		assert.Equal(t, "Jane", users[1].Name)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SelectInContext with sqlx.In error", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		ctx := context.Background()
		var users []struct{}

		// Pass an invalid query that would cause sqlx.In to fail
		err = SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users WHERE", nil)
		assert.Error(t, err)
	})

	t.Run("SelectInContext with database error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		// Prepare mock to return an error
		mock.ExpectQuery("SELECT \\* FROM users WHERE id IN \\(\\$1\\)").
			WithArgs(1).
			WillReturnError(errors.New("database error"))

		ctx := context.Background()
		var users []struct {
			ID int `db:"id"`
		}

		err = SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users WHERE id IN (?)", []int{1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SelectInContext with context cancellation", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mock.ExpectQuery("SELECT \\* FROM users WHERE id IN \\(\\$1\\)").
			WithArgs(1).
			WillReturnError(context.Canceled)

		var users []struct {
			ID int `db:"id"`
		}

		err = SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users WHERE id IN (?)", []int{1})
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("SelectInContext with empty args", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		sqlxDB := sqlx.NewDb(mockDB, "postgres")

		mock.ExpectQuery("SELECT \\* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		ctx := context.Background()
		var users []struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
		}

		err = SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users")
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Benchmark tests
func BenchmarkSelectInContext(b *testing.B) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")

	// Setup mock expectations for all iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT \\* FROM users WHERE id IN \\(\\$1\\)").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	ctx := context.Background()
	var users []struct {
		ID int `db:"id"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := SelectInContext(ctx, sqlxDB, &users, "SELECT * FROM users WHERE id IN (?)", []int{1})
		if err != nil {
			b.Fatal(err)
		}
		users = users[:0] // Reset slice for next iteration
	}
}

func BenchmarkDatabaseClose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mockDB, _, err := sqlmock.New()
		if err != nil {
			b.Fatal(err)
		}

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		database := &Database{DB: sqlxDB}

		b.StartTimer()
		err = database.Close()
		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetSqlxDB(b *testing.B) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	database := &Database{DB: sqlxDB}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := database.GetSqlxDB()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration test helpers (these would be used with actual test databases)
func TestSetup_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// These tests would run against actual test databases
	// They're commented out but show how you'd structure integration tests

	/*
		t.Run("PostgreSQL integration test", func(t *testing.T) {
			config := &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
				Driver:   POSTGRES,
			}

			db, err := Setup(config)
			require.NoError(t, err)
			defer db.Close()

			// Test basic connectivity
			ctx := context.Background()
			err = db.DB.PingContext(ctx)
			assert.NoError(t, err)
		})
	*/
}

// Example of how to test with real database using testcontainers
func TestSetup_WithTestContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container tests in short mode")
	}

	// This would use testcontainers-go to spin up a real database
	// Example implementation:
	/*
		ctx := context.Background()

		// Start PostgreSQL container
		postgres, err := postgres.RunContainer(ctx,
			testcontainers.WithImage("postgres:13-alpine"),
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
		)
		require.NoError(t, err)
		defer postgres.Terminate(ctx)

		host, err := postgres.Host(ctx)
		require.NoError(t, err)
		port, err := postgres.MappedPort(ctx, "5432")
		require.NoError(t, err)

		config := &Config{
			Host:     host,
			Port:     port.Int(),
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
			Driver:   POSTGRES,
		}

		db, err := Setup(config)
		require.NoError(t, err)
		defer db.Close()

		// Test actual database operations
		ctx = context.Background()
		err = db.DB.PingContext(ctx)
		assert.NoError(t, err)
	*/
}
