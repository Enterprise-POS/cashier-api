package controller

import (
	"github.com/gofiber/fiber/v2"
)

type CategoryController interface {
	/*
		Create new category
	*/
	Create(*fiber.Ctx) error
}
