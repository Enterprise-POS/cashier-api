package controller

import (
	common "cashier-api/helper"
	"cashier-api/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CategoryControllerImpl struct {
	Service service.CategoryService
}

func NewCategoryControllerImpl(service service.CategoryService) CategoryController {
	return &CategoryControllerImpl{Service: service}
}

// Create implements CategoryController.
func (controller *CategoryControllerImpl) Create(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	type BodyCategory struct {
		CategoryName string `json:"category_name"`
	}
	type RequestBody struct {
		Categories []*BodyCategory `json:"categories"`
	}
	var body RequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tobeCreatedCategories := make([]string, 0, len(body.Categories))
	for _, category := range body.Categories {
		tobeCreatedCategories = append(tobeCreatedCategories, category.CategoryName)
	}
	categories, err := controller.Service.Create(tenantId, tobeCreatedCategories)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"categories": categories,
		}))
}
