package logger

import (
	"context"
	"log/slog"
	"os"
)

type loggerKey struct{}

// New creates a new base slog JSON logger
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// WithCtx injects a logger into the context
func WithCtx(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromCtx extracts the logger from context. If none is found, it returns the default logger.
func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
