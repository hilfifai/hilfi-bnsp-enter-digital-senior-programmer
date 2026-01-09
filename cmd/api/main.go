package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	config "api-jellyfish/internal/config"
	database "api-jellyfish/internal/pkg/db"
	log "api-jellyfish/internal/pkg/logger"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/middleware"
	"api-jellyfish/internal/pkg/validation"
	serverApp "api-jellyfish/internal/server"
)

const (
	// Server configuration
	shutdownTimeout = 30 * time.Second
	minPort         = 1
	maxPort         = 65535

	// Database configuration
	postgresDriver = "postgres"
	sslModeDisable = "disable"

	// Timezone
	Timezone = "Asia/Jakarta"
)

// setupServer initializes and starts the HTTP server with graceful shutdown
func setupServer(cfg *config.SetupServerDto) {
	ctx := cfg.Ctx
	wg := cfg.Wg
	db := cfg.Db

	// Setup validation
	if err := validation.Setup(); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to setup validation")
		panic(err)
	}

	// Configure Gin mode
	setupGinMode()

	// Initialize Gin engine with middleware
	engine := setupGinEngine()

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.AppPort),
		Handler: engine,
	}

	// Setup application routes and handlers
	serverApp.Setup(engine, *ctx, wg, db)

	// Start server in goroutine
	startServer(server)

	// Wait for shutdown signal and gracefully shutdown
	waitForShutdown(server, *ctx)
}

// setupGinMode configures Gin mode based on debug setting
func setupGinMode() {
	if config.Config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// setupGinEngine creates and configures Gin engine with middleware
func setupGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(
		middleware.RequestInit(),
		middleware.Logging(),
		middleware.ResponseHeaderMiddleware(),
		middleware.Recovery(),
	)
	return engine
}

// startServer starts the HTTP server in a goroutine
func startServer(server *http.Server) {
	go func() {
		logger.Log.Info().Msgf("Starting server on port %d", config.Config.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error().Err(err).Msg("Server failed to start")
			os.Exit(1)
		}
	}()
}

// waitForShutdown waits for shutdown signal and gracefully shuts down the server
func waitForShutdown(server *http.Server, ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	logger.Log.Info().Msg("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		logger.Log.Info().Msg("Server exited gracefully")
	}
}

// setupTimezone configures the application timezone to Asia/Jakarta
func setupTimezone() error {
	loc, err := time.LoadLocation(Timezone)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error loading timezone, falling back to UTC")
		// Fallback to UTC if timezone loading fails
		loc = time.UTC
	}
	time.Local = loc
	return nil
}

// @title						General API
// @version					1.0
// @description				API documentations for the BE.
// @host						localhost:9000
// @BasePath					/api
// @schemes					http https
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	// Load application configuration
	config.LoadConfig()

	// Initialize logging systems
	setupLogging()

	// Configure timezone
	if err := setupTimezone(); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to setup timezone")
	}

	// Create context for graceful shutdown
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup database connection
	db, err := setupDB()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to setup database")
		return
	}

	// Log successful initialization
	logger.Log.Info().Msg("Application initialized successfully")
	// Ensure database connection is closed on exit
	defer func() {
		if db != nil {
			logger.Log.Info().Msg("Closing database connection...")
			// Note: Add db.Close() if your database package supports it
		}
	}()

	// Setup and start server
	setupServer(&config.SetupServerDto{
		Ctx:    &ctx,
		Cancel: cancel,
		Db:     db,
		Wg:     &wg,
	})
}

// setupLogging initializes both old and new logging systems
func setupLogging() {
	// Support old logging for backward compatibility
	log.Setup()

	// Initialize new structured logging
	logger.Init(logger.Config{
		Service: config.Config.AppName,
		Env:     config.Config.AppEnvironment,
		Version: config.Config.AppVersion,
		Pretty:  config.Config.Debug,
		Level:   config.Config.LogLevel,
	})
}

// setupDB initializes and returns database connection
func setupDB() (*database.Database, error) {
	dbConfig := &database.Config{
		Host:     config.Config.Database.Host,
		Port:     config.Config.Database.Port,
		User:     config.Config.Database.User,
		Password: config.Config.Database.Pass,
		Database: config.Config.Database.Name,
		SSLMode:  sslModeDisable,
		Driver:   postgresDriver,
	}

	// Validate database configuration
	if err := validateDBConfig(dbConfig); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	logger.Log.Info().Msgf("Connecting to database: %s at %s:%d",
		dbConfig.Database, dbConfig.Host, dbConfig.Port)

	return database.Setup(dbConfig)
}

// validateDBConfig validates database configuration parameters
func validateDBConfig(cfg *database.Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Port < minPort || cfg.Port > maxPort {
		return fmt.Errorf("invalid database port: %d (must be between %d-%d)", cfg.Port, minPort, maxPort)
	}
	if cfg.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if cfg.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}
