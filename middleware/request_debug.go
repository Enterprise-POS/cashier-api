package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func RequestDebug() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Generate or reuse request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)

		log.Debugf(
			"[REQ START] id=%s method=%s path=%s ip=%s",
			requestID,
			c.Method(),
			c.OriginalURL(),
			c.IP(),
		)

		err := c.Next()

		log.Debugf(
			"[REQ END]   id=%s status=%d duration=%s",
			requestID,
			c.Response().StatusCode(),
			time.Since(start),
		)

		return err
	}
}
