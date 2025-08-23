package middleware

import (
	"cashier-api/controller"
	common "cashier-api/helper"
	"cashier-api/helper/client"
	constant "cashier-api/helper/constant/cookie"
	"cashier-api/repository"
	"cashier-api/service"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("JWTExpired", func(t *testing.T) {
		// Create user
		body := strings.NewReader(`{
				"email": "jwtexpireduser@gmail.com",
				"password": "12345678",
				"name": "JWTExpired User"
		}`)
		request := httptest.NewRequest("POST", "/users/sign_up", body)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, int(time.Second*5))
		require.NoError(t, err)
		require.NotNil(t, response)

		require.Nil(t, err)
		require.NotNil(t, response)
		require.Equal(t, 201, response.StatusCode)

		bytes, err := io.ReadAll(response.Body)
		require.Nil(t, err)
		require.NotNil(t, bytes)

		result := string(bytes)
		require.Contains(t, result, `"status":"success"`)

		var enterprisePOSCookie *http.Cookie = nil
		for _, c := range response.Cookies() {
			if c.Name == constant.EnterprisePOS {
				enterprisePOSCookie = c
			}
		}

		// Manually get the jwt payload from cookie
		claims := jwt.MapClaims{}
		token, err := common.ClaimJWT(enterprisePOSCookie.Value, &claims)
		require.NotEqual(t, "", claims["sub"])
		sub, ok := claims["sub"].(float64)
		require.True(t, ok)
		require.True(t, token.Valid)
		userId := sub

		// Generate an expired JWT
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":        userId,
			"email":      "test@gmail.com",
			"name":       "test name",
			"uuid":       "candidateUser UserUuid",
			"created_at": "candidateUser.CreatedAt",
			"exp":        time.Now().Add(-time.Minute).Unix(),
		})
		signedToken, err := expiredToken.SignedString([]byte(os.Getenv("JWT_S")))
		require.NoError(t, err)

		require.NotNil(t, enterprisePOSCookie)
		enterprisePOSCookie.Value = signedToken

		request = httptest.NewRequest("GET", "/test", nil)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, int(time.Second*5))
		require.NoError(t, err)
		require.NotNil(t, response)

		bytes, err = io.ReadAll(response.Body)
		require.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Contains(t, string(bytes), "token has invalid claims: token is expired")

		_, _, err = supabaseClient.From("user").
			Delete("", "").
			Eq("id", fmt.Sprint(userId)).
			Execute()
		require.Nil(t, err)
	})
}
