package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func RequestDebug() fiber.Handler {
	if os.Getenv("MODE") == "prod" {
		// Nothing to do
		return func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	} else {
		// Even in production mode, logrus already configured to not print
		// any debug code, but we want to save compute time here
		return func(ctx *fiber.Ctx) error {
			start := time.Now()

			// Generate or reuse request ID
			requestID := ctx.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			ctx.Set("X-Request-ID", requestID)

			method := ctx.Method()
			path := ctx.OriginalURL()
			ip := ctx.IP()
			log.Debugf("[REQ START] id=%s method=%s path=%s ip=%s", requestID, method, path, ip)

			// Execute everything downstream, then come back here
			err := ctx.Next()

			statusCode := ctx.Response().StatusCode()
			duration := time.Since(start)
			log.Debugf("[REQ END]   id=%s status=%d duration=%s", requestID, statusCode, duration)

			return err
		}
	}
}
