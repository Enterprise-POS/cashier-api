package controller

import "github.com/gofiber/fiber/v2"

type ExpenseController interface {
	// POST /expenses/:tenantId
	Create(ctx *fiber.Ctx) error

	// POST /expenses/search/:tenantId
	Get(ctx *fiber.Ctx) error

	// GET /expenses/details/:tenantId?expense_id=1
	FindById(ctx *fiber.Ctx) error

	// PUT /expenses/:tenantId
	Edit(ctx *fiber.Ctx) error

	// DELETE /expenses/:tenantId
	Delete(ctx *fiber.Ctx) error

	// GET /expense_categories/:tenantId
	GetCategories(ctx *fiber.Ctx) error

	// POST /expense_categories/:tenantId
	CreateCategory(ctx *fiber.Ctx) error

	// PUT /expense_categories/:tenantId
	EditCategory(ctx *fiber.Ctx) error

	// DELETE /expense_categories/:tenantId
	DeleteCategory(ctx *fiber.Ctx) error
}
