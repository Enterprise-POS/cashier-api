package controller

import (
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestWarehouseControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	/*
		In current test file / warehouse test file we skip the
		user registration and tenant creation.

		but we still make use login
	*/
	app := fiber.New()
	supabaseClient := client.CreateSupabaseClient()
	warehouseRepo := repository.NewWarehouseRepositoryImpl(supabaseClient)
	warehouseService := service.NewWarehouseServiceImpl(warehouseRepo)
	warehouseController := NewWarehouseControllerImpl(warehouseService)

	// user
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)

	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	// warehouse
	app.Use(middleware.ProtectedRoute)                               // Must login
	tenantRestriction := middleware.RestrictByTenant(supabaseClient) // User only allowed to access associated tenant
	app.Get("/warehouses/:tenantId", tenantRestriction, warehouseController.Get)
	app.Post("/warehouse/create_item/:tenantId", tenantRestriction, warehouseController.CreateItem)

	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	testUser := &model.UserRegisterForm{
		Name:     "Test_CreateItem" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		Email:    uniqueIdentity + "@gmail.com",
		Password: "$2a$10$V6ZP0rm./adZ9kryl3mYf.MB9IY80Y8ZCjtKslUEPWoH.9PCsX7vK",
	}
	createdTestUser := createUser(supabaseClient, testUser)
	testTenant := &model.Tenant{
		Name:        createdTestUser.Name + "'Group",
		OwnerUserId: createdTestUser.Id,
		IsActive:    true}
	createdTenant := createTenant(supabaseClient, testTenant)

	byteBody, err := json.Marshal(fiber.Map{
		"email":    createdTestUser.Email,
		"password": "12345678",
	})
	require.NoError(t, err)

	body := strings.NewReader(string(byteBody))
	request := httptest.NewRequest("POST", "/users/sign_in", body)
	request.Header.Set("Content-Type", "application/json")

	// We need to get the cookie
	response, err := app.Test(request, int(time.Second*5))
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

	t.Run("CreateItem", func(t *testing.T) {
		t.Run("NormalCreate1Item", func(t *testing.T) {
			byteBody, err = json.Marshal(fiber.Map{
				"items": []*fiber.Map{
					{
						"item_name": "Test 1 item",
						"stocks":    10,
					},
				},
			})
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouse/create_item/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, int(time.Second*5))
			require.Nil(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.StatusCode)

			// Clean up
			_, _, err := supabaseClient.From(repository.WarehouseTable).
				Delete("", "").
				Eq("item_name", "Test 1 item").
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/NormalCreateItem")
		})

		t.Run("NormalCreateMultipleItem", func(t *testing.T) {
			byteBody, err = json.Marshal(fiber.Map{
				"items": []*fiber.Map{
					{
						"item_name": "Test NormalCreateMultipleItem 1 item",
						"stocks":    0,
					},
					{
						"item_name": "Test NormalCreateMultipleItem 2 item",
						"stocks":    -10,
					},
					{
						"item_name": "Test NormalCreateMultipleItem 3 item",
						"stocks":    999,
					},
				},
			})
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouse/create_item/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, int(time.Second*5))
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.StatusCode)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, responseBody, "Test NormalCreateMultipleItem 1 item")
			assert.Contains(t, responseBody, "Test NormalCreateMultipleItem 2 item")
			assert.Contains(t, responseBody, "Test NormalCreateMultipleItem 3 item")
			// Clean up
			_, _, err = supabaseClient.From(repository.WarehouseTable).
				Delete("", "").
				In("item_name", []string{"Test NormalCreateMultipleItem 1 item", "Test NormalCreateMultipleItem 2 item", "Test NormalCreateMultipleItem 3 item"}).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/NormalCreateMultipleItem")
		})

		t.Run("CreateItemButNotAssociateWithRequestedTenant", func(t *testing.T) {
			someoneTenantId := 1
			byteBody, err = json.Marshal(fiber.Map{
				"items": []*fiber.Map{
					{
						"item_name": "Test 1 item",
						"stocks":    10,
					},
				},
			})
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouse/create_item/%d", someoneTenantId), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, int(time.Second*5))
			assert.NotNil(t, response)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, responseBody, "Access denied to tenant. Current user is not associate with requested tenant.")
		})
	})

	// Clean up
	_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
		Delete("", "").
		Eq("user_id", fmt.Sprint(createdTestUser.Id)).
		Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
		Execute()
	require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/TestWarehouseControllerImpl (1)")
	_, _, err = supabaseClient.From(repository.TenantTable).
		Delete("", "").
		Eq("id", fmt.Sprint(createdTenant.Id)).
		Execute()
	require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/TestWarehouseControllerImpl (2)")
	_, _, err = supabaseClient.From(repository.UserTable).
		Delete("", "").
		Eq("id", fmt.Sprint(createdTestUser.Id)).
		Execute()
	require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/TestWarehouseControllerImpl (3)")
}

func createUser(client *supabase.Client, user *model.UserRegisterForm) *model.User {
	var result *model.User
	_, err := client.From(repository.UserTable).Insert(user, false, "", "", "").Single().ExecuteTo(&result)
	if err != nil {
		panic("[DEV] Could not create user, check input")
	}

	return result
}

func createTenant(client *supabase.Client, tenant *model.Tenant) *model.Tenant {
	var result *model.Tenant
	_, err := client.From(repository.TenantTable).
		Insert(tenant, false, "", "", "").
		Single().
		ExecuteTo(&result)
	if err != nil {
		panic("[DEV] Could not create tenant, check input (1)")
	}
	_, _, err = client.From("user_mtm_tenant").
		Insert(&model.UserMtmTenant{UserId: tenant.OwnerUserId, TenantId: result.Id}, false, "", "", "").
		Execute()
	if err != nil {
		panic(fmt.Sprintf("[DEV] Could not create tenant, check input (2). Reason: %s", err.Error()))
	}

	return result
}
