package controller

import (
	"cashier-api/helper/client"
	"cashier-api/middleware"
	"cashier-api/model"
	"cashier-api/repository"
	"cashier-api/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderItemControllerImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	const STORE_ID = 1

	//SETUP//
	supabaseClient := client.CreateSupabaseClient()
	testTimeout := int((time.Second * 5).Milliseconds())
	app := fiber.New()

	//IMPLEMENTATION//
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)

	//ROUTE//
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	// These 2 protection are required
	app.Use(middleware.ProtectedRoute)
	tenantRestriction := middleware.RestrictByTenant(supabaseClient) // User only allowed to access associated tenant

	orderItemServiceMock := service.NewOrderItemServiceMock(&mock.Mock{}).(*service.OrderItemServiceMock)
	orderItemController := NewOrderItemControllerImpl(orderItemServiceMock)

	app.Post("/order_items/transactions/:tenantId", tenantRestriction, orderItemController.Transactions)

	/*
		Required to test order_item controller
		- user
		- tenant (will connect by user_mtm_tenant)
		- warehouse
		- store (the test itself)
		- cookie to simulate real request

		createUser & createTenant fn is from separate file.
	*/

	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	expectedUser := &model.UserRegisterForm{
		Name:     "TestOrderItemControllerImpl Test User",
		Email:    uniqueIdentity + "@gmail.com",
		Password: "$2a$10$V6ZP0rm./adZ9kryl3mYf.MB9IY80Y8ZCjtKslUEPWoH.9PCsX7vK",
	}
	createdTestUser := createUser(supabaseClient, expectedUser)
	require.Equal(t, expectedUser.Email, createdTestUser.Email)

	expectedTenant := &model.Tenant{
		Name:        createdTestUser.Name + "'Group",
		OwnerUserId: createdTestUser.Id,
		IsActive:    true,
	}

	createdTestTenant := createTenant(supabaseClient, expectedTenant)
	require.Equal(t, expectedTenant.Name, createdTestTenant.Name)
	require.Equal(t, expectedTenant.OwnerUserId, createdTestTenant.OwnerUserId)
	require.True(t, createdTestTenant.IsActive)

	// Cookie
	byteBody, err := json.Marshal(fiber.Map{
		"email":    createdTestUser.Email,
		"password": "12345678",
	})
	require.NoError(t, err)
	body := strings.NewReader(string(byteBody))
	request := httptest.NewRequest("POST", "/users/sign_in", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, testTimeout)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.StatusCode)

	var enterprisePOSCookie *http.Cookie = extractEnterprisePOSCookie(response.Cookies())
	require.NotNil(t, enterprisePOSCookie)

	t.Run("Transactions", func(t *testing.T) {
		t.Run("NormalTransactions", func(t *testing.T) {
			// Add query
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 30_000,
				TotalQuantity:  5,
				TotalAmount:    27_700,
				DiscountAmount: 1_300,
				SubTotal:       29_000, // 20_000 + 9_000

				Items: []*model.PurchasedItemList{
					{
						Quantity:       2,
						PurchasedPrice: 10_000,
						DiscountAmount: 500,
						TotalAmount:    19_000, // (10_000 * 2) - (500 * 2)
						ItemId:         1,
					},
					{
						Quantity:       3,
						PurchasedPrice: 3_000,
						DiscountAmount: 100,
						TotalAmount:    8_700, // (3_000 * 3) - (100 * 3)
						ItemId:         2,
					},
				},

				UserId:   createdTestUser.Id,
				TenantId: createdTestTenant.Id,
				StoreId:  STORE_ID,
			}

			orderItemServiceMock.Mock.On("Transactions", expectedParams).Return(1, nil)

			byteBody, err := json.Marshal(expectedParams)
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))

			request = httptest.NewRequest("POST", fmt.Sprintf("/order_items/transactions/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("RequestWithInvalidBody", func(t *testing.T) {
			// Add query
			errorParams := fiber.Map{
				"purchased_price": 30_000,
				"total_quantity":  5,
				"total_amount":    27_700,
				"discount_amount": 1_300,
				"sub_total":       "29_000", // Should be int

				"items": []*model.PurchasedItemList{
					{
						Quantity:       2,
						PurchasedPrice: 10_000,
						DiscountAmount: 500,
						TotalAmount:    19_000, // (10_000 * 2) - (500 * 2)
						ItemId:         1,
					},
				},

				"user_id":   createdTestUser.Id,
				"tenant_id": createdTestTenant.Id,
				"store_id":  STORE_ID,
			}

			byteBody, err := json.Marshal(errorParams)
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))

			request = httptest.NewRequest("POST", fmt.Sprintf("/order_items/transactions/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(byteResponseBody), "Something gone wrong ! The request body is malformed")
		})
	})

	t.Cleanup(func() {
		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("user_id", strconv.Itoa(createdTestUser.Id)).
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestOrderItemControllerImpl (1)")

		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestOrderItemControllerImpl (2)")

		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestUser.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestOrderItemControllerImpl (3)")
	})
}
