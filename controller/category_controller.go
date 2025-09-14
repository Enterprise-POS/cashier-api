package controller

import "github.com/gofiber/fiber/v2"

type CategoryController interface {
	/*
		Create new category
	*/
	Create(*fiber.Ctx) error

	/*
		Get the category name only
	*/
	Get(*fiber.Ctx) error

	/*
		Register warehouse.item into category,
		this will inserting data into category_mtm_warehouse
	*/
	Register(*fiber.Ctx) error

	/*
		Unregister, deleting category_mtm_warehouse
		- Only 1 operation allowed for now
	*/
	Unregister(*fiber.Ctx) error
}
