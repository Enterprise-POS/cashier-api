package controller

import (
	"cashier-api/exception"
	common "cashier-api/helper"
	"cashier-api/repository"
	"cashier-api/service"
	"fmt"
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

// Get implements OrderItemController.
func (controller *OrderItemControllerImpl) Get(ctx *fiber.Ctx) error {
	panic("unimplemented")
}

// PlaceOrderItem implements OrderItemController.
func (controller *OrderItemControllerImpl) PlaceOrderItem(ctx *fiber.Ctx) error {
	panic("unimplemented")
}

// Transactions implements OrderItemController.
func (controller *OrderItemControllerImpl) Transactions(ctx *fiber.Ctx) error {
	// Even PurchasedItemList need ID, but Go will create default id as 0 if we don't specify it
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
