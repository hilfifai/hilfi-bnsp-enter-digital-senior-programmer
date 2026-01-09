// internal/service/backup.service.go
package service

import (
	"api-jellyfish/internal/config"
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupInfo represents information about a backup
type BackupInfo struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type IBackupService interface {
	CreateBackup(ctx context.Context) (string, error)
	RestoreBackup(ctx context.Context, backupPath string) error
	ListBackups(ctx context.Context) ([]BackupInfo, error)
	DeleteOldBackups(ctx context.Context, daysToKeep int) error
}

type BackupService struct {
	dbConfig     config.DatabaseConfig
	backupConfig config.BackupConfig
}

func NewBackupService(dbConfig config.DatabaseConfig, backupConfig config.BackupConfig) IBackupService {
	return &BackupService{
		dbConfig:     dbConfig,
		backupConfig: backupConfig,
	}
}

func (s *BackupService) CreateBackup(ctx context.Context) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupDir := s.backupConfig.Directory
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.zip", timestamp))

	// Create backup directory if not exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create zip file
	zipFile, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Backup database
	if err := s.backupDatabase(ctx, zipWriter); err != nil {
		return "", err
	}

	// Backup important files
	if err := s.backupFiles(ctx, zipWriter); err != nil {
		return "", err
	}

	return backupFile, nil
}

func (s *BackupService) backupDatabase(ctx context.Context, zipWriter *zip.Writer) error {
	// Use pg_dump for PostgreSQL
	// This is a simplified example
	dumpFile := "database_dump.sql"

	// Create dump file in zip
	dumpWriter, err := zipWriter.Create(dumpFile)
	if err != nil {
		return fmt.Errorf("failed to create dump file in zip: %w", err)
	}

	// In a real implementation, you would run pg_dump and pipe to zipWriter
	// For now, we'll write a placeholder
	dumpContent := fmt.Sprintf(
		"-- Database Backup\n-- Time: %s\n-- Database: %s\n",
		time.Now().Format(time.RFC3339),
		s.dbConfig.Name,
	)

	if _, err := dumpWriter.Write([]byte(dumpContent)); err != nil {
		return fmt.Errorf("failed to write dump content: %w", err)
	}

	return nil
}

func (s *BackupService) backupFiles(ctx context.Context, zipWriter *zip.Writer) error {
	// Backup configuration files
	files := []string{".env", "docker-compose.yml"}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}
		fileWriter, err := zipWriter.Create(file)
		if err != nil {
			return fmt.Errorf("failed to create file in zip: %w", err)
		}
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if _, err := fileWriter.Write(content); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}
	}
	return nil
}

func (s *BackupService) ListBackups(ctx context.Context) ([]BackupInfo, error) {
	dir := s.backupConfig.Directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			ID:        info.Name(),
			Filename:  info.Name(),
			Path:      filepath.Join(dir, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}
	return backups, nil
}

func (s *BackupService) DeleteOldBackups(ctx context.Context, daysToKeep int) error {
	backups, err := s.ListBackups(ctx)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)
	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoff) {
			if err := os.Remove(backup.Path); err != nil {
				continue
			}
		}
	}
	return nil
}

func (s *BackupService) RestoreBackup(ctx context.Context, backupPath string) error {
	// Implementation for restoring backup
	return nil
}
