package controller

import (
	"github.com/gofiber/fiber/v2"
)

type CategoryController interface {
	/*
		Return an all items within category,
		items maybe double return, but different category id is required
	*/
	GetCategoryWithItems(*fiber.Ctx) error

	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategoryId(*fiber.Ctx) error

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

	/*
		Edit item category
		- Only 1 item allowed for now
	*/
	EditItemCategory(*fiber.Ctx) error

	/*
		Update existing category (Not updating category_mtm_warehouse table)
		- only category name allowed to edit
		- only update 1 category
	*/
	Update(*fiber.Ctx) error

	/*
		Deleting category; NOT category_mtm_warehouse
		If 1 category deleted, the other warehouse item that
		associate with that category also deleted at category_mtn_warehouse

		Only 1 category could be delete at the moment
	*/
	Delete(*fiber.Ctx) error
}
