package middleware

import (
	"api-jellyfish/internal/config"
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/logger/v2"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestConfig holds configuration for request middleware
type RequestConfig struct {
	// GenerateRequestID enables automatic request ID generation
	GenerateRequestID bool
	// RequestIDHeader is the header name for request ID
	RequestIDHeader string
	// MaxRequestSize limits the maximum request size in bytes
	MaxRequestSize int64
	// Timeout sets the request timeout duration
	Timeout time.Duration
	// RequiredHeaders lists headers that must be present
	RequiredHeaders []string
	// ValidateContentType enables content type validation
	ValidateContentType bool
}

// DefaultRequestConfig returns a default request configuration
func DefaultRequestConfig() *RequestConfig {
	return &RequestConfig{
		GenerateRequestID:   true,
		RequestIDHeader:     "X-Request-ID",
		MaxRequestSize:      50 * 1024 * 1024, // 50MB
		Timeout:             30 * time.Second,
		RequiredHeaders:     []string{},
		ValidateContentType: false,
	}
}

// RequestInit initializes request context with basic information
func RequestInit() gin.HandlerFunc {
	return RequestInitWithConfig(DefaultRequestConfig())
}

// RequestInitWithConfig initializes request context with custom configuration
func RequestInitWithConfig(reqConfig *RequestConfig) gin.HandlerFunc {
	if reqConfig == nil {
		reqConfig = DefaultRequestConfig()
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Single logger instance for this middleware execution
		log := logger.FromContext(c.Request.Context())

		// Validate request size
		if reqConfig.MaxRequestSize > 0 && c.Request.ContentLength > reqConfig.MaxRequestSize {
			log.Warn().
				Int64("content_length", c.Request.ContentLength).
				Int64("max_size", reqConfig.MaxRequestSize).
				Str("client_ip", c.ClientIP()).
				Msg("Request size exceeds limit")
			helper.APIBadRequest(c, "Request size too large", errors.New("request exceeds maximum allowed size"))
			return
		}

		// Generate or extract request ID
		reqID := extractOrGenerateRequestID(c, reqConfig)
		if reqID == "" {
			helper.APIInternalServerError(c, "Failed to generate request ID", errors.New("request ID generation failed"))
			return
		}

		// Validate required headers
		if len(reqConfig.RequiredHeaders) > 0 {
			for _, header := range reqConfig.RequiredHeaders {
				if c.GetHeader(header) == "" {
					log.Debug().
						Str("missing_header", header).
						Str("request_id", reqID).
						Msg("Required header missing")
					helper.APIBadRequest(c, "Required header missing: "+header, errors.New("missing required header"))
					return
				}
			}
		}

		// Validate content type for POST/PUT requests
		if reqConfig.ValidateContentType {
			if err := validateContentType(c); err != nil {
				log.Debug().
					Err(err).
					Str("request_id", reqID).
					Str("content_type", c.GetHeader("Content-Type")).
					Msg("Invalid content type")
				helper.APIBadRequest(c, "Invalid content type", err)
				return
			}
		}

		// Set context values
		c.Set("request_id", reqID)
		c.Set("version", config.Config.AppVersion)
		c.Set("start-time", start)

		// Add timeout context if configured
		if reqConfig.Timeout > 0 {
			ctx, cancel := context.WithTimeout(c.Request.Context(), reqConfig.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// Set response headers
		c.Writer.Header().Set(reqConfig.RequestIDHeader, reqID)
		c.Writer.Header().Set("X-API-Version", config.Config.AppVersion)

		// Inject request ID into request context for logging
		ctx := logger.WithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		// Log request initialization
		log = logger.FromContext(ctx)
		log.Debug().
			Str("request_id", reqID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("Request initialized")

		c.Next()
	}
}

// extractOrGenerateRequestID extracts request ID from header or generates a new one
func extractOrGenerateRequestID(c *gin.Context, reqConfig *RequestConfig) string {
	// Try to get request ID from header
	reqID := c.GetHeader(reqConfig.RequestIDHeader)

	// Validate existing request ID
	if reqID != "" {
		// Basic validation - should be a valid UUID or alphanumeric string
		if isValidRequestID(reqID) {
			return reqID
		}
		log := logger.FromContext(c.Request.Context())
		log.Warn().
			Str("invalid_request_id", reqID).
			Str("client_ip", c.ClientIP()).
			Msg("Invalid request ID provided, generating new one")
	}

	// Generate new request ID if not provided or invalid
	if reqConfig.GenerateRequestID {
		newUUID := uuid.New()
		return newUUID.String()
	}

	return ""
}

// isValidRequestID validates request ID format
func isValidRequestID(reqID string) bool {
	// Check if it's a valid UUID
	if _, err := uuid.Parse(reqID); err == nil {
		return true
	}

	// Check if it's a valid alphanumeric string (length between 8-64)
	if len(reqID) >= 8 && len(reqID) <= 64 {
		for _, char := range reqID {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '-' || char == '_') {
				return false
			}
		}
		return true
	}

	return false
}

// validateContentType validates request content type
func validateContentType(c *gin.Context) error {
	method := c.Request.Method
	contentType := c.GetHeader("Content-Type")

	// Only validate for methods that should have body
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return nil
	}

	// Skip validation if no content
	if c.Request.ContentLength == 0 {
		return nil
	}

	// Check for valid content types
	validContentTypes := []string{
		"application/json",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
		"text/plain",
	}

	// Extract base content type (without parameters)
	baseContentType := strings.Split(contentType, ";")[0]
	baseContentType = strings.TrimSpace(strings.ToLower(baseContentType))

	for _, valid := range validContentTypes {
		if baseContentType == valid {
			return nil
		}
	}

	return errors.New("unsupported content type: " + contentType)
}

// RequestSizeLimitMiddleware enforces request size limits
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			log := logger.FromContext(c.Request.Context())
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Int64("content_length", c.Request.ContentLength).
				Int64("max_size", maxSize).
				Str("client_ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Msg("Request size limit exceeded")

			helper.APIBadRequest(c, "Request entity too large",
				errors.New("request size exceeds maximum allowed limit"))
			return
		}
		c.Next()
	}
}

// RequestTimeoutMiddleware adds timeout to request context
func RequestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Monitor for timeout
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			log := logger.FromContext(c.Request.Context())
			log.Error().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Dur("timeout", timeout).
				Msg("Request timeout")

			helper.APICustomError(c, 408, "Request timeout", errors.New("request processing timeout"))
			c.Abort()
		}
	}
}

// RequestValidationMiddleware validates common request properties
func RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request method
		allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
		if !contains(allowedMethods, c.Request.Method) {
			helper.APICustomError(c, 405, "Method not allowed",
				errors.New("unsupported HTTP method"))
			return
		}

		// Validate path length
		if len(c.Request.URL.Path) > 2048 {
			helper.APIBadRequest(c, "Request URI too long",
				errors.New("path exceeds maximum length"))
			return
		}

		// Validate query parameter count
		if len(c.Request.URL.Query()) > 100 {
			helper.APIBadRequest(c, "Too many query parameters",
				errors.New("query parameter count exceeds limit"))
			return
		}

		// Validate header count and size
		if len(c.Request.Header) > 50 {
			helper.APIBadRequest(c, "Too many headers",
				errors.New("header count exceeds limit"))
			return
		}

		// Validate total header size
		totalHeaderSize := 0
		for name, values := range c.Request.Header {
			totalHeaderSize += len(name)
			for _, value := range values {
				totalHeaderSize += len(value)
			}
		}
		if totalHeaderSize > 8192 { // 8KB
			helper.APIBadRequest(c, "Headers too large",
				errors.New("total header size exceeds limit"))
			return
		}

		c.Next()
	}
}

// RequestRateLimitMiddleware adds basic rate limiting information to context
func RequestRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add rate limiting headers (implement actual rate limiting logic as needed)
		clientIP := c.ClientIP()

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", "1000")
		c.Header("X-RateLimit-Remaining", "999")
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))

		// Log rate limit info
		log := logger.FromContext(c.Request.Context())
		log.Debug().
			Str("request_id", c.GetString("request_id")).
			Str("client_ip", clientIP).
			Str("path", c.Request.URL.Path).
			Msg("Rate limit info added")

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security-related request validation
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Check for suspicious user agents
		userAgent := c.Request.UserAgent()
		if isSuspiciousUserAgent(userAgent) {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("user_agent", userAgent).
				Str("client_ip", c.ClientIP()).
				Msg("Suspicious user agent detected")
		}

		// Check for suspicious paths
		if isSuspiciousPath(c.Request.URL.Path) {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Msg("Suspicious path detected")
		}

		// Add security context
		c.Set("client_fingerprint", generateClientFingerprint(c))

		c.Next()
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isSuspiciousUserAgent(userAgent string) bool {
	suspiciousAgents := []string{
		"curl",
		"wget",
		"python-requests",
		"bot",
		"crawler",
		"spider",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, suspicious := range suspiciousAgents {
		if strings.Contains(userAgentLower, suspicious) {
			return true
		}
	}
	return false
}

func isSuspiciousPath(path string) bool {
	suspiciousPaths := []string{
		"/.env",
		"/admin",
		"/wp-admin",
		"/config",
		"/backup",
		"/.git",
	}

	pathLower := strings.ToLower(path)
	for _, suspicious := range suspiciousPaths {
		if strings.Contains(pathLower, suspicious) {
			return true
		}
	}
	return false
}

func generateClientFingerprint(c *gin.Context) string {
	// Simple client fingerprinting based on IP and User-Agent
	fingerprint := c.ClientIP() + "|" + c.Request.UserAgent()
	hash, _ := helper.HMACSHA256(fingerprint)
	return hash
}

// RequestMetricsMiddleware adds request metrics to context
func RequestMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Set request start time for metrics
		c.Set("request_start", start)
		c.Set("request_path", c.Request.URL.Path)
		c.Set("request_method", c.Request.Method)

		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		c.Set("request_duration", duration)

		// Log metrics
		log := logger.FromContext(c.Request.Context())
		log.Debug().
			Str("request_id", c.GetString("request_id")).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Int("response_size", c.Writer.Size()).
			Msg("Request metrics")
	}
}
