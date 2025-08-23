package middleware

import (
	common "cashier-api/helper"
	constant "cashier-api/helper/constant/cookie"
	"fmt"

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
			JSON(common.NewWebResponseError(400, common.StatusError, fmt.Sprintf("JWT malformed, Please try sign in again. reason: %s", err.Error())))
	}
	if !token.Valid {
		log.Warn("Invalid token detected")
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "JWT malformed, Please try sign in again."))
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! JWT body contain invalid value"))
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! JWT body contain invalid value (2)"))
	}

	// Store valuable data to send to next handler
	ctx.Locals("sub", int(sub))

	log.Debugf("Accessing protected route from sub/id: %d", int(sub))
	log.Debugf("Current user will logged in until: %f", exp)

	// If ok then go to next handler
	return ctx.Next()
}
