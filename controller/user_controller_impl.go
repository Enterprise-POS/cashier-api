package controller

import (
	common "cashier-api/helper"
	constant "cashier-api/helper/constant/cookie"
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
		Name:     constant.EnterprisePOS,
		Value:    tokenString,
		Expires:  time.Now().Add(oneMonthFromNow), // This is cookie expiration time not jwt exp
		SameSite: "Lax",
		Secure:   true,
		HTTPOnly: true,
	}

	// Set cookie
	ctx.Cookie(cookie)

	return ctx.Status(fiber.StatusCreated).
		JSON(common.NewWebResponse(201, common.StatusSuccess, fiber.Map{
			"user":  newCreatedUser,
			"token": tokenString,
		}))
}

func (controller *UserControllerImpl) SignInWithEmailAndPassword(ctx *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	if body.Email == "" || body.Password == "" {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Do not leave empty input ! Please check the inputted data"))
	}

	// Only _enterprise_post is not defined could proceed
	userCookie := ctx.Cookies(constant.EnterprisePOS)
	if userCookie != "" {
		// TODO: For now this is do nothing. later maybe improve user cookie duration ?
		return ctx.Status(fiber.StatusOK).
			JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{"message": "Already created"}))
	}

	user, tokenString, err := controller.Service.SignInWithEmailAndPassword(body.Email, body.Password)
	if err != nil {
		if err.Error() == "No user with this credentials" {
			return ctx.Status(fiber.StatusUnauthorized).
				JSON(common.NewWebResponseError(401, common.StatusError, err.Error()))
		}
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
		Name:     constant.EnterprisePOS,
		Value:    tokenString,
		Expires:  time.Now().Add(oneMonthFromNow), // This is cookie expiration time not jwt exp
		SameSite: "Lax",
		Secure:   true,
		HTTPOnly: true,
	}

	// Set cookie
	ctx.Cookie(cookie)

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"user":  user,
			"token": tokenString,
		}))
}

// SignOut implements UserController.
func (controller *UserControllerImpl) SignOut(ctx *fiber.Ctx) error {
	enterprisePosCookie := ctx.Cookies(constant.EnterprisePOS, "")
	if enterprisePosCookie == "" {
		return ctx.Status(fiber.StatusOK).
			JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
				"message": "Required cookie not available / already signed out",
			}))
	} else {
		// Clear the credential jwt
		// ctx.ClearCookie(constant.EnterprisePOS)

		ctx.Cookie(&fiber.Cookie{
			Name:     constant.EnterprisePOS,
			Value:    "",
			Expires:  time.Unix(0, 0), // Unix epoch time
			MaxAge:   -1,
			HTTPOnly: true,
			Secure:   false, // Set to true if using HTTPS
			SameSite: "Lax",
			Path:     "/",
		})

		return ctx.Status(fiber.StatusOK).
			JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
				"message": "Signed out successfully",
			}))
	}
}
