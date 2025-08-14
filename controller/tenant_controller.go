package controller

import (
	"github.com/gofiber/fiber/v2"
)

/*
Every route here required authentication
*/
type TenantController interface {
	GetTenantWithUser(*fiber.Ctx) error
	NewTenant(*fiber.Ctx) error
	RemoveUserFromTenant(*fiber.Ctx) error
	AddUserToTenant(*fiber.Ctx) error
	GetTenantMembers(*fiber.Ctx) error
}
