package controller

import "github.com/gofiber/fiber/v2"

type OrderItemController interface {
	/*
		When cashier app press the button, then
		this will called
	*/
	PlaceOrderItem(ctx *fiber.Ctx) error

	/*
		Always minus page by 1 because PostgreSQL start index from 0
	*/
	FindById(ctx *fiber.Ctx) error

	/*
		Get the list of order_item, purchased_item_list will not included
		2nd params return is the count of all data
	*/
	Get(ctx *fiber.Ctx) error

	/*
		This method will insert into 2 table
	*/
	Transactions(ctx *fiber.Ctx) error

	/*
		Using aggregate function from SQL to get report
	*/
	GetSalesReport(ctx *fiber.Ctx) error
}
