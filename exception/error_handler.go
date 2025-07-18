package exception

import (
	common "cashier-api/helper"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	code := 500
	message := "internal server error"

	// if e, ok := err.(*NotFoundError); ok {
	// 	return ctx.Status(e.Code).JSON(web.WebResponse{
	// 		Code:   e.Code,
	// 		Status: e.Status,
	// 		Data: fiber.Map{
	// 			"message": e.Error(),
	// 		},
	// 	})
	// }

	// if e, ok := err.(*EmptyUidError); ok {
	// 	return ctx.Status(e.Code).JSON(web.WebResponse{
	// 		Code:   e.Code,
	// 		Status: e.Status,
	// 		Data: fiber.Map{
	// 			"message": e.Error(),
	// 		},
	// 	})
	// }

	// struct *fiber.Error is default error by go fiber
	code = err.(*fiber.Error).Code
	message = err.(*fiber.Error).Message
	log.Errorf("Unknown error occurred, code: %d", code)
	log.Errorf("Message: %s", message)

	errResponse := common.NewWebResponseError(code, common.StatusInternalServerError, message)
	return ctx.Status(code).JSON(errResponse)
}
