package jwt

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecretKey = "test-secret-key-for-jwt-testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts *Options
		want *Auth
	}{
		{
			name: "should create auth with custom options",
			opts: &Options{
				TokenExpiredTime:        1 * time.Hour,
				RefreshTokenExpiredTime: 2 * time.Hour,
				TokenSecretKey:          testSecretKey,
				SigningMethod:           "HS256",
				SaveMethod:              JWT,
			},
			want: &Auth{
				TokenExpiredTime:        1 * time.Hour,
				RefreshTokenExpiredTime: 2 * time.Hour,
				TokenSecretKey:          testSecretKey,
				SigningMethod:           "HS256",
				SaveMethod:              JWT,
			},
		},
		{
			name: "should create auth with default options",
			opts: DefaultOptions(testSecretKey),
			want: &Auth{
				TokenExpiredTime:        DefaultTokenExpiredTime,
				RefreshTokenExpiredTime: DefaultRefreshTokenExpiredTime,
				TokenSecretKey:          testSecretKey,
				SigningMethod:           DefaultSigningMethod,
				SaveMethod:              JWT,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.opts)
			auth, ok := got.(*Auth)
			require.True(t, ok)
			assert.Equal(t, tt.want, auth)
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	secretKey := "test-secret"
	opts := DefaultOptions(secretKey)

	assert.Equal(t, DefaultTokenExpiredTime, opts.TokenExpiredTime)
	assert.Equal(t, DefaultRefreshTokenExpiredTime, opts.RefreshTokenExpiredTime)
	assert.Equal(t, secretKey, opts.TokenSecretKey)
	assert.Equal(t, DefaultSigningMethod, opts.SigningMethod)
	assert.Equal(t, JWT, opts.SaveMethod)
}

func TestAuth_GenerateToken(t *testing.T) {
	auth := New(DefaultOptions(testSecretKey))

	userID := uuid.New()
	roleID := uuid.New()
	claims := AuthClaims{
		Username:    "testuser",
		Email:       "test@example.com",
		Fullname:    "Test User",
		Roles:       []uuid.UUID{roleID},
		RoleString:  []string{"admin"},
		Permissions: []string{"read", "write"},
		UserID:      userID,
	}

	t.Run("should generate valid tokens", func(t *testing.T) {
		accessToken, refreshToken, err := auth.GenerateToken(claims)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotEqual(t, accessToken, refreshToken)

		// Verify access token can be parsed
		parsedClaims, err := auth.ValidateToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.Username, parsedClaims.Username)
		assert.Equal(t, claims.Email, parsedClaims.Email)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
	})

	t.Run("should handle invalid signing method", func(t *testing.T) {
		invalidAuth := &Auth{
			TokenExpiredTime:        1 * time.Hour,
			RefreshTokenExpiredTime: 2 * time.Hour,
			TokenSecretKey:          testSecretKey,
			SigningMethod:           "INVALID_METHOD",
			SaveMethod:              JWT,
		}

		_, _, err := invalidAuth.GenerateToken(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signing method")
	})
}

func TestAuth_ValidateToken(t *testing.T) {
	auth := New(DefaultOptions(testSecretKey))

	userID := uuid.New()
	roleID := uuid.New()
	claims := AuthClaims{
		Username:    "testuser",
		Email:       "test@example.com",
		Fullname:    "Test User",
		Roles:       []uuid.UUID{roleID},
		RoleString:  []string{"admin"},
		Permissions: []string{"read", "write"},
		UserID:      userID,
	}

	t.Run("should validate valid token", func(t *testing.T) {
		accessToken, _, err := auth.GenerateToken(claims)
		require.NoError(t, err)

		parsedClaims, err := auth.ValidateToken(accessToken)
		assert.NoError(t, err)
		assert.NotNil(t, parsedClaims)
		assert.Equal(t, claims.Username, parsedClaims.Username)
		assert.Equal(t, claims.Email, parsedClaims.Email)
		assert.Equal(t, claims.Fullname, parsedClaims.Fullname)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
		assert.Equal(t, claims.Roles, parsedClaims.Roles)
		assert.Equal(t, claims.RoleString, parsedClaims.RoleString)
		assert.Equal(t, claims.Permissions, parsedClaims.Permissions)
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"

		parsedClaims, err := auth.ValidateToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})

	t.Run("should reject empty token", func(t *testing.T) {
		parsedClaims, err := auth.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})

	t.Run("should reject token with wrong secret", func(t *testing.T) {
		// Generate token with different auth instance
		differentAuth := New(&Options{
			TokenExpiredTime:        1 * time.Hour,
			RefreshTokenExpiredTime: 2 * time.Hour,
			TokenSecretKey:          "different-secret",
			SigningMethod:           "HS256",
			SaveMethod:              JWT,
		})

		accessToken, _, err := differentAuth.GenerateToken(claims)
		require.NoError(t, err)

		// Try to validate with original auth (different secret)
		parsedClaims, err := auth.ValidateToken(accessToken)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})

	t.Run("should reject token with different signing method", func(t *testing.T) {
		// Create token manually with different signing method
		tokenClaims := claims
		tokenClaims.RegisteredClaims = jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}

		// Create token with HS512 instead of HS256
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, tokenClaims)
		tokenString, err := token.SignedString([]byte(testSecretKey))
		require.NoError(t, err)

		// Try to validate with HS256 auth
		parsedClaims, err := auth.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected signing method")
		assert.Nil(t, parsedClaims)
	})

	t.Run("should reject expired token", func(t *testing.T) {
		// Create auth with very short expiry
		shortAuth := New(&Options{
			TokenExpiredTime:        1 * time.Millisecond,
			RefreshTokenExpiredTime: 2 * time.Millisecond,
			TokenSecretKey:          testSecretKey,
			SigningMethod:           "HS256",
			SaveMethod:              JWT,
		})

		accessToken, _, err := shortAuth.GenerateToken(claims)
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		parsedClaims, err := shortAuth.ValidateToken(accessToken)
		assert.Error(t, err)
		assert.Nil(t, parsedClaims)
	})
}

func TestGetUser(t *testing.T) {
	t.Run("should return user from context", func(t *testing.T) {
		// Setup gin context
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(nil)

		userID := uuid.New()
		expectedUser := &AuthClaims{
			Username: "testuser",
			Email:    "test@example.com",
			UserID:   userID,
		}

		c.Set("auth", expectedUser)

		user := GetUser(c)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("should return nil for nil context", func(t *testing.T) {
		user := GetUser(nil)
		assert.Nil(t, user)
	})

	t.Run("should panic when auth not set in context", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(nil)

		assert.Panics(t, func() {
			GetUser(c)
		})
	})
}

func TestTokenExpiry(t *testing.T) {
	t.Run("should generate tokens with correct expiry times", func(t *testing.T) {
		tokenExpiry := 1 * time.Hour
		refreshTokenExpiry := 2 * time.Hour

		auth := New(&Options{
			TokenExpiredTime:        tokenExpiry,
			RefreshTokenExpiredTime: refreshTokenExpiry,
			TokenSecretKey:          testSecretKey,
			SigningMethod:           "HS256",
			SaveMethod:              JWT,
		})

		userID := uuid.New()
		claims := AuthClaims{
			Username: "testuser",
			UserID:   userID,
		}

		accessToken, refreshToken, err := auth.GenerateToken(claims)
		require.NoError(t, err)

		// Parse tokens to check expiry
		accessClaims, err := auth.ValidateToken(accessToken)
		require.NoError(t, err)

		refreshClaims, err := auth.ValidateToken(refreshToken)
		require.NoError(t, err)

		// Check that refresh token expires later than access token
		assert.True(t, refreshClaims.ExpiresAt.Time.After(accessClaims.ExpiresAt.Time))

		// Check approximate expiry times (allow some tolerance for execution time)
		now := time.Now()
		expectedAccessExpiry := now.Add(tokenExpiry)
		expectedRefreshExpiry := now.Add(refreshTokenExpiry)

		// Allow 1 second tolerance
		assert.WithinDuration(t, expectedAccessExpiry, accessClaims.ExpiresAt.Time, 1*time.Second)
		assert.WithinDuration(t, expectedRefreshExpiry, refreshClaims.ExpiresAt.Time, 1*time.Second)
	})
}

func TestSaveMethodJWTEnum(t *testing.T) {
	t.Run("should have correct enum values", func(t *testing.T) {
		assert.Equal(t, SaveMethodJWTEnum("REDIS"), REDIS)
		assert.Equal(t, SaveMethodJWTEnum("JWT"), JWT)
	})
}

func TestAuthClaims(t *testing.T) {
	t.Run("should create AuthClaims with all fields", func(t *testing.T) {
		userID := uuid.New()
		roleID := uuid.New()

		claims := AuthClaims{
			Username:    "testuser",
			Email:       "test@example.com",
			Fullname:    "Test User",
			Roles:       []uuid.UUID{roleID},
			RoleString:  []string{"admin", "user"},
			Permissions: []string{"read", "write", "delete"},
			UserID:      userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.Equal(t, "Test User", claims.Fullname)
		assert.Equal(t, []uuid.UUID{roleID}, claims.Roles)
		assert.Equal(t, []string{"admin", "user"}, claims.RoleString)
		assert.Equal(t, []string{"read", "write", "delete"}, claims.Permissions)
		assert.Equal(t, userID, claims.UserID)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
	})
}

func TestIntegration(t *testing.T) {
	t.Run("should complete full JWT lifecycle", func(t *testing.T) {
		// Initialize auth
		auth := New(DefaultOptions(testSecretKey))

		// Create user claims
		userID := uuid.New()
		roleID := uuid.New()
		originalClaims := AuthClaims{
			Username:    "integrationuser",
			Email:       "integration@example.com",
			Fullname:    "Integration Test User",
			Roles:       []uuid.UUID{roleID},
			RoleString:  []string{"admin"},
			Permissions: []string{"read", "write"},
			UserID:      userID,
		}

		// Generate tokens
		accessToken, refreshToken, err := auth.GenerateToken(originalClaims)
		require.NoError(t, err)
		require.NotEmpty(t, accessToken)
		require.NotEmpty(t, refreshToken)

		// Validate access token
		validatedClaims, err := auth.ValidateToken(accessToken)
		require.NoError(t, err)
		require.NotNil(t, validatedClaims)

		// Verify all claims are preserved
		assert.Equal(t, originalClaims.Username, validatedClaims.Username)
		assert.Equal(t, originalClaims.Email, validatedClaims.Email)
		assert.Equal(t, originalClaims.Fullname, validatedClaims.Fullname)
		assert.Equal(t, originalClaims.Roles, validatedClaims.Roles)
		assert.Equal(t, originalClaims.RoleString, validatedClaims.RoleString)
		assert.Equal(t, originalClaims.Permissions, validatedClaims.Permissions)
		assert.Equal(t, originalClaims.UserID, validatedClaims.UserID)

		// Validate refresh token
		refreshClaims, err := auth.ValidateToken(refreshToken)
		require.NoError(t, err)
		require.NotNil(t, refreshClaims)

		// Verify refresh token has same user data but different expiry
		assert.Equal(t, originalClaims.Username, refreshClaims.Username)
		assert.True(t, refreshClaims.ExpiresAt.Time.After(validatedClaims.ExpiresAt.Time))

		// Test with Gin context
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(nil)
		c.Set("auth", validatedClaims)

		contextUser := GetUser(c)
		require.NotNil(t, contextUser)
		assert.Equal(t, validatedClaims, contextUser)
	})
}

func TestAuth_GenerateToken_EdgeCases(t *testing.T) {
	t.Run("should handle empty secret key", func(t *testing.T) {
		auth := &Auth{
			TokenExpiredTime:        1 * time.Hour,
			RefreshTokenExpiredTime: 2 * time.Hour,
			TokenSecretKey:          "",
			SigningMethod:           "HS256",
			SaveMethod:              JWT,
		}

		userID := uuid.New()
		claims := AuthClaims{
			Username: "testuser",
			UserID:   userID,
		}

		accessToken, refreshToken, err := auth.GenerateToken(claims)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("should handle zero expiry time", func(t *testing.T) {
		auth := &Auth{
			TokenExpiredTime:        0,
			RefreshTokenExpiredTime: 0,
			TokenSecretKey:          testSecretKey,
			SigningMethod:           "HS256",
			SaveMethod:              JWT,
		}

		userID := uuid.New()
		claims := AuthClaims{
			Username: "testuser",
			UserID:   userID,
		}

		accessToken, refreshToken, err := auth.GenerateToken(claims)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("should handle all claims fields", func(t *testing.T) {
		auth := New(DefaultOptions(testSecretKey))

		userID := uuid.New()
		roleID1 := uuid.New()
		roleID2 := uuid.New()
		claims := AuthClaims{
			Username:    "fulluser",
			Email:       "full@example.com",
			Fullname:    "Full Test User",
			Roles:       []uuid.UUID{roleID1, roleID2},
			RoleString:  []string{"admin", "user", "moderator"},
			Permissions: []string{"read", "write", "delete", "admin"},
			UserID:      userID,
		}

		accessToken, refreshToken, err := auth.GenerateToken(claims)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// Validate all fields are preserved
		parsedClaims, err := auth.ValidateToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, claims.Username, parsedClaims.Username)
		assert.Equal(t, claims.Email, parsedClaims.Email)
		assert.Equal(t, claims.Fullname, parsedClaims.Fullname)
		assert.Equal(t, claims.Roles, parsedClaims.Roles)
		assert.Equal(t, claims.RoleString, parsedClaims.RoleString)
		assert.Equal(t, claims.Permissions, parsedClaims.Permissions)
		assert.Equal(t, claims.UserID, parsedClaims.UserID)
	})
}

func TestAuth_ValidateToken_EdgeCases(t *testing.T) {
	auth := New(DefaultOptions(testSecretKey))

	t.Run("should handle malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"not.a.token",
			"onlyonepart",
			"two.parts",
			"three.parts.but.malformed",
			"",
			".",
			"..",
			"...",
		}

		for _, token := range malformedTokens {
			parsedClaims, err := auth.ValidateToken(token)
			assert.Error(t, err)
			assert.Nil(t, parsedClaims)
		}
	})

	t.Run("should handle token with nil expiry", func(t *testing.T) {
		// Create a token manually without expiry
		claims := AuthClaims{
			Username: "testuser",
			UserID:   uuid.New(),
		}
		// Don't set RegisteredClaims so ExpiresAt will be nil

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecretKey))
		require.NoError(t, err)

		// Should still validate successfully (no expiry check needed)
		parsedClaims, err := auth.ValidateToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, parsedClaims)
		assert.Equal(t, claims.Username, parsedClaims.Username)
	})
}

func TestGetUser_EdgeCases(t *testing.T) {
	t.Run("should handle context with wrong auth type", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(nil)

		// Set auth as wrong type
		c.Set("auth", "not-auth-claims")

		assert.Panics(t, func() {
			GetUser(c)
		})
	})

	t.Run("should handle context with nil auth", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(nil)

		// Set auth as nil
		c.Set("auth", (*AuthClaims)(nil))

		user := GetUser(c)
		assert.Nil(t, user)
	})
}

func TestDefaultOptions_Validation(t *testing.T) {
	t.Run("should create options with provided secret", func(t *testing.T) {
		secretKey := "my-super-secret-key"
		opts := DefaultOptions(secretKey)

		assert.Equal(t, secretKey, opts.TokenSecretKey)
		assert.Equal(t, DefaultTokenExpiredTime, opts.TokenExpiredTime)
		assert.Equal(t, DefaultRefreshTokenExpiredTime, opts.RefreshTokenExpiredTime)
		assert.Equal(t, DefaultSigningMethod, opts.SigningMethod)
		assert.Equal(t, JWT, opts.SaveMethod)
	})

	t.Run("should handle empty secret key", func(t *testing.T) {
		opts := DefaultOptions("")
		assert.Equal(t, "", opts.TokenSecretKey)
	})
}

func TestSaveMethodJWTEnum_Values(t *testing.T) {
	t.Run("should have correct string values", func(t *testing.T) {
		assert.Equal(t, "REDIS", string(REDIS))
		assert.Equal(t, "JWT", string(JWT))
	})

	t.Run("should be usable in comparisons", func(t *testing.T) {
		var method SaveMethodJWTEnum = JWT
		assert.True(t, method == JWT)
		assert.False(t, method == REDIS)

		method = REDIS
		assert.True(t, method == REDIS)
		assert.False(t, method == JWT)
	})
}

// Benchmark tests
func BenchmarkAuth_GenerateToken(b *testing.B) {
	auth := New(DefaultOptions(testSecretKey))
	userID := uuid.New()
	roleID := uuid.New()
	claims := AuthClaims{
		Username:    "benchuser",
		Email:       "bench@example.com",
		Fullname:    "Benchmark User",
		Roles:       []uuid.UUID{roleID},
		RoleString:  []string{"admin"},
		Permissions: []string{"read", "write"},
		UserID:      userID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := auth.GenerateToken(claims)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAuth_ValidateToken(b *testing.B) {
	auth := New(DefaultOptions(testSecretKey))
	userID := uuid.New()
	claims := AuthClaims{
		Username: "benchuser",
		UserID:   userID,
	}

	// Generate token once
	token, _, err := auth.GenerateToken(claims)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := auth.ValidateToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetUser(b *testing.B) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	user := &AuthClaims{
		Username: "benchuser",
		UserID:   userID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := gin.CreateTestContext(nil)
		c.Set("auth", user)
		_ = GetUser(c)
	}
}

// Additional test cases for better coverage
func TestAuth_ValidateToken_SpecialCases(b *testing.T) {
	auth := New(DefaultOptions(testSecretKey))

	b.Run("should handle token with future issued time", func(b *testing.T) {
		userID := uuid.New()
		claims := AuthClaims{
			Username: "futureuser",
			UserID:   userID,
		}

		// Manually create claims with future issued time
		claims.RegisteredClaims = jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Future time
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecretKey))
		require.NoError(b, err)

		// Should still validate (JWT library handles this)
		parsedClaims, err := auth.ValidateToken(tokenString)
		assert.NoError(b, err)
		assert.NotNil(b, parsedClaims)
	})

	b.Run("should handle different signing methods", func(b *testing.T) {
		testCases := []struct {
			name   string
			method jwt.SigningMethod
		}{
			{"HS256", jwt.SigningMethodHS256},
			{"HS384", jwt.SigningMethodHS384},
			{"HS512", jwt.SigningMethodHS512},
		}

		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.T) {
				customAuth := &Auth{
					TokenExpiredTime:        1 * time.Hour,
					RefreshTokenExpiredTime: 2 * time.Hour,
					TokenSecretKey:          testSecretKey,
					SigningMethod:           tc.method.Alg(),
					SaveMethod:              JWT,
				}

				userID := uuid.New()
				claims := AuthClaims{
					Username: "testuser",
					UserID:   userID,
				}

				accessToken, refreshToken, err := customAuth.GenerateToken(claims)
				assert.NoError(b, err)
				assert.NotEmpty(b, accessToken)
				assert.NotEmpty(b, refreshToken)

				// Validate token
				parsedClaims, err := customAuth.ValidateToken(accessToken)
				assert.NoError(b, err)
				assert.Equal(b, claims.Username, parsedClaims.Username)
			})
		}
	})
}

func TestAuthClaims_JSONSerialization(b *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	original := AuthClaims{
		Username:    "jsonuser",
		Email:       "json@example.com",
		Fullname:    "JSON Test User",
		Roles:       []uuid.UUID{roleID},
		RoleString:  []string{"admin", "user"},
		Permissions: []string{"read", "write", "delete"},
		UserID:      userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Test JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(original)
	assert.NoError(b, err)
	assert.NotEmpty(b, jsonData)

	var unmarshalled AuthClaims
	err = json.Unmarshal(jsonData, &unmarshalled)
	assert.NoError(b, err)
	assert.Equal(b, original.Username, unmarshalled.Username)
	assert.Equal(b, original.Email, unmarshalled.Email)
	assert.Equal(b, original.Fullname, unmarshalled.Fullname)
	assert.Equal(b, original.Roles, unmarshalled.Roles)
	assert.Equal(b, original.RoleString, unmarshalled.RoleString)
	assert.Equal(b, original.Permissions, unmarshalled.Permissions)
	assert.Equal(b, original.UserID, unmarshalled.UserID)
}
