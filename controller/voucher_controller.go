package controller

import "github.com/gofiber/fiber/v3"

type VoucherController interface {
	Create(fiber.Ctx) error
	Delete(fiber.Ctx) error
	Use(fiber.Ctx) error
}
