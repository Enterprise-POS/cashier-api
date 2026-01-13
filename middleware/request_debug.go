package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

		log.Printf(
			"[REQ START] id=%s method=%s path=%s ip=%s",
			requestID,
			c.Method(),
			c.OriginalURL(),
			c.IP(),
		)

		err := c.Next()

		log.Printf(
			"[REQ END]   id=%s status=%d duration=%s",
			requestID,
			c.Response().StatusCode(),
			time.Since(start),
		)

		return err
	}
}
