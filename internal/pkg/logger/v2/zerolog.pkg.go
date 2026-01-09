package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	userIDKey    ctxKey = "user_id"
)

type Config struct {
	Service string
	Env     string
	Version string
	Pretty  bool
	Level   int
}

func Init(cfg Config) {
	zerolog.SetGlobalLevel(zerolog.Level(cfg.Level))
	zerolog.TimeFieldFormat = time.RFC3339
	output := os.Stdout
	var l zerolog.Logger
	if cfg.Pretty {
		l = zerolog.New(zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}).
			With().
			Timestamp().
			Str("service", cfg.Service).
			Str("env", cfg.Env).
			Str("version", cfg.Version).
			Logger()
	} else {
		l = zerolog.New(output).
			With().
			Timestamp().
			Str("service", cfg.Service).
			Str("env", cfg.Env).
			Str("version", cfg.Version).
			Logger()
	}

	Log = l
}

// Inject ke context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// Ambil dari context
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// Logger dengan field standar
func FromContext(ctx context.Context) zerolog.Logger {
	l := Log
	if ctx == nil {
		return l
	}
	if reqID := GetRequestID(ctx); reqID != "" {
		l = l.With().Str("request_id", reqID).Logger()
	}
	if userID := GetUserID(ctx); userID != "" {
		l = l.With().Str("user_id", userID).Logger()
	}
	return l
}
