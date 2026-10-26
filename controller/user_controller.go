package controller

import "github.com/gofiber/fiber/v3"

type UserController interface {
	SignUpWithEmailAndPassword(ctx fiber.Ctx) error
	SignInWithEmailAndPassword(ctx fiber.Ctx) error
	SignOut(ctx fiber.Ctx) error
}
