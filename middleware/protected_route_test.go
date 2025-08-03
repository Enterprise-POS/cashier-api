package middleware

import (
	"cashier-api/controller"
	"cashier-api/helper/client"
	"cashier-api/repository"
	"cashier-api/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestProtectedRoute(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	supabaseClient := client.CreateSupabaseClient()

	// To create new user, User endpoint are needed
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := controller.NewUserControllerImpl(userService)

	app := fiber.New()
	app.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	app.Get("/test", ProtectedRoute, func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	t.Run("TellUserToSignInFirst", func(t *testing.T) {
		// signUpRequest := &common.NewRequest{
		// 	Url:       "/test",
		// 	Method:    "GET",
		// 	Body:      nil,
		// 	TimeoutMs: 0,
		// }
		request := httptest.NewRequest("GET", "/test", nil)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, int(time.Second*3))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})
}
