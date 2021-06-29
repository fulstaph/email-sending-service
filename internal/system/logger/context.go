package logger

import (
	"context"

	"go.uber.org/zap"
)

const loggerKey = "LOGGER"

func Enrich(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l) //nolint:staticcheck
}

func Fetch(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}

	return l
}
