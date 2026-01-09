package config

import (
	"context"
	"log"
	"sync"

	"api-jellyfish/internal/pkg/helper"

	database "api-jellyfish/internal/pkg/db"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppName        string
	AppPort        int
	AppEnvironment string
	AppVersion     string
	Debug          bool
	LogLevel       int
	Database       DatabaseConfig
	Backup         BackupConfig
	JwtSecret      string
	AppUrl         string
	AppPortStr     string
	AppSwagger     bool
}

type SetupServerDto struct {
	Ctx    *context.Context
	Cancel context.CancelFunc
	Db     *database.Database
	Wg     *sync.WaitGroup
}

type DatabaseConfig struct {
	Host    string
	Port    int
	User    string
	Pass    string
	Name    string
	Type    string
	Driver  string
	SSLMode string
	URL     string
}

type BackupConfig struct {
	Directory string
	Retention int
}

var Config AppConfig

func LoadConfig() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse and map to struct
	Config = AppConfig{
		AppName:        helper.GetEnvDefault("APP_NAME", "GO Application"),
		AppPort:        helper.GetEnvAsInt("APP_PORT", 8080),
		AppEnvironment: helper.GetEnvDefault("APP_ENV", "development"),
		AppVersion:     helper.GetEnvDefault("APP_VERSION", "1.0.0"),
		Debug:          helper.GetEnvAsBool("APP_DEBUG", false),
		LogLevel:       helper.GetEnvAsInt("APP_LOG_LEVEL", 1),
		// LogLevel:       helper.GetEnvAsInt("LOG_LEVEL", 4), // default ke debug
		// karena di zerolog levelnya 4 itu debug, 1 itu panic
		// jadi kalau mau production set ke 1 atau 2
		AppUrl:     helper.GetEnvDefault("APP_URL", "http://localhost:8080"),
		AppPortStr: helper.GetEnvDefault("APP_PORT", "8080"),
		JwtSecret:  helper.GetEnvDefault("JWT_SECRET", ""),
		AppSwagger: helper.GetEnvAsBool("APP_SWAGGER", true),
		Database: DatabaseConfig{
			Host:    helper.GetEnvDefault("DB_HOST", "localhost"),
			Port:    helper.GetEnvAsInt("DB_PORT", 5432),
			User:    helper.GetEnvDefault("DB_USER", "postgres"),
			Pass:    helper.GetEnvDefault("DB_PASS", "secret"),
			Name:    helper.GetEnvDefault("DB_NAME", "be"),
			Type:    helper.GetEnvDefault("DB_TYPE", "postgres"),
			Driver:  helper.GetEnvDefault("DB_DRIVER", "pgx"),
			SSLMode: helper.GetEnvDefault("DB_SSL_MODE", "require"),
			URL:     helper.GetEnvDefault("DB_URL", "postgres://user:password@host:port/dbname"),
		},
		Backup: BackupConfig{
			Directory: helper.GetEnvDefault("BACKUP_DIRECTORY", "./backups"),
			Retention: helper.GetEnvAsInt("BACKUP_RETENTION", 7),
		},
	}
}
