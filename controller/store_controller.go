package controller

import "github.com/gofiber/fiber/v2"

type StoreController interface {
	/*
		Get All store, and filter available either active / non active only
	*/
	GetAll(ctx *fiber.Ctx) error

	/*
		Create new store
	*/
	Create(ctx *fiber.Ctx) error

	/*
		Either set to active / non-active
	*/
	SetActivate(ctx *fiber.Ctx) error
}
