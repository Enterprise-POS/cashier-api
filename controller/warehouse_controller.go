package controller

import "github.com/gofiber/fiber/v2"

/*
package controller

	handle input from user, to parsing correct parameter into user
	make router cleaner

	Warning! Do not handle business logic here
	only handle logic given by user input here
	- param
	- url
	- cookie
	- session
*/
type WarehouseController interface {
	/*
		Return all warehouse items by using current at given tenant id
	*/
	Get(ctx *fiber.Ctx) error
}
