package controller

import (
	common "cashier-api/helper"
	"cashier-api/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserControllerImpl struct {
	Service service.UserService
}

func NewUserControllerImpl(service service.UserService) UserController {
	return &UserControllerImpl{
		Service: service,
	}
}

func (controller *UserControllerImpl) SignUpWithEmailAndPassword(ctx *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	if body.Email == "" || body.Name == "" || body.Password == "" {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Do not leave empty input ! Please check the inputted data"))
	}

	// We success create user, but the user not signed in
	newCreatedUser, err := controller.Service.SignUpWithEmailAndPassword(body.Email, body.Password, body.Name)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	// We automatically signed in user
	newCreatedUser, tokenString, err := controller.Service.SignInWithEmailAndPassword(newCreatedUser.Email, body.Password)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}
	if tokenString == "" {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! Could not signed in user"))
	}

	// Apply to user cookie
	// Create cookie
	oneMonthFromNow := 24 * time.Hour * 30
	cookie := &fiber.Cookie{
		Name:     "_enterprise_pos",
		Value:    tokenString,
		Expires:  time.Now().Add(oneMonthFromNow), // This is cookie expiration time not jwt exp
		SameSite: "Lax",
		Secure:   true,
		HTTPOnly: true,
	}

	// Set cookie
	ctx.Cookie(cookie)

	return ctx.Status(fiber.StatusCreated).
		JSON(common.NewWebResponse(201, common.StatusSuccess, newCreatedUser))
}

func (controller *UserControllerImpl) SignInWithEmailAndPassword(ctx *fiber.Ctx) error {
	panic("not implemented")
}
