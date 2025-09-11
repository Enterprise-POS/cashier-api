package controller

import (
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCategoryControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}
	testTimeout := int((time.Second * 5).Milliseconds())
	app := fiber.New()
	supabase := client.CreateSupabaseClient()
	categoryRepository := repository.NewCategoryRepositoryImpl(supabase)
	categoryService := service.NewCategoryServiceImpl(categoryRepository)
	categoryController := NewCategoryControllerImpl(categoryService)

	tenantRestriction := middleware.RestrictByTenant(supabase) // User only allowed to access associated tenant
	app.Post("/categories/create/:tenantId", tenantRestriction, categoryController.Create)

	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	testUser := &model.UserRegisterForm{
		Name:     "Test_CreateItem" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		Email:    uniqueIdentity + "@gmail.com",
		Password: "$2a$10$V6ZP0rm./adZ9kryl3mYf.MB9IY80Y8ZCjtKslUEPWoH.9PCsX7vK",
	}
	createdTestUser := createUser(supabase, testUser)
	testTenant := &model.Tenant{
		Name:        createdTestUser.Name + "'Group",
		OwnerUserId: createdTestUser.Id,
		IsActive:    true}
	createdTenant := createTenant(supabase, testTenant)

	byteBody, err := json.Marshal(fiber.Map{
		"email":    createdTestUser.Email,
		"password": "12345678",
	})
	require.NoError(t, err)

	body := strings.NewReader(string(byteBody))
	request := httptest.NewRequest("POST", "/users/sign_in", body)
	request.Header.Set("Content-Type", "application/json")

	// We need to get the cookie
	response, err := app.Test(request, testTimeout)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.StatusCode)
	var enterprisePOSCookie *http.Cookie
	for _, c := range response.Cookies() {
		if c.Name == constant.EnterprisePOS {
			enterprisePOSCookie = c
			break
		}
	}
	require.NotNil(t, enterprisePOSCookie)

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			requestBody := fiber.Map{
				"categories": []string{"Fast Food"},
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/create/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Nil(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.StatusCode)
		})
	})
}
