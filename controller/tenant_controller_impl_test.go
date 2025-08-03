package controller

import (
	"bytes"
	common "cashier-api/helper"
	"cashier-api/helper/client"
	constant "cashier-api/helper/constant/cookie"
	"cashier-api/middleware"
	"cashier-api/model"
	"cashier-api/repository"
	"cashier-api/service"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantControllerImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	supabaseClient := client.CreateSupabaseClient()

	// To create new user, User endpoint are needed
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)

	tenantRepo := repository.NewTenantRepositoryImpl(supabaseClient)
	tenantService := service.NewTenantServiceImpl(tenantRepo)
	tenantController := NewTenantControllerImpl(tenantService)

	app := fiber.New()
	app.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	app.Get("/tenants/:userId", middleware.ProtectedRoute, tenantController.GetTenantWithUser)
	app.Post("/tenants/new", middleware.ProtectedRoute, tenantController.NewTenant)

	t.Run("NewTenant", func(t *testing.T) {
		// Create dummy user
		userIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
		userPassword := "12345678"
		dummyUser := &model.User{
			Email: "test_tenant_controller_" + userIdentity + "@gmail.com",
			Name:  "test tenant" + userIdentity,
		}
		body := strings.NewReader(fmt.Sprintf(`{
				"email": "%s",
				"password": "%s",
				"name": "%s"
			}`, dummyUser.Email, userPassword, dummyUser.Name))

		signUpRequest := &common.NewRequest{
			Url:       "/users/sign_up",
			Method:    "POST",
			Body:      body,
			TimeoutMs: 0,
		}
		_, response, err := signUpRequest.RunRequest(app)

		require.Nil(t, err)
		require.Equal(t, 201, response.StatusCode)

		var enterprisePOSCookie *http.Cookie
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
		dummyUser.Id = int(sub)

		t.Run("NormalNewTenant", func(t *testing.T) {
			dummyTenant := &model.Tenant{
				Name:        dummyUser.Name + " Group's",
				OwnerUserId: dummyUser.Id,
			}

			bodyBytes, err := json.Marshal(dummyTenant)
			require.NoError(t, err)

			request := httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))

			// The test itself
			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, "Created", responseBody)
			assert.Equal(t, http.StatusCreated, response.StatusCode)

			// Clean up
			// user_mtm_tenant
			_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
				Delete("", "").
				Eq("user_id", fmt.Sprint(dummyUser.Id)).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 1")

			// tenant
			_, _, err = supabaseClient.From(repository.TenantTable).
				Delete("", "").
				Eq("owner_user_id", fmt.Sprint(dummyTenant.OwnerUserId)).
				Eq("name", dummyTenant.Name).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 2")
		})

		t.Run("ForbiddenActionForCreateTenantWithOtherUserId", func(t *testing.T) {
			dummyTenant := &model.Tenant{
				Name:        dummyUser.Name + " Group's",
				OwnerUserId: 99,
			}

			bodyBytes, err := json.Marshal(dummyTenant)
			require.NoError(t, err)

			request := httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))

			// The test itself
			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			assert.Contains(t, responseBody, "Forbidden action detected ! Do not proceed")
		})

		// clean up
		// user
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", dummyUser.Email).
			Eq("name", dummyUser.Name).
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl")
	})

	t.Run("GetTenantWithUser", func(t *testing.T) {
		// Create dummy user
		userIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
		userPassword := "12345678"
		dummyUser := &model.User{
			Email: "test_tenant_controller_" + userIdentity + "@gmail.com",
			Name:  "test tenant" + userIdentity,
		}
		body := strings.NewReader(fmt.Sprintf(`{
				"email": "%s",
				"password": "%s",
				"name": "%s"
			}`, dummyUser.Email, userPassword, dummyUser.Name))

		signUpRequest := &common.NewRequest{
			Url:       "/users/sign_up",
			Method:    "POST",
			Body:      body,
			TimeoutMs: 0,
		}
		_, response, err := signUpRequest.RunRequest(app)

		require.Nil(t, err)
		require.Equal(t, 201, response.StatusCode)

		var enterprisePOSCookie *http.Cookie
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
		dummyUser.Id = int(sub)

		t.Run("NormalGet", func(t *testing.T) {
			// Create 1
			dummyTenant := &model.Tenant{
				Name:        dummyUser.Name + " Group's",
				OwnerUserId: dummyUser.Id,
			}

			bodyBytes, err := json.Marshal(dummyTenant)
			require.NoError(t, err)

			request := httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))

			// Create 2
			dummyTenant2 := &model.Tenant{
				Name:        dummyUser.Name + " Group's 2",
				OwnerUserId: dummyUser.Id,
			}

			bodyBytes, err = json.Marshal(dummyTenant2)
			require.NoError(t, err)

			request = httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))

			// The test itself
			request = httptest.NewRequest("GET", fmt.Sprintf("/tenants/%d", dummyUser.Id), nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, dummyTenant.Name)
			assert.Contains(t, responseBody, dummyTenant2.Name)

			// user_mtm_tenant
			_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
				Delete("", "").
				Eq("user_id", fmt.Sprint(dummyUser.Id)).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 1")

			// tenant
			_, _, err = supabaseClient.From(repository.TenantTable).
				Delete("", "").
				Eq("owner_user_id", fmt.Sprint(dummyTenant.OwnerUserId)).
				// Eq("name", dummyTenant.Name). by commenting this line, allow to supabase to delete 2 tenants
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 2")
		})

		// clean up
		// user
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", dummyUser.Email).
			Eq("name", dummyUser.Name).
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl")
	})
}
