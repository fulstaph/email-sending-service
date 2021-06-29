package logger

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var muted = map[string]struct{}{
	"":              {},
	"/":             {},
	"/health":       {},
	"/healthz":      {},
	"/health-check": {},
}

func WithLogger(log *zap.Logger) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		uri := string(ctx.Context().RequestURI())
		if _, ok := muted[uri]; !ok && !strings.HasPrefix(uri, "/swagger/") {
			log.Debug(ctx.String(),
				zap.String("request_id", ctx.Locals("requestid").(string)),
				zap.String("method", ctx.Method()),
				zap.String("url", uri),
				zap.String("body", string(ctx.Body())),
			)
		}
		start := time.Now()
		err := ctx.Next()

		status := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
		}

		if _, ok := muted[uri]; !ok && !strings.HasPrefix(uri, "/swagger/") {
			log.Debug(ctx.String(),
				zap.String("request_id", ctx.Locals("requestid").(string)),
				zap.String("duration", time.Since(start).String()),
				zap.Int("status", status),
				zap.String("body", string(ctx.Response().Body())),
			)
		}
		return err
	}
}
