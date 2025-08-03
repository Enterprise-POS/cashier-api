package controller

import (
	"github.com/gofiber/fiber/v2"
)

type TenantController interface {
	GetTenantWithUser(*fiber.Ctx) error
	NewTenant(*fiber.Ctx) error
}
