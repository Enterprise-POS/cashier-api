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
		Next: func(ctx *fiber.Ctx) bool {
			return ctx.IP() == "127.0.0.1"
		},
		Max:        50,
		Expiration: time.Second,
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return ctx.Get("x-forwarded-for")
		},
		LimitReached: func(ctx *fiber.Ctx) error {
			return ctx.Status(fiber.StatusTooManyRequests).
				JSON(common.NewWebResponseError(fiber.StatusTooManyRequests, common.StatusError, "Too many request. Please wait for 30 seconds."))
		},
	})
}
