package controller

import "github.com/gofiber/fiber/v2"

type StoreStockController interface {
	/*
		Will get all available stock at requested storeId
	*/
	Get(ctx *fiber.Ctx) error

	GetV2(ctx *fiber.Ctx) error

	/*
		Even it's said to edit, the method not allowed for editing stock data,
		Other metadata such as 'price' is allowed
	*/
	Edit(ctx *fiber.Ctx) error

	/*
		There is no create method for store_stock, so from warehouse transfer stock into store_stock
		warehouse quantity is always mandatory, could not transfer stock to store_stock if quantity insufficient
		will do the same for store_stock, could not transfer stock to warehouse if quantity insufficient
	*/
	TransferStockToWarehouse(ctx *fiber.Ctx) error
	TransferStockToStoreStock(ctx *fiber.Ctx) error
}
