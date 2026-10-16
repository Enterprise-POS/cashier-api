package middleware

import (
	common "cashier-api/helper"
	constant "cashier-api/helper/constant/cookie"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

func ProtectedRoute(ctx *fiber.Ctx) error {
	var tokenString string

	// Check cookie first, then fall back to Authorization header
	if cookie := ctx.Cookies(constant.EnterprisePOS); cookie != "" {
		tokenString = cookie
	} else {
		authHeader := ctx.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Infof("[AUTH] no token provided: %s %s", ctx.Method(), ctx.Path())
			return unauthorized(ctx, "Sign in to access this route")
		}
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			log.Warnf("[AUTH] empty bearer token: %s %s", ctx.Method(), ctx.Path())
			return unauthorized(ctx, "Sign in to access this route")
		}
	}

	// Parse and validate the token
	claims := jwt.MapClaims{}
	token, err := common.ClaimJWT(tokenString, &claims)
	if err != nil {
		log.Warnf("[AUTH] invalid JWT: %s %s, reason: %v", ctx.Method(), ctx.Path(), err)
		return unauthorized(ctx, fmt.Sprintf("JWT malformed, Please try sign in again. reason: %s", err.Error()))
	}
	if !token.Valid {
		log.Warnf("[AUTH] JWT marked invalid: %s %s", ctx.Method(), ctx.Path())
		return unauthorized(ctx, "JWT malformed, Please try sign in again.")
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		log.Errorf("[AUTH] bad 'sub' claim: %s %s", ctx.Method(), ctx.Path())
		return unauthorized(ctx, "Unexpected behavior ! JWT body contain invalid value")
	}
	if _, ok := claims["exp"].(float64); !ok {
		log.Errorf("[AUTH] bad 'exp' claim: %s %s", ctx.Method(), ctx.Path())
		return unauthorized(ctx, "Unexpected behavior ! JWT body contain invalid value (2)")
	}

	ctx.Locals("sub", int(sub))
	log.Debugf("[AUTH] ok: user_id=%d %s %s", int(sub), ctx.Method(), ctx.Path())

	return ctx.Next()
}

func unauthorized(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusUnauthorized).
		JSON(common.NewWebResponseError(401, common.StatusError, message))
}
