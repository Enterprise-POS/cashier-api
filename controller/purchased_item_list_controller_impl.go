package controller

import (
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PurchasedItemControllerImpl struct {
	Service service.PurchasedItemService
}

func NewPurchasedItemControllerImpl(service service.PurchasedItemService) PurchasedItemController {
	return &PurchasedItemControllerImpl{Service: service}
}

type PurchasedItemListLogsRequest struct {
	ItemIds    []int               `json:"item_ids"`
	StoreId    int                 `json:"store_id"`
	Limit      int                 `json:"limit"`
	Page       int                 `json:"page"`
	DateFilter *query.DateFilter   `json:"date_filter"`
	Filters    []query.QueryFilter `json:"filters"`
}

type PurchasedItemListLogsResponse struct {
	Logs                []*model.PurchasedItem `json:"logs"`
	TotalCount          int                    `json:"total_count"`
	Page                int                    `json:"page"`
	Limit               int                    `json:"limit"`
	RequestedByTenantId int                    `json:"requested_by_tenant_id"`
}

// PurchasedItemListLogs implements PurchasedItemController.
func (controller *PurchasedItemControllerImpl) PurchasedItemListLogs(ctx *fiber.Ctx) error {
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	var body PurchasedItemListLogsRequest
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	logs, count, err := controller.Service.PurchasedItemListLogs(
		tenantId,
		body.StoreId,
		body.ItemIds,
		body.Limit,
		body.Page,
		body.DateFilter,
		body.Filters,
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, &PurchasedItemListLogsResponse{
			Logs:                logs,
			TotalCount:          count,
			Page:                body.Page,
			Limit:               body.Limit,
			RequestedByTenantId: tenantId,
		}))
}
