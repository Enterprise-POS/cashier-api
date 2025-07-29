package controller

import "github.com/gofiber/fiber/v2"

type UserController interface {
	SignUpWithEmailAndPassword(ctx *fiber.Ctx) error
	SignInWithEmailAndPassword(ctx *fiber.Ctx) error
}
