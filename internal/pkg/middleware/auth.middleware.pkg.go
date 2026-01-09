package middleware

import (
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/pkg/logger/v2"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets authentication context
func AuthMiddleware(auth jwt.IJWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Validate auth service
		if auth == nil {
			log.Error().Msg("Auth service is nil in AuthMiddleware")
			helper.APIInternalServerError(c, "Authentication service unavailable", nil)
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Debug().
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Msg("Authorization header missing")
			helper.APIUnauthorized(c, "Authorization header required", errors.New("missing authorization header"))
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Debug().
				Str("auth_header", authHeader).
				Str("path", c.Request.URL.Path).
				Msg("Invalid authorization header format")
			helper.APIUnauthorized(c, "Invalid authorization header format", errors.New("authorization header must be 'Bearer <token>'"))
			return
		}

		token := parts[1]
		if token == "" {
			log.Debug().
				Str("path", c.Request.URL.Path).
				Msg("Empty token in authorization header")
			helper.APIUnauthorized(c, "Empty token provided", errors.New("token cannot be empty"))
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			log.Debug().
				Err(err).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Msg("Token validation failed")
			helper.APIUnauthorized(c, "Invalid or expired token", err)
			return
		}

		// Validate claims
		if claims == nil {
			log.Error().
				Str("path", c.Request.URL.Path).
				Msg("Token validation returned nil claims")
			helper.APIUnauthorized(c, "Invalid token claims", errors.New("token claims are invalid"))
			return
		}

		// Log successful authentication
		log.Debug().
			Str("user_id", claims.UserID.String()).
			Str("username", claims.Username).
			Str("path", c.Request.URL.Path).
			Msg("Authentication successful")

		// Set authentication context
		c.Set("auth", claims)

		// Add user context to request context for logging
		ctx := logger.WithRequestID(c.Request.Context(), claims.UserID.String())
		c.Request = c.Request.WithContext(ctx)

		// Continue to next handler
		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
func OptionalAuthMiddleware(auth jwt.IJWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Validate auth service
		if auth == nil {
			log.Error().Msg("Auth service is nil in OptionalAuthMiddleware")
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// Invalid format, continue without authentication
			log.Debug().
				Str("auth_header", authHeader).
				Str("path", c.Request.URL.Path).
				Msg("Invalid authorization header format in optional auth")
			c.Next()
			return
		}

		token := parts[1]
		if token == "" {
			// Empty token, continue without authentication
			c.Next()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			log.Debug().
				Err(err).
				Str("path", c.Request.URL.Path).
				Msg("Optional token validation failed")
			c.Next()
			return
		}

		// Valid token, set authentication context
		if claims != nil {
			log.Debug().
				Str("user_id", claims.UserID.String()).
				Str("username", claims.Username).
				Str("path", c.Request.URL.Path).
				Msg("Optional authentication successful")

			c.Set("auth", claims)

			// Add user context to request context for logging
			ctx := logger.WithRequestID(c.Request.Context(), claims.UserID.String())
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// RequirePermissions middleware checks if authenticated user has required permissions
func RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Get authentication claims
		authData, exists := c.Get("auth")
		if !exists {
			helper.APIUnauthorized(c, "Authentication required", errors.New("no authentication data found"))
			return
		}

		claims, ok := authData.(*jwt.AuthClaims)
		if !ok {
			helper.APIInternalServerError(c, "Invalid authentication data", errors.New("cannot cast auth data to claims"))
			return
		}

		// Check if user has required permissions
		userPermissions := make(map[string]bool)
		for _, perm := range claims.Permissions {
			userPermissions[perm] = true
		}

		var missingPermissions []string
		for _, requiredPerm := range permissions {
			if !userPermissions[requiredPerm] {
				missingPermissions = append(missingPermissions, requiredPerm)
			}
		}

		if len(missingPermissions) > 0 {
			log.Debug().
				Str("user_id", claims.UserID.String()).
				Strs("missing_permissions", missingPermissions).
				Strs("required_permissions", permissions).
				Strs("user_permissions", claims.Permissions).
				Str("path", c.Request.URL.Path).
				Msg("Insufficient permissions")
			helper.APIForbidden(c, "Insufficient permissions", errors.New("missing required permissions"))
			return
		}

		log.Debug().
			Str("user_id", claims.UserID.String()).
			Strs("required_permissions", permissions).
			Str("path", c.Request.URL.Path).
			Msg("Permission check passed")

		c.Next()
	}
}

// RequireRoles middleware checks if authenticated user has required roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Get authentication claims
		authData, exists := c.Get("auth")
		if !exists {
			helper.APIUnauthorized(c, "Authentication required", errors.New("no authentication data found"))
			return
		}

		claims, ok := authData.(*jwt.AuthClaims)
		if !ok {
			helper.APIInternalServerError(c, "Invalid authentication data", errors.New("cannot cast auth data to claims"))
			return
		}

		// Check if user has required roles
		userRoles := make(map[string]bool)
		for _, role := range claims.RoleString {
			userRoles[role] = true
		}

		var missingRoles []string
		for _, requiredRole := range roles {
			if !userRoles[requiredRole] {
				missingRoles = append(missingRoles, requiredRole)
			}
		}

		if len(missingRoles) > 0 {
			log.Debug().
				Str("user_id", claims.UserID.String()).
				Strs("missing_roles", missingRoles).
				Strs("required_roles", roles).
				Strs("user_roles", claims.RoleString).
				Str("path", c.Request.URL.Path).
				Msg("Insufficient roles")
			helper.APIForbidden(c, "Insufficient roles", errors.New("missing required roles"))
			return
		}

		log.Debug().
			Str("user_id", claims.UserID.String()).
			Strs("required_roles", roles).
			Str("path", c.Request.URL.Path).
			Msg("Role check passed")

		c.Next()
	}
}
