package middleware

import (
	common "cashier-api/helper"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		// Next defines a function to skip this middleware when returned true.
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		Max:        20,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("x-forwarded-for")
		},
		LimitReached: func(ctx *fiber.Ctx) error {
			return ctx.Status(fiber.StatusNotFound).
				JSON(common.NewWebResponseError(fiber.StatusNotFound, common.StatusError, "Too many request. Please wait for 30 seconds."))
		},
	})
}
