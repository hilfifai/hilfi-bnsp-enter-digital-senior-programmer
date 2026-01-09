package middleware

import (
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/logger/v2"
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingConfig holds configuration for logging middleware
type LoggingConfig struct {
	// SkipPaths contains paths that should not be logged
	SkipPaths []string
	// LogRequestBody enables request body logging (be careful with sensitive data)
	LogRequestBody bool
	// LogResponseBody enables response body logging (be careful with large responses)
	LogResponseBody bool
	// MaxBodySize limits the size of body to log
	MaxBodySize int
	// LogLevel determines the log level for different response codes
	LogLevel map[int]string
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/ping",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024, // 1KB
		LogLevel: map[int]string{
			200: "info",  // 2xx success
			300: "info",  // 3xx redirection
			400: "warn",  // 4xx client errors
			500: "error", // 5xx server errors
		},
	}
}

// responseWriter wraps gin.ResponseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.body != nil {
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Status() int {
	if w.statusCode == 0 {
		return 200 // Default status code
	}
	return w.statusCode
}

// Logging creates a logging middleware with default configuration
func Logging() gin.HandlerFunc {
	return LoggingWithConfig(DefaultLoggingConfig())
}

// LoggingWithConfig creates a logging middleware with custom configuration
func LoggingWithConfig(config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(c *gin.Context) {
		// Check if path should be skipped
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Start timing
		start := time.Now()

		// Capture request information
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		reqID := c.GetString("request_id")
		raw := c.Request.URL.RawQuery

		// Capture request body if configured
		var requestBody string
		if config.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if len(bodyBytes) <= config.MaxBodySize {
					requestBody = string(bodyBytes)
				} else {
					requestBody = string(bodyBytes[:config.MaxBodySize]) + "... (truncated)"
				}
				// Restore request body for further processing
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Wrap response writer to capture response
		var responseBody *bytes.Buffer
		if config.LogResponseBody {
			responseBody = &bytes.Buffer{}
		}

		wrapped := &responseWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = wrapped

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := wrapped.Status()

		// Build log event
		log := logger.FromContext(c.Request.Context())
		logEvent := log.With().
			Str("request_id", reqID).
			Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Int("status", status).
			Dur("latency", latency).
			Int64("latency_ms", latency.Milliseconds()).
			Logger()

		// Add optional fields
		if raw != "" {
			logEvent = logEvent.With().Str("query", raw).Logger()
		}

		if userAgent != "" {
			logEvent = logEvent.With().Str("user_agent", userAgent).Logger()
		}

		// Add request body if configured and available
		if requestBody != "" {
			logEvent = logEvent.With().Str("request_body", requestBody).Logger()
		}

		// Add response body if configured and available
		if responseBody != nil && responseBody.Len() > 0 {
			respBodyStr := responseBody.String()
			if len(respBodyStr) > config.MaxBodySize {
				respBodyStr = respBodyStr[:config.MaxBodySize] + "... (truncated)"
			}
			logEvent = logEvent.With().Str("response_body", respBodyStr).Logger()
		}

		// Add response size
		if wrapped.ResponseWriter.Size() > 0 {
			logEvent = logEvent.With().Int("response_size", wrapped.ResponseWriter.Size()).Logger()
		}

		// Add error information if present
		errors := c.Errors.ByType(gin.ErrorTypeAny)
		if len(errors) > 0 {
			errorMsgs := make([]string, len(errors))
			for i, err := range errors {
				errorMsgs[i] = err.Error()
			}
			logEvent = logEvent.With().Strs("errors", errorMsgs).Logger()
		}

		// Determine log level based on status code
		logLevel := getLogLevel(status, config.LogLevel)
		message := "HTTP request processed"

		// Add context-specific message
		if status >= 400 {
			message = "HTTP request failed"
		}

		// Log with appropriate level
		switch logLevel {
		case "error":
			logEvent.Error().Msg(message)
		case "warn":
			logEvent.Warn().Msg(message)
		case "debug":
			logEvent.Debug().Msg(message)
		default:
			logEvent.Info().Msg(message)
		}

		// Log slow requests separately
		if latency > 5*time.Second {
			log.Warn().
				Str("request_id", reqID).
				Str("method", method).
				Str("path", path).
				Dur("latency", latency).
				Msg("Slow request detected")
		}
	}
}

// StructuredLogging creates a more detailed structured logging middleware
func StructuredLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Build structured log data
		logData := map[string]interface{}{
			"timestamp":    start.Format(time.RFC3339),
			"request_id":   c.GetString("request_id"),
			"method":       c.Request.Method,
			"path":         path,
			"status":       c.Writer.Status(),
			"latency_ms":   time.Since(start).Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"referer":      c.Request.Referer(),
			"content_type": c.Request.Header.Get("Content-Type"),
		}

		// Add query parameters if present
		if raw != "" {
			logData["query"] = raw
		}

		// Add user information if available
		if authData, exists := c.Get("auth"); exists {
			if claims, ok := authData.(*struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
			}); ok {
				logData["user_id"] = claims.UserID
				logData["username"] = claims.Username
			}
		}

		// Add response headers
		responseHeaders := make(map[string]string)
		for key, values := range c.Writer.Header() {
			if len(values) > 0 {
				responseHeaders[key] = values[0]
			}
		}
		logData["response_headers"] = responseHeaders

		// Add error information
		if len(c.Errors) > 0 {
			errorList := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorList[i] = err.Error()
			}
			logData["errors"] = errorList
		}

		// Log with structured data
		log := logger.FromContext(c.Request.Context())
		status := c.Writer.Status()
		if status >= 500 {
			log.Error().
				Fields(logData).
				Msg("Server error occurred")
		} else if status >= 400 {
			log.Warn().
				Fields(logData).
				Msg("Client error occurred")
		} else {
			log.Info().
				Fields(logData).
				Msg("Request completed")
		}
	}
}

// DetailedLogging creates middleware with maximum logging detail
func DetailedLogging() gin.HandlerFunc {
	config := &LoggingConfig{
		SkipPaths:       []string{},
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     2048, // 2KB
		LogLevel: map[int]string{
			200: "debug",
			300: "info",
			400: "warn",
			500: "error",
		},
	}
	return LoggingWithConfig(config)
}

// MinimalLogging creates middleware with minimal logging for high-traffic endpoints
func MinimalLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log := logger.FromContext(c.Request.Context())
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("Request")
	}
}

// getLogLevel determines the appropriate log level based on status code
func getLogLevel(status int, levelMap map[int]string) string {
	// Check for exact match first
	if level, exists := levelMap[status]; exists {
		return level
	}

	// Check for range matches
	statusRange := (status / 100) * 100
	if level, exists := levelMap[statusRange]; exists {
		return level
	}

	// Default to info
	return "info"
}

// RequestSizeMiddleware logs requests that exceed a certain size
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			log := logger.FromContext(c.Request.Context())
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int64("content_length", c.Request.ContentLength).
				Int64("max_size", maxSize).
				Msg("Large request detected")
		}
		c.Next()
	}
}

// SecurityLogging logs security-related events
func SecurityLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Log authentication attempts
		if c.Request.URL.Path == "/api/v1/auth/login" {
			log.Info().
				Str("request_id", c.GetString("request_id")).
				Str("client_ip", c.ClientIP()).
				Str("user_agent", c.Request.UserAgent()).
				Msg("Authentication attempt")
		}

		c.Next()

		// Log failed authentication
		if c.Writer.Status() == 401 {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Str("user_agent", c.Request.UserAgent()).
				Msg("Unauthorized access attempt")
		}

		// Log access to sensitive endpoints
		if isSensitiveEndpoint(c.Request.URL.Path) {
			log.Info().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("client_ip", c.ClientIP()).
				Int("status", c.Writer.Status()).
				Msg("Sensitive endpoint access")
		}
	}
}

// isSensitiveEndpoint checks if the endpoint is considered sensitive
func isSensitiveEndpoint(path string) bool {
	sensitivePatterns := []string{
		"/admin",
		"/api/v1/users",
		"/api/v1/auth",
		"/api/v1/files/upload",
	}

	for _, pattern := range sensitivePatterns {
		if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
			return true
		}
	}
	return false
}

// PerformanceLogging logs performance metrics
func PerformanceLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		log := logger.FromContext(c.Request.Context())

		// Log slow requests
		if latency > time.Second {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Dur("latency", latency).
				Msg("Slow request detected")
		}

		// Log performance metrics every 100 requests (sampling)
		if helper.SafeIntFromInterface(c.GetString("request_id"), 0)%100 == 0 {
			log.Info().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Dur("latency", latency).
				Int("status", c.Writer.Status()).
				Msg("Performance sample")
		}
	}
}
