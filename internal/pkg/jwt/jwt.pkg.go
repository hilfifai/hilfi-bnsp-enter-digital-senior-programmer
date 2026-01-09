package jwt

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Auth struct
type Auth struct {
	TokenExpiredTime        time.Duration
	RefreshTokenExpiredTime time.Duration
	TokenSecretKey          string
	SigningMethod           string
	SaveMethod              SaveMethodJWTEnum
}

type AuthClaims struct {
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	Fullname    string      `json:"full_name"`
	Roles       []uuid.UUID `json:"role_ids"`
	RoleString  []string    `json:"roles"`
	Permissions []string    `json:"permissions"`
	UserID      uuid.UUID   `json:"user_id"`
	jwt.RegisteredClaims
}

type IJWTAuth interface {
	GenerateToken(data AuthClaims) (string, string, error)
	ValidateToken(jwtToken string) (*AuthClaims, error)
}

// New Auth object
func New(opt *Options) IJWTAuth {
	return &Auth{
		TokenExpiredTime:        opt.TokenExpiredTime,
		RefreshTokenExpiredTime: opt.RefreshTokenExpiredTime,
		TokenSecretKey:          opt.TokenSecretKey,
		SigningMethod:           opt.SigningMethod,
		SaveMethod:              opt.SaveMethod,
	}
}

// GenerateToken generate jwt token
func (a *Auth) GenerateToken(data AuthClaims) (string, string, error) {
	// Validate signing method
	signingMethod := jwt.GetSigningMethod(a.SigningMethod)
	if signingMethod == nil {
		return "", "", errors.New("invalid signing method: " + a.SigningMethod)
	}

	exp := time.Now().Add(a.TokenExpiredTime)

	// Fill in standard claims
	data.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	// Create token with signing method
	token := jwt.NewWithClaims(signingMethod, data)

	// Sign token
	accessToken, err := token.SignedString([]byte(a.TokenSecretKey))
	if err != nil {
		return "", "", err
	}

	// refresh token
	refresh_token_exp := time.Now().Add(a.RefreshTokenExpiredTime)
	data.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refresh_token_exp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	refreshToken := jwt.NewWithClaims(signingMethod, data)

	accessRefreshToken, err := refreshToken.SignedString([]byte(a.TokenSecretKey))
	if err != nil {
		return "", "", err
	}

	return accessToken, accessRefreshToken, nil
}

// ValidateToken validate jwt token
func (a *Auth) ValidateToken(jwtToken string) (*AuthClaims, error) {
	claims := &AuthClaims{}

	// Parse token with custom claims
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		// Prevent algorithm substitution attacks
		if token.Method.Alg() != a.SigningMethod {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.TokenSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Expiry check
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func GetUser(c *gin.Context) *AuthClaims {
	if c == nil {
		return nil
	}
	user := c.MustGet("auth").(*AuthClaims)
	return user
}
