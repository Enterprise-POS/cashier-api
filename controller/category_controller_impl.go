package controller

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"
	"strings"

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

	type RequestBody struct {
		Categories []string `json:"categories"`
	}
	var body RequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	categories, err := controller.Service.Create(tenantId, body.Categories)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"categories": categories,
		}))
}

// Get implements CategoryController.
func (controller *CategoryControllerImpl) Get(ctx *fiber.Ctx) error {
	paramLimit := ctx.Query("limit", "5")    // default 5
	paramPage := ctx.Query("page", "1")      // default 1
	nameQuery := ctx.Query("name_query", "") // default 1

	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	limit, err := strconv.Atoi(paramLimit)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	page, err := strconv.Atoi(paramPage)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	// Return []Category models
	categories, count, err := controller.Service.Get(tenantId, page, limit, nameQuery)
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST103)") {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "Requested range not satisfiable"))
		}
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"count":      count,
			"categories": categories,
		}))
}

// Register implements CategoryController.
func (controller *CategoryControllerImpl) Register(ctx *fiber.Ctx) error {
	type BodyCategories struct {
		CategoryId int `json:"category_id"`
		ItemId     int `json:"item_id"`
	}

	type RequestBody struct {
		TobeRegisters []BodyCategories `json:"tobe_registers"`
	}

	var body RequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	categoriesMtmWarehouse := make([]*model.CategoryMtmWarehouse, 0, len(body.TobeRegisters))
	for _, tobeRegister := range body.TobeRegisters {
		categoriesMtmWarehouse = append(categoriesMtmWarehouse, &model.CategoryMtmWarehouse{
			CategoryId: tobeRegister.CategoryId,
			ItemId:     tobeRegister.ItemId,
		})
	}

	err = controller.Service.Register(categoriesMtmWarehouse)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

// Unregister implements CategoryController.
func (controller *CategoryControllerImpl) Unregister(ctx *fiber.Ctx) error {
	var body struct {
		CategoryId int `json:"category_id"`
		ItemId     int `json:"item_id"`
	}
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	err = controller.Service.Unregister(&model.CategoryMtmWarehouse{CategoryId: body.CategoryId, ItemId: body.ItemId})
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// Update implements CategoryController.
func (controller *CategoryControllerImpl) Update(ctx *fiber.Ctx) error {
	var body struct {
		CategoryId   int    `json:"category_id"`
		CategoryName string `json:"category_name"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	updatedCategory, err := controller.Service.Update(tenantId, body.CategoryId, body.CategoryName)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"updated_category": updatedCategory,
		}))
}

// Delete implements CategoryController.
func (controller *CategoryControllerImpl) Delete(ctx *fiber.Ctx) error {
	var body struct {
		CategoryId int `json:"category_id"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	tobeDeletedCategory := &model.Category{
		Id:       body.CategoryId,
		TenantId: tenantId,
	}
	err = controller.Service.Delete(tobeDeletedCategory)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// GetItemsByCategoryId implements CategoryController.
func (controller *CategoryControllerImpl) GetItemsByCategoryId(ctx *fiber.Ctx) error {
	// It will be POST method so URL param will not use here
	var body struct {
		CategoryId int `json:"category_id"`
		Page       int `json:"page"`
		Limit      int `json:"limit"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	categoryWithItems, count, err := controller.Service.GetItemsByCategoryId(tenantId, body.CategoryId, body.Limit, body.Page)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by_tenant_id": tenantId,
			"category_id":            body.CategoryId,
			"count":                  count,
			"items":                  categoryWithItems,
		}))
}

// GetCategoryWithItems implements CategoryController.
func (controller *CategoryControllerImpl) GetCategoryWithItems(ctx *fiber.Ctx) error {
	// It will be POST method so URL param will not use here
	var body struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	categoryWithItems, count, err := controller.Service.GetCategoryWithItems(tenantId, body.Page, body.Limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by_tenant_id": tenantId,
			"count":                  count,
			"items":                  categoryWithItems,
		}))
}
