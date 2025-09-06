package controller

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type WarehouseControllerImpl struct {
	Service service.WarehouseService
}

func NewWarehouseControllerImpl(service service.WarehouseService) WarehouseController {
	return &WarehouseControllerImpl{Service: service}
}

func (controller *WarehouseControllerImpl) Get(ctx *fiber.Ctx) error {
	paramTenantId := ctx.Params("tenantId")
	paramLimit := ctx.Query("limit", "5")        // default 5
	paramPage := ctx.Query("page", "1")          // default 1
	paramNameQuery := ctx.Query("nameQuery", "") // default empty

	tenantId, err := strconv.Atoi(paramTenantId)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

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

	/*
		Warning! Do not handle business logic here
		only handle logic given by user input
		- param
		- url
		- cookie
		- session

		'page' may be 1 but should be convert into 0
		in that case let service handle the logic
	*/
	result, count, err := controller.Service.GetWarehouseItems(tenantId, limit, page, paramNameQuery)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	successResponse := common.NewWebResponse(fiber.StatusOK, common.StatusSuccess, fiber.Map{
		"items": result,
		"count": count,
	})
	return ctx.Status(fiber.StatusOK).JSON(successResponse)
}

// GetActiveItem implements WarehouseController.
func (controller *WarehouseControllerImpl) GetActiveItem(ctx *fiber.Ctx) error {
	paramTenantId := ctx.Params("tenantId")
	paramLimit := ctx.Query("limit", "5")        // default 5
	paramPage := ctx.Query("page", "1")          // default 1
	paramNameQuery := ctx.Query("nameQuery", "") // default empty

	tenantId, err := strconv.Atoi(paramTenantId)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

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

	result, count, err := controller.Service.GetActiveItem(tenantId, limit, page, paramNameQuery)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	successResponse := common.NewWebResponse(fiber.StatusOK, common.StatusSuccess, fiber.Map{
		"items": result,
		"count": count,
	})
	return ctx.Status(fiber.StatusOK).JSON(successResponse)
}

// CreateItem implements WarehouseController.
func (controller *WarehouseControllerImpl) CreateItem(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	// Define item fields
	type BodyItems struct {
		ItemName string `json:"item_name"`
		Stocks   int    `json:"stocks"`
	}

	// Define full request body (embedding BodyItems)
	type CreateItemRequest struct {
		Items []*BodyItems `json:"items"`
	}

	var body CreateItemRequest
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	// Map to model
	tobeCreatedItems := make([]*model.Item, 0, len(body.Items))
	for _, item := range body.Items {
		tobeCreatedItems = append(tobeCreatedItems, &model.Item{
			ItemName: item.ItemName,
			Stocks:   item.Stocks,
			TenantId: tenantId,
			IsActive: true, // Always true because this is creating new item
		})
	}

	if len(tobeCreatedItems) == 0 {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Could not proceed, items body empty. At least 1 item required"))
	}

	createdItems, err := controller.Service.CreateItem(tobeCreatedItems)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"target_tenant":  tenantId,
			"new_item_count": len(createdItems),
			"items":          createdItems,
		}))
}

// FindById implements WarehouseController.
func (controller *WarehouseControllerImpl) FindById(ctx *fiber.Ctx) error {
	var body struct {
		ItemId int `json:"item_id"`
	}
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))
	item, err := controller.Service.FindById(body.ItemId, tenantId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_tenant": tenantId,
			"item":             item,
		}))
}

// Edit implements WarehouseController.
func (controller *WarehouseControllerImpl) Edit(ctx *fiber.Ctx) error {
	// TenantId & ItemId = Indicate which item will be edit
	// Currently allowed to update properties
	// - ItemName
	// - Quantity (via quantity parameter)
	type ReqItem struct {
		ItemName string `json:"item_name"`
		ItemId   int    `json:"item_id"`
	}
	var body struct {
		Quantity int     `json:"quantity"`
		Item     ReqItem `json:"item"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	tobeEditItem := &model.Item{
		ItemId:   body.Item.ItemId,
		TenantId: tenantId,
		ItemName: body.Item.ItemName,
	}
	err = controller.Service.Edit(body.Quantity, tobeEditItem)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

// SetActivate implements WarehouseController.
func (controller *WarehouseControllerImpl) SetActivate(ctx *fiber.Ctx) error {
	var body struct {
		ItemId  int  `json:"item_id"`
		SetInto bool `json:"set_into"`
	}
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))
	err = controller.Service.SetActivate(tenantId, body.ItemId, body.SetInto)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}
