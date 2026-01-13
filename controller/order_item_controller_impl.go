package controller

import (
	"cashier-api/exception"
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"cashier-api/service"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type OrderItemControllerImpl struct {
	Service service.OrderItemService
}

func NewOrderItemControllerImpl(service service.OrderItemService) OrderItemController {
	return &OrderItemControllerImpl{Service: service}
}

type OrderItemControllerGetRequest struct {
	TenantId   int                  `json:"tenant_id" binding:"required,gt=0"`
	StoreId    int                  `json:"store_id" binding:"required,gt=0"`
	Limit      int                  `json:"limit" binding:"required,gte=1,lte=100"`
	Page       int                  `json:"page" binding:"required,gte=1"`
	Filters    []*query.QueryFilter `json:"filters"`
	DateFilter *query.DateFilter    `json:"date_filter"`
}
type OrderItemControllerGetResponse struct {
	OrderItems          []*model.OrderItem `json:"order_items"`
	TotalCount          int                `json:"total_count"`
	Page                int                `json:"page"`
	Limit               int                `json:"limit"`
	RequestedBy         int                `json:"requested_by"`
	RequestedByTenantId int                `json:"requested_by_tenant_id"`
}

// Get implements OrderItemController.
func (controller *OrderItemControllerImpl) Get(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	sub := ctx.Locals("sub")
	id, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(401, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	var body OrderItemControllerGetRequest
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	orderItems, count, err := controller.Service.Get(body.TenantId, body.StoreId, body.Limit, body.Page, body.Filters, body.DateFilter)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}
	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, &OrderItemControllerGetResponse{
			OrderItems:          orderItems,
			TotalCount:          count,
			Page:                body.Page,
			Limit:               body.Limit,
			RequestedBy:         id,
			RequestedByTenantId: tenantId,
		}))
}

// PlaceOrderItem implements OrderItemController.
func (controller *OrderItemControllerImpl) PlaceOrderItem(ctx *fiber.Ctx) error {
	panic("unimplemented")
}

// FindById implements OrderItemController.
func (controller *OrderItemControllerImpl) FindById(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	rawOrderItemId := ctx.Query("order_item_id", "")
	orderItemId, err := strconv.Atoi(rawOrderItemId)
	if err != nil {
		errorMsg := fmt.Sprintf("Error while get order_item_id, given error_item_id = %s", rawOrderItemId)
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, errorMsg))
	}

	orderItem, purchasedItemList, err := controller.Service.FindById(orderItemId, tenantId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_order_item_id": orderItemId,
			"order_item":              orderItem,
			"purchased_item_list":     purchasedItemList,
		}))
}

// Transactions implements OrderItemController.
func (controller *OrderItemControllerImpl) Transactions(ctx *fiber.Ctx) error {
	// Even PurchasedItem need ID, but Go will create default id as 0 if we don't specify it
	// Expected body
	/*
		{
				"purchased_price": 30_000,
				"total_quantity":  5,
				"total_amount":    27_700,
				"discount_amount": 1_300,
				"sub_total":       29_000, // 20_000 + 9_000

				"items": [
					{
						"item_id":         1,
						"quantity":        2,
						"purchased_price": 10_000,
						"discount_amount": 500,
						"total_amount":    19_000, // (10_000 * 2) - (500 * 2)
					},
				],

				"user_id":   USER_ID,
				"tenant_id": TENANT_ID,
				"store_id":  STORE_ID,
			}
	*/
	var body repository.CreateTransactionParams
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	transactionId, err := controller.Service.Transactions(&body)
	if err != nil {
		if pgErr, ok := err.(*exception.PostgreSQLException); ok {
			errMessage := pgErr.Message
			switch pgErr.Code {
			case "P0001":
				if strings.Contains(errMessage, "Security violation") {
					return ctx.Status(fiber.StatusForbidden).
						JSON(common.NewWebResponseError(403, common.StatusError, err.Error()))
				}

				return ctx.Status(fiber.StatusBadRequest).
					JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
			default:
				log.Error(fmt.Sprintf("[DB ERROR] message: %s, code: %s, hint: %s", pgErr.Message, pgErr.Code, pgErr.Hint))
				ctx.Status(fiber.StatusInternalServerError).
					JSON(common.NewWebResponseError(500, common.StatusError, fmt.Sprintf("[DB ERROR] message: %s, code: %s, hint: %s", pgErr.Message, pgErr.Code, pgErr.Hint)))
			}
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"transaction_id": transactionId,
		}))
}

// GetSalesReport implements OrderItemController.
func (controller *OrderItemControllerImpl) GetSalesReport(ctx *fiber.Ctx) error {
	// It's guaranteed to be not "", because restrict by tenant already did check first
	tenantId, _ := strconv.Atoi(ctx.Params("tenantId"))

	var body struct {
		StoreId    int               `json:"store_id"`
		DateFilter *query.DateFilter `json:"date_filter"`
	}
	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	salesReport, err := controller.Service.GetSalesReport(tenantId, body.StoreId, body.DateFilter)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, salesReport))
}
