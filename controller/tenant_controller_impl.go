package controller

import (
	common "cashier-api/helper"
	"cashier-api/service"

	"github.com/gofiber/fiber/v2"
)

type TenantControllerImpl struct {
	Service service.TenantService
}

func NewTenantControllerImpl(service service.TenantService) TenantController {
	return &TenantControllerImpl{
		Service: service,
	}
}

// GetTenantWithUser implements TenantController.
func (controller *TenantControllerImpl) GetTenantWithUser(ctx *fiber.Ctx) error {
	var body struct {
		UserId int `json:"user_id"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	sub := ctx.Locals("sub")
	id, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(401, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	data, err := controller.Service.GetTenantWithUser(body.UserId, id)
	if err != nil {
		if err.Error() == "[TenantService:GetTenantWithUser:1]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, " Forbidden action detected ! Do not proceed"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusCreated).
		JSON(common.NewWebResponse(201, common.StatusSuccess, fiber.Map{
			"requested_by": body.UserId,
			"tenants":      data,
		}))
}
