package controller

import (
	common "cashier-api/helper"
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
	paramLimit := ctx.Query("limit", "5") // default 5
	paramPage := ctx.Query("page", "1")   // default 1

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
	result, count, err := controller.Service.GetWarehouseItems(tenantId, limit, page)
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
