package controller

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StoreControllerImpl struct {
	Service service.StoreService
}

func NewStoreControllerImpl(service service.StoreService) StoreController {
	return &StoreControllerImpl{Service: service}
}

// Create implements StoreController.
func (controller *StoreControllerImpl) Create(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	type StoreCreateRequestBody struct {
		Name string `json:"name"`
	}
	var body StoreCreateRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	createdStore, err := controller.Service.Create(tenantId, body.Name)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"created_store": createdStore,
		}))
}

// GetAll implements StoreController.
func (controller *StoreControllerImpl) GetAll(ctx *fiber.Ctx) error {
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	paramLimit := ctx.Query("limit", "5") // default 5
	paramPage := ctx.Query("page", "1")   // default 1
	paramIncludeNonActive := ctx.Query("include_non_active", "false")

	limit, err := strconv.Atoi(paramLimit)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, "Please check limit URL parameter")
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	page, err := strconv.Atoi(paramPage)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, "Please check page URL parameter")
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	includeNonActive, err := strconv.ParseBool(paramIncludeNonActive)
	if err != nil {
		response := common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, "Please check include_non_active parameter")
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	}

	stores, count, err := controller.Service.GetAll(tenantId, page, limit, includeNonActive)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"page":                   page,
			"limit":                  limit,
			"count":                  count,
			"stores":                 stores,
			"requested_by_tenant_id": tenantId,
		}))
}

// SetActivate implements StoreController.
func (controller *StoreControllerImpl) SetActivate(ctx *fiber.Ctx) error {
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))
	type StoreSetActiveRequestBody struct {
		StoreId int  `json:"store_id"`
		SetInto bool `json:"set_into"`
	}
	var body StoreSetActiveRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(
				400, common.StatusError, "Something gone wrong ! The request body is malformed",
			))
	}

	err = controller.Service.SetActivate(tenantId, body.StoreId, body.SetInto)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

// Edit implements StoreController.
func (controller *StoreControllerImpl) Edit(ctx *fiber.Ctx) error {
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))
	type StoreSetActiveRequestBody struct {
		StoreId int    `json:"store_id"`
		Name    string `json:"name"`
	}
	var body StoreSetActiveRequestBody
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(
				400, common.StatusError, "Something gone wrong ! The request body is malformed",
			))
	}

	editedStore, err := controller.Service.Edit(
		&model.Store{TenantId: tenantId, Id: body.StoreId, Name: body.Name},
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"edited_store": editedStore,
		}))
}
