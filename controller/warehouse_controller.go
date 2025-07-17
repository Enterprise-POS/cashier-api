package controller

import "github.com/gofiber/fiber/v2"

type WarehouseController interface {
	/*
		Return all warehouse items by using current at given tenant id
	*/
	GetWarehouseItems(ctx *fiber.Ctx) error
}
