package controller

import (
	"cashier-api/service"

	"github.com/gofiber/fiber/v2"
)

type WarehouseControllerImpl struct {
	Service service.WarehouseService
}

func NewWarehouseControllerImpl(service service.WarehouseService) WarehouseController {
	return &WarehouseControllerImpl{Service: service}
}

func (controller *WarehouseControllerImpl) GetWarehouseItems(ctx *fiber.Ctx) error {

	return ctx.Send([]byte("sended"))
}
