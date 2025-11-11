package controller

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StoreStockControllerImpl struct {
	Service service.StoreStockService
}

func NewStoreStockControllerImpl(service service.StoreStockService) StoreStockController {
	return &StoreStockControllerImpl{Service: service}
}

// Get implements StoreStockController.
func (controller *StoreStockControllerImpl) Get(ctx *fiber.Ctx) error {
	paramLimit := ctx.Query("limit", "5") // default 5
	paramPage := ctx.Query("page", "1")   // default 1
	paramStoreId := ctx.Query("store_id", "must specify")
	// nameQuery := ctx.Query("name_query", "")

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

	storeId, err := strconv.Atoi(paramStoreId)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	storeStocks, count, err := controller.Service.Get(tenantId, storeId, limit, page)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"count":        count,
			"store_stocks": storeStocks,
		}))
}

// GetV2 implements StoreStockController.
func (controller *StoreStockControllerImpl) GetV2(ctx *fiber.Ctx) error {
	paramLimit := ctx.Query("limit", "5") // default 5
	paramPage := ctx.Query("page", "1")   // default 1
	paramStoreId := ctx.Query("store_id", "must specify")
	nameQuery := ctx.Query("name_query", "")

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

	storeId, err := strconv.Atoi(paramStoreId)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, "Please check store id. Store id is required")
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	storeStocks, count, err := controller.Service.GetV2(tenantId, storeId, limit, page, nameQuery)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"count":        count,
			"store_stocks": storeStocks,
		}))
}

// Edit implements StoreStockController.
func (controller *StoreStockControllerImpl) Edit(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant
	// already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))
	type StoreStockEditRequestBody struct {
		Id     int `json:"id,omitempty"`
		Price  int `json:"price"`
		ItemId int `json:"item_id"`
		// TenantId int `json:"tenant_id"`
		StoreId int `json:"store_id"`
	}
	var body StoreStockEditRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	editedItem := &model.StoreStock{
		Id:       body.Id,
		Price:    body.Price,
		ItemId:   body.ItemId,
		TenantId: tenantId,
		StoreId:  body.StoreId,
	}
	err = controller.Service.Edit(editedItem)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

// TransferStockToStoreStock implements StoreStockController.
func (controller *StoreStockControllerImpl) TransferStockToStoreStock(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant
	// already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	type StoreStockTransferStockToStoreStockRequestBody struct {
		Quantity int `json:"quantity"`
		ItemId   int `json:"item_id"`
		StoreId  int `json:"store_id"`
		// TenantId int `json:"tenantId"` // Handled at url
	}
	var body StoreStockTransferStockToStoreStockRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	err = controller.Service.TransferStockToStoreStock(body.Quantity, body.ItemId, body.StoreId, tenantId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

// TransferStockToWarehouse implements StoreStockController.
func (controller *StoreStockControllerImpl) TransferStockToWarehouse(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant
	// already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	type StoreStockTransferStockToWarehouseRequestBody struct {
		Quantity int `json:"quantity"`
		ItemId   int `json:"item_id"`
		StoreId  int `json:"store_id"`
		// TenantId int `json:"tenantId"` // Handled at url
	}
	var body StoreStockTransferStockToWarehouseRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	err = controller.Service.TransferStockToWarehouse(body.Quantity, body.ItemId, body.StoreId, tenantId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}
