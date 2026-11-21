package controller

import "github.com/gofiber/fiber/v3"

type PurchasedItemController interface {
	PurchasedItemListLogs(ctx fiber.Ctx) error
}
