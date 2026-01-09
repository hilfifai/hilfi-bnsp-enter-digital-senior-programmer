package middleware

import (
	"api-jellyfish/internal/pkg/logger/v2"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseHeaderMiddleware adds standard response headers
func ResponseHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add standard response headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Add request ID to response header if available
		if reqID := c.GetString("request_id"); reqID != "" {
			c.Header("X-Request-ID", reqID)
		}

		// Add API version to response header if available
		if version := c.GetString("version"); version != "" {
			c.Header("X-API-Version", version)
		}

		c.Next()
	}
}

// ResponseTimeMiddleware adds response time header
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Calculate response time
		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())
	}
}

// CacheControlMiddleware adds cache control headers
func CacheControlMiddleware(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set cache control headers based on request method and path
		switch c.Request.Method {
		case "GET":
			if isStaticResource(c.Request.URL.Path) {
				c.Header("Cache-Control", "public, max-age=86400") // 1 day for static resources
			} else {
				c.Header("Cache-Control", "no-cache, must-revalidate")
			}
		case "POST", "PUT", "DELETE", "PATCH":
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// isStaticResource checks if the path is for static resources
func isStaticResource(path string) bool {
	staticExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg"}
	for _, ext := range staticExtensions {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// ETagMiddleware adds ETag support (basic implementation)
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only add ETag for GET requests with 200 status
		if c.Request.Method == "GET" && c.Writer.Status() == 200 {
			// Simple ETag based on response size and current time
			// In production, you might want to use content hash
			etag := `"` + c.GetString("request_id") + `"`
			c.Header("ETag", etag)

			// Check If-None-Match header
			if match := c.GetHeader("If-None-Match"); match == etag {
				c.AbortWithStatus(304) // Not Modified
				return
			}
		}
	}
}

// CompressionHeaderMiddleware adds headers to indicate compression support
func CompressionHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set Vary header to indicate that response varies based on Accept-Encoding
		c.Header("Vary", "Accept-Encoding")

		c.Next()

		// Log compression information if response was compressed
		if encoding := c.Writer.Header().Get("Content-Encoding"); encoding != "" {
			log := logger.FromContext(c.Request.Context())
			log.Debug().
				Str("request_id", c.GetString("request_id")).
				Str("path", c.Request.URL.Path).
				Str("encoding", encoding).
				Int("response_size", c.Writer.Size()).
				Msg("Response compressed")
		}
	}
}

// JSONResponseMiddleware ensures JSON content type for API responses
func JSONResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set JSON content type for API endpoints
		if isAPIEndpoint(c.Request.URL.Path) {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}

		c.Next()
	}
}

// isAPIEndpoint checks if the path is an API endpoint
func isAPIEndpoint(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

// DebugResponseMiddleware adds debug information in debug mode
func DebugResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() == gin.DebugMode {
			start := time.Now()

			c.Next()

			// Add debug headers in debug mode
			c.Header("X-Debug-Request-ID", c.GetString("request_id"))
			c.Header("X-Debug-Processing-Time", time.Since(start).String())
			c.Header("X-Debug-Go-Version", "go1.21")
			c.Header("X-Debug-Gin-Mode", gin.Mode())

			log := logger.FromContext(c.Request.Context())
			log.Debug().
				Str("request_id", c.GetString("request_id")).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", c.Writer.Status()).
				Dur("processing_time", time.Since(start)).
				Msg("Debug response headers added")
		} else {
			c.Next()
		}
	}
}
