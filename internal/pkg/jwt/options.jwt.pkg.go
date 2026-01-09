package jwt

import (
	"time"
)

const (
	DefaultTokenExpiredTime        = 3 * time.Hour
	DefaultRefreshTokenExpiredTime = 24 * time.Hour
	DefaultSigningMethod           = "HS256"
)

type SaveMethodJWTEnum string

const (
	REDIS SaveMethodJWTEnum = "REDIS"
	JWT   SaveMethodJWTEnum = "JWT"
)

type Options struct {
	TokenExpiredTime        time.Duration
	RefreshTokenExpiredTime time.Duration
	TokenSecretKey          string
	SigningMethod           string
	SaveMethod              SaveMethodJWTEnum
}

func DefaultOptions(secretKey string) *Options {
	return &Options{
		TokenExpiredTime:        DefaultTokenExpiredTime,
		RefreshTokenExpiredTime: DefaultRefreshTokenExpiredTime,
		TokenSecretKey:          secretKey,
		SigningMethod:           DefaultSigningMethod,
		SaveMethod:              JWT,
	}
}
