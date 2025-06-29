package exception

import (
	common "cashier-api/helper"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler_InternalServerError(t *testing.T) {
	app := fiber.New(fiber.Config{
		IdleTimeout:             time.Second * 5,
		ReadTimeout:             time.Second * 5,
		WriteTimeout:            time.Second * 5,
		Prefork:                 false,
		EnableTrustedProxyCheck: true,
		ErrorHandler:            ErrorHandler,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "Some internal error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var body common.WebResponse
	json.NewDecoder(resp.Body).Decode(&body)

	assert.Equal(t, 500, body.Code)
	assert.Equal(t, "Internal Server Error", body.Status)

	data := body.Data.(map[string]interface{})
	assert.Contains(t, data["message"], "Unhandled error occurred")
}
