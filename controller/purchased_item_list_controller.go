package controller

import "github.com/gofiber/fiber/v2"

type PurchasedItemController interface {
	PurchasedItemListLogs(ctx *fiber.Ctx) error
}
