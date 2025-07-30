package controller

import (
	"cashier-api/helper/client"
	constant "cashier-api/helper/constant/cookie"
	"cashier-api/repository"
	"cashier-api/service"
	"io"
	"net/http"
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
	app.Delete("users/sign_out", userController.SignOut)

	t.Run("SignUp", func(t *testing.T) {
		t.Run("NormalSignUp", func(t *testing.T) {
			body := strings.NewReader(`{
				"email": "testusercontroller_sign_up@gmail.com",
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
			assert.Contains(t, result, `"email":"testusercontroller_sign_up@gmail.com"`)

			// Check the cookie
			enterprisePosCookie, err := request.Cookie(constant.EnterprisePOS)
			assert.NotEqual(t, "", enterprisePosCookie)

			// clean up
			_, _, err = supabaseClient.From(repository.UserTable).
				Delete("", "").
				Eq("email", "testusercontroller_sign_up@gmail.com").
				Eq("name", "Test User").
				Execute()

			require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignUp_NormalSignUp")
		})

		t.Run("InvalidInput", func(t *testing.T) {
			// Invalid email
			body := strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_invalidinput.com",
				"password": "12345678",
				"name": "Test User"
			}`)
			request := httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 400, response.StatusCode)

			bytes, err := io.ReadAll(response.Body)
			assert.Nil(t, err)
			assert.NotNil(t, bytes)

			result := string(bytes)
			assert.Contains(t, result, `"status":"error"`)
			assert.Contains(t, result, `"message":"Could not create user account. Check input email"`)

			// Invalid password requirement
			body = strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_malformed@gmail.com",
				"password": "12345",
				"name": "Test User"
			}`)
			request = httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request)

			require.Nil(t, err)

			bytes, err = io.ReadAll(response.Body)
			require.Nil(t, err)

			result = string(bytes)
			assert.Contains(t, result, `"status":"error"`)
			assert.Contains(t, result, `"message":"Could not create user account. Check input password"`)

			// Invalid name requirement
			body = strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_malformed@gmail.com",
				"password": "12345678",
				"name": "T"
			}`)
			request = httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request)

			require.Nil(t, err)

			bytes, err = io.ReadAll(response.Body)
			require.Nil(t, err)

			result = string(bytes)
			assert.Contains(t, result, `"status":"error"`)
			assert.Contains(t, result, `"message":"Could not create user account. Check input name"`)
		})

		t.Run("InputMalformed", func(t *testing.T) {
			body := strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_malformed@gmail.com",
				"password": "12345678",
				"name": "Test User",,
			}`)
			request := httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 400, response.StatusCode)

			bytes, err := io.ReadAll(response.Body)
			assert.Nil(t, err)
			assert.NotNil(t, bytes)

			result := string(bytes)
			assert.Contains(t, result, `"status":"error"`)
			assert.Contains(t, result, `"message":"Something gone wrong ! The request body is malformed"`)
		})

		t.Run("LeaveBlankInput", func(t *testing.T) {
			body := strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_malformed@gmail.com",
				"password": "12345678",
				"name": ""
			}`)
			request := httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 400, response.StatusCode)

			bytes, err := io.ReadAll(response.Body)
			assert.Nil(t, err)
			assert.NotNil(t, bytes)

			result := string(bytes)
			assert.Contains(t, result, `"status":"error"`)
			assert.Contains(t, result, `"message":"Do not leave empty input ! Please check the inputted data"`)

			body = strings.NewReader(`{
				"email": "",
				"password": "12345678",
				"name": "Test User"
			}`)
			request = httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 400, response.StatusCode)

			body = strings.NewReader(`{
				"email": "testusercontroller_sign_up_input_malformed@gmail.com",
				"password": "",
				"name": "Test User"
			}`)
			request = httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request)

			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 400, response.StatusCode)
		})
	})

	t.Run("SignIn", func(t *testing.T) {
		// Create test user
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

		t.Run("NormalSignIn", func(t *testing.T) {
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
		})

		// clean up
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", "testusercontroller_sign_in@gmail.com").
			Eq("name", "Test User").
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignIn")
	})

	t.Run("SignOut", func(t *testing.T) {
		t.Run("NormalSignOut", func(t *testing.T) {
			body := strings.NewReader(`{
				"email": "testusercontroller_sign_out1@gmail.com",
				"password": "12345678",
				"name": "Test User"
			}`)
			request := httptest.NewRequest("POST", "/users/sign_up", body)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			require.Nil(t, err)
			require.NotNil(t, response)
			require.Equal(t, 201, response.StatusCode)

			// Logged in / Signed in cookie
			// Extract cookie from response
			var enterprisePOSCookie *http.Cookie
			for _, c := range response.Cookies() {
				if c.Name == constant.EnterprisePOS {
					enterprisePOSCookie = c
					break
				}
			}
			require.NotNil(t, enterprisePOSCookie)

			// Pass cookie from previous request
			request = httptest.NewRequest("DELETE", "/users/sign_out", nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request)

			assert.Nil(t, err)
			assert.Equal(t, 200, response.StatusCode)

			// Check if cookie is cleared
			cleared := false
			for _, c := range response.Cookies() {
				if c.Name == constant.EnterprisePOS {
					cleared = true
					assert.Equal(t, "", c.Value)
					assert.Equal(t, 0, c.MaxAge)
				}
			}
			assert.True(t, cleared)

			// Clean up
			_, _, err = supabaseClient.From(repository.UserTable).
				Delete("", "").
				Eq("email", "testusercontroller_sign_out1@gmail.com").
				Eq("name", "Test User").
				Execute()

			require.Nil(t, err, "If this failed, then delete data at DB. error at TestUserControllerImpl_SignOut1")
		})
	})
}
