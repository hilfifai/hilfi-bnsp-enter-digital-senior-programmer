package middleware

import (
	"api-jellyfish/internal/common/enum"
	_type "api-jellyfish/internal/common/type"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/logger/v2"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
)

// FileValidationConfig holds configuration for file validation
type FileValidationConfig struct {
	MaxFileSize      int64    // Maximum file size in bytes
	AllowedMimeTypes []string // Allowed MIME types
	AllowedExts      []string // Allowed file extensions
	RequireExtMatch  bool     // Require extension to match MIME type
}

// FieldConfig holds configuration for a specific field
type FieldConfig struct {
	Name       string                // Field name
	Min        int                   // Minimum number of files
	Max        int                   // Maximum number of files
	Required   bool                  // Whether the field is required
	Validation *FileValidationConfig // File validation rules
}

// DefaultImageConfig returns default configuration for image files
func DefaultImageConfig() *FileValidationConfig {
	return &FileValidationConfig{
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/jpg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		AllowedExts: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
		},
		RequireExtMatch: true,
	}
}

// DefaultVideoConfig returns default configuration for video files
func DefaultVideoConfig() *FileValidationConfig {
	return &FileValidationConfig{
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		AllowedMimeTypes: []string{
			"video/mp4",
			"video/avi",
			"video/mov",
			"video/wmv",
			"video/flv",
			"video/webm",
		},
		AllowedExts: []string{
			".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm",
		},
		RequireExtMatch: true,
	}
}

// DefaultDocumentConfig returns default configuration for document files
func DefaultDocumentConfig() *FileValidationConfig {
	return &FileValidationConfig{
		MaxFileSize: 50 * 1024 * 1024, // 50MB
		AllowedMimeTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		AllowedExts: []string{
			".pdf", ".doc", ".docx", ".xls", ".xlsx",
		},
		RequireExtMatch: true,
	}
}

// MultipartFormMiddleware creates middleware for handling multipart form data
func MultipartFormMiddleware(fields []FieldConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Single logger instance for this middleware execution
		log := logger.FromContext(c.Request.Context())

		// Validate input
		if len(fields) == 0 {
			log.Warn().Msg("MultipartFormMiddleware called with no fields")
			helper.APIBadRequest(c, "No file fields configured", errors.New("middleware misconfiguration"))
			return
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			log.Error().
				Err(err).
				Str("content_type", c.GetHeader("Content-Type")).
				Msg("Failed to parse multipart form")
			helper.APIBadRequest(c, "Invalid multipart form data", err)
			return
		}

		if form == nil {
			helper.APIBadRequest(c, "No multipart form data received", errors.New("form is nil"))
			return
		}

		bufferedFiles := make(_type.BufferedFiles)

		// Process each configured field
		for _, fieldConfig := range fields {
			err := processField(c, form.File, fieldConfig, bufferedFiles)
			if err != nil {
				log.Error().
					Err(err).
					Str("field", fieldConfig.Name).
					Msg("Failed to process multipart field")
				return // Error already sent in processField
			}
		}

		// Validate field requirements
		for _, fieldConfig := range fields {
			err := validateFieldRequirements(c, fieldConfig, bufferedFiles)
			if err != nil {
				return // Error already sent in validateFieldRequirements
			}
		}

		// Set processed files in context
		c.Set("bufferedFiles", bufferedFiles)

		log.Debug().
			Interface("fields", getFieldSummary(bufferedFiles)).
			Msg("Multipart form processed successfully")

		c.Next()
	}
}

// processField processes files for a specific field
func processField(c *gin.Context, files map[string][]*multipart.FileHeader, fieldConfig FieldConfig, bufferedFiles _type.BufferedFiles) error {
	fileHeaders := files[fieldConfig.Name]
	if len(fileHeaders) == 0 {
		// No files for this field
		if fieldConfig.Required {
			helper.APIRequiredField(c, fieldConfig.Name)
			return errors.New("required field missing")
		}
		return nil
	}

	// Process each file in the field
	for i, fileHeader := range fileHeaders {
		bufferedFile, err := processFile(fileHeader, fieldConfig)
		if err != nil {
			helper.APIBadRequest(c, fmt.Sprintf("Invalid file in field '%s' at position %d: %s", fieldConfig.Name, i, err.Error()), err)
			return err
		}

		// Add to buffered files
		bufferedFiles[fieldConfig.Name] = append(bufferedFiles[fieldConfig.Name], *bufferedFile)
	}

	return nil
}

// processFile processes a single file
func processFile(fileHeader *multipart.FileHeader, fieldConfig FieldConfig) (*_type.BufferedFile, error) {
	if fileHeader == nil {
		return nil, errors.New("file header is nil")
	}

	// Validate filename
	if fileHeader.Filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	// Validate file size
	if fieldConfig.Validation != nil && fieldConfig.Validation.MaxFileSize > 0 {
		if fileHeader.Size > fieldConfig.Validation.MaxFileSize {
			return nil, fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes",
				fileHeader.Size, fieldConfig.Validation.MaxFileSize)
		}
	}

	// Open and read file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileBuffer, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Get MIME type
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Create buffered file
	bufferedFile := &_type.BufferedFile{
		MediaType:    fieldConfig.Name,
		OriginalName: fileHeader.Filename,
		Encoding:     "binary",
		MimeType:     mimeType,
		Size:         int(fileHeader.Size),
		Buffer:       fileBuffer,
	}

	// Validate file content and type
	if fieldConfig.Validation != nil {
		err := validateFile(bufferedFile, fieldConfig)
		if err != nil {
			return nil, err
		}
	}

	// Additional validation based on field type
	err = validateByFieldType(bufferedFile, fieldConfig.Name)
	if err != nil {
		return nil, err
	}

	return bufferedFile, nil
}

// validateFile validates file based on validation config
func validateFile(bufferedFile *_type.BufferedFile, fieldConfig FieldConfig) error {
	config := fieldConfig.Validation

	// Validate MIME type
	if len(config.AllowedMimeTypes) > 0 {
		allowed := false
		for _, allowedType := range config.AllowedMimeTypes {
			if bufferedFile.MimeType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("MIME type '%s' is not allowed. Allowed types: %v",
				bufferedFile.MimeType, config.AllowedMimeTypes)
		}
	}

	// Validate file extension
	if len(config.AllowedExts) > 0 {
		filename := strings.ToLower(bufferedFile.OriginalName)
		allowed := false
		for _, allowedExt := range config.AllowedExts {
			if strings.HasSuffix(filename, strings.ToLower(allowedExt)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension is not allowed. Allowed extensions: %v", config.AllowedExts)
		}
	}

	return nil
}

// validateByFieldType performs additional validation based on field type
func validateByFieldType(bufferedFile *_type.BufferedFile, fieldType string) error {
	switch fieldType {
	case enum.IMAGE.ToString():
		if !enum.IMAGE.IsValidImage(bufferedFile) {
			return errors.New("invalid image file format or corrupted image")
		}
	case enum.VIDEO.ToString():
		if !enum.VIDEO.IsValidVideo(bufferedFile) {
			return errors.New("invalid video file format or corrupted video")
		}
	}
	return nil
}

// validateFieldRequirements validates min/max requirements for fields
func validateFieldRequirements(c *gin.Context, fieldConfig FieldConfig, bufferedFiles _type.BufferedFiles) error {
	fileCount := len(bufferedFiles[fieldConfig.Name])

	// Check minimum requirement
	if fileCount < fieldConfig.Min {
		message := fmt.Sprintf("Field '%s' requires at least %d file(s), got %d",
			fieldConfig.Name, fieldConfig.Min, fileCount)
		helper.APIBadRequest(c, message, errors.New("minimum file requirement not met"))
		return errors.New("minimum requirement not met")
	}

	// Check maximum requirement
	if fieldConfig.Max > 0 && fileCount > fieldConfig.Max {
		message := fmt.Sprintf("Field '%s' allows at most %d file(s), got %d",
			fieldConfig.Name, fieldConfig.Max, fileCount)
		helper.APIBadRequest(c, message, errors.New("maximum file limit exceeded"))
		return errors.New("maximum requirement exceeded")
	}

	return nil
}

// getFieldSummary creates a summary for logging
func getFieldSummary(bufferedFiles _type.BufferedFiles) map[string]int {
	summary := make(map[string]int)
	for fieldName, files := range bufferedFiles {
		summary[fieldName] = len(files)
	}
	return summary
}

// Convenience functions for common field configurations

// ImageField creates a field configuration for images
func ImageField(name string, min, max int, required bool) FieldConfig {
	return FieldConfig{
		Name:       name,
		Min:        min,
		Max:        max,
		Required:   required,
		Validation: DefaultImageConfig(),
	}
}

// VideoField creates a field configuration for videos
func VideoField(name string, min, max int, required bool) FieldConfig {
	return FieldConfig{
		Name:       name,
		Min:        min,
		Max:        max,
		Required:   required,
		Validation: DefaultVideoConfig(),
	}
}

// DocumentField creates a field configuration for documents
func DocumentField(name string, min, max int, required bool) FieldConfig {
	return FieldConfig{
		Name:       name,
		Min:        min,
		Max:        max,
		Required:   required,
		Validation: DefaultDocumentConfig(),
	}
}

// AnyFileField creates a field configuration for any file type
func AnyFileField(name string, min, max int, required bool, maxFileSize int64) FieldConfig {
	return FieldConfig{
		Name:     name,
		Min:      min,
		Max:      max,
		Required: required,
		Validation: &FileValidationConfig{
			MaxFileSize: maxFileSize,
		},
	}
}

// SingleImageMiddleware is a convenience middleware for single image upload
func SingleImageMiddleware(fieldName string, required bool) gin.HandlerFunc {
	return MultipartFormMiddleware([]FieldConfig{
		ImageField(fieldName, helper.SafeIntFromInterface(required, 1), 1, required),
	})
}

// MultipleImageMiddleware is a convenience middleware for multiple image uploads
func MultipleImageMiddleware(fieldName string, min, max int) gin.HandlerFunc {
	return MultipartFormMiddleware([]FieldConfig{
		ImageField(fieldName, min, max, min > 0),
	})
}
