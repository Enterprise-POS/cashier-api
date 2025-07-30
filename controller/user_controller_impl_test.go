package controller

import (
	"cashier-api/helper/client"
	constant "cashier-api/helper/constant/cookie"
	"cashier-api/repository"
	"cashier-api/service"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserControllerImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	supabaseClient := client.CreateSupabaseClient()
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)
	app := fiber.New()
	app.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	t.Run("SignUp", func(t *testing.T) {
		t.Run("NormalSignUp", func(t *testing.T) {
			body := strings.NewReader(`{
				"email": "testwarehousecontroller_sign_sup@gmail.com",
				"password": "12345678",
				"name": "Test User"
			}`)
			request := httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 201, response.StatusCode)

			bytes, err := io.ReadAll(response.Body)
			assert.Nil(t, err)
			assert.NotNil(t, bytes)

			result := string(bytes)
			assert.Contains(t, result, `"status":"success"`)
			assert.Contains(t, result, `"name":"Test User"`)
			assert.Contains(t, result, `"email":"testwarehousecontroller_sign_up@gmail.com"`)

			// Check the cookie
			enterprisePosCookie, err := request.Cookie(constant.EnterprisePOS)
			assert.NotEqual(t, "", enterprisePosCookie)

			// clean up
			_, _, err = supabaseClient.From(repository.UserTable).
				Delete("", "").
				Eq("email", "testwarehousecontroller_sign_up@gmail.com").
				Eq("name", "Test User").
				Execute()

			require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignUp_NormalSignUp")
		})

		t.Run("InputMalformed", func(t *testing.T) {})
	})

	t.Run("SignIn", func(t *testing.T) {
		body := strings.NewReader(`{
			"email": "testusercontroller_sign_in@gmail.com",
			"password": "12345678",
			"name": "Test User"
		}`)

		request := httptest.NewRequest("POST", "/users/sign_up", body)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.Nil(t, err)
		require.Equal(t, 201, response.StatusCode)

		// Begin test itself
		body = strings.NewReader(`{
			"email": "testusercontroller_sign_in@gmail.com",
			"password": "12345678"
		}`)

		request = httptest.NewRequest("POST", "/users/sign_in", body)
		request.Header.Set("Content-Type", "application/json")
		response, err = app.Test(request)

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignIn")
		require.Equal(t, 200, response.StatusCode, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignIn")

		bytes, err := io.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.NotNil(t, bytes)

		result := string(bytes)
		assert.Contains(t, result, `"status":"success"`)
		assert.Contains(t, result, `"name":"Test User"`)
		assert.Contains(t, result, `"email":"testusercontroller_sign_in@gmail.com"`)

		// Check the cookie
		enterprisePosCookie, err := request.Cookie(constant.EnterprisePOS)
		assert.NotEqual(t, "", enterprisePosCookie)

		// clean up
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", "testusercontroller_sign_in@gmail.com").
			Eq("name", "Test User").
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignIn")
	})
}
