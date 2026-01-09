package middleware

import (
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/logger/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Authorization",
			"Cache-Control",
			"Connection",
			"DNT",
			"Host",
			"If-Modified-Since",
			"Keep-Alive",
			"User-Agent",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// ProductionCORSConfig returns a production-safe CORS configuration
func ProductionCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{
			"https://yourdomain.com",
			"https://www.yourdomain.com",
			"https://app.yourdomain.com",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"Authorization",
			"Cache-Control",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}
}

// CorsMiddleware creates a CORS middleware with default configuration
func CorsMiddleware() gin.HandlerFunc {
	return CorsMiddlewareWithConfig(DefaultCORSConfig())
}

// CorsMiddlewareWithConfig creates a CORS middleware with custom configuration
func CorsMiddlewareWithConfig(corsConfig *CORSConfig) gin.HandlerFunc {
	if corsConfig == nil {
		logger.Log.Warn().Msg("CORS config is nil, using default config")
		corsConfig = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Single logger instance for this middleware execution
		log := logger.FromContext(c.Request.Context())

		// Set Access-Control-Allow-Origin
		if len(corsConfig.AllowOrigins) == 1 && corsConfig.AllowOrigins[0] == "*" {
			// Allow all origins
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && isOriginAllowed(origin, corsConfig.AllowOrigins) {
			// Allow specific origins
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Set Access-Control-Allow-Methods
		if len(corsConfig.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowMethods, ", "))
		}

		// Set Access-Control-Allow-Headers
		if len(corsConfig.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowHeaders, ", "))
		}

		// Set Access-Control-Expose-Headers
		if len(corsConfig.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposeHeaders, ", "))
		}

		// Set Access-Control-Allow-Credentials
		if corsConfig.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set Access-Control-Max-Age
		if corsConfig.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", helper.SafeStringFromInterface(int(corsConfig.MaxAge.Seconds())))
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == http.MethodOptions {
			log.Debug().
				Str("origin", origin).
				Str("method", c.Request.Header.Get("Access-Control-Request-Method")).
				Str("headers", c.Request.Header.Get("Access-Control-Request-Headers")).
				Msg("CORS preflight request")

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Log CORS request for debugging
		if gin.Mode() == gin.DebugMode && origin != "" {
			log.Debug().
				Str("origin", origin).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Msg("CORS request processed")
		}

		c.Next()
	}
}

// SmartCorsMiddleware creates a CORS middleware that adapts based on environment
func SmartCorsMiddleware() gin.HandlerFunc {
	var corsConfig *CORSConfig

	switch gin.Mode() {
	case gin.ReleaseMode:
		corsConfig = ProductionCORSConfig()
		// Note: This is initialization time, no request context available yet
		logger.Log.Info().Msg("Using production CORS configuration")
	case gin.TestMode:
		corsConfig = &CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: false,
			MaxAge:           time.Hour,
		}
		// Note: This is initialization time, no request context available yet
		logger.Log.Info().Msg("Using test CORS configuration")
	default: // DebugMode
		corsConfig = DefaultCORSConfig()
		// Note: This is initialization time, no request context available yet
		logger.Log.Info().Msg("Using development CORS configuration")
	}

	return CorsMiddlewareWithConfig(corsConfig)
}

// ConfigurableCorsMiddleware creates CORS middleware from config
func ConfigurableCorsMiddleware() gin.HandlerFunc {
	corsConfig := &CORSConfig{
		AllowOrigins:     getConfigStringSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
		AllowMethods:     getConfigStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowHeaders:     getConfigStringSlice("CORS_ALLOW_HEADERS", DefaultCORSConfig().AllowHeaders),
		ExposeHeaders:    getConfigStringSlice("CORS_EXPOSE_HEADERS", []string{"Content-Length", "X-Request-ID"}),
		AllowCredentials: getConfigBool("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           getConfigDuration("CORS_MAX_AGE", 12*time.Hour),
	}

	// Note: This is initialization time, no request context available yet
	logger.Log.Info().
		Strs("allow_origins", corsConfig.AllowOrigins).
		Strs("allow_methods", corsConfig.AllowMethods).
		Bool("allow_credentials", corsConfig.AllowCredentials).
		Dur("max_age", corsConfig.MaxAge).
		Msg("CORS middleware configured from environment")

	return CorsMiddlewareWithConfig(corsConfig)
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}

		// Support for wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}
	return false
}

// Helper functions to get configuration values
func getConfigStringSlice(key string, defaultValue []string) []string {
	// This would typically read from config.Config or environment variables
	// For now, return default values
	return defaultValue
}

func getConfigBool(key string, defaultValue bool) bool {
	// This would typically read from config.Config or environment variables
	// For now, return default value
	return defaultValue
}

func getConfigDuration(key string, defaultValue time.Duration) time.Duration {
	// This would typically read from config.Config or environment variables
	// For now, return default value
	return defaultValue
}

// SecureHeadersMiddleware adds security headers alongside CORS
func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only add HSTS in production with HTTPS
		if gin.Mode() == gin.ReleaseMode && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy (adjust based on your needs)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		}

		c.Next()
	}
}
