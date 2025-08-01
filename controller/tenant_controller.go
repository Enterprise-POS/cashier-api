package controller

import (
	"github.com/gofiber/fiber/v2"
)

type TenantController interface {
	GetTenantWithUser(ctx *fiber.Ctx) error
}
