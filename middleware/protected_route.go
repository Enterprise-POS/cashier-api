package middleware

import (
	common "cashier-api/helper"
	constant "cashier-api/helper/constant/cookie"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

func ProtectedRoute(ctx *fiber.Ctx) error {
	// Check user cookie, if user cookie not available immediately return unauthorized
	enterprisePOSCookie := ctx.Cookies(constant.EnterprisePOS)
	if enterprisePOSCookie == "" {
		return ctx.Status(fiber.StatusUnauthorized).
			JSON(common.NewWebResponseError(401, common.StatusError, "Sign in to access this route"))
	}

	// Check user jwt token validity
	claims := jwt.MapClaims{}

	// Parse and validate the token
	token, err := common.ClaimJWT(enterprisePOSCookie, &claims)

	if err != nil {
		log.Warnf("Malformed JWT detected, reason: %s", err.Error())
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "JWT malformed, try sign in again"))
	}
	if !token.Valid {
		log.Warn("Invalid token detected")
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "JWT malformed, try sign in again"))
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	// Store valuable data to send to next handler
	ctx.Locals("sub", int(sub))

	log.Debugf("Accessing protected route from sub/id: %d", int(sub))

	// If ok then go to next handler
	return ctx.Next()
}
