package controller

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/middleware"
	"cashier-api/model"
	"cashier-api/repository"
	"cashier-api/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	now := time.Now()
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
	app.Post("/order_items/search/:tenantId", tenantRestriction, orderItemController.Get)
	app.Get("/order_items/details/:tenantId", tenantRestriction, orderItemController.FindById) // Params=order_item_id

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

	t.Run("Get", func(t *testing.T) {
		t.Run("NormalGet", func(t *testing.T) {
			expectedResponse := &OrderItemControllerGetResponse{
				OrderItems: []*model.OrderItem{
					{
						Id:             1,
						PurchasedPrice: 1000,
						TotalQuantity:  1,
						TotalAmount:    1000,
						DiscountAmount: 0,
						Subtotal:       1000,
						CreatedAt:      &now,
						StoreId:        STORE_ID,
						TenantId:       createdTestTenant.Id,
					},
					{
						Id:             2,
						PurchasedPrice: 5000,
						TotalQuantity:  1,
						TotalAmount:    5000,
						DiscountAmount: 0,
						Subtotal:       5000,
						CreatedAt:      &now,
						StoreId:        STORE_ID,
						TenantId:       createdTestTenant.Id,
					},
				},
			}

			startDate := int64(1600000000)
			endDate := int64(1700000000)
			dateFilter := &query.DateFilter{
				StartDate: &startDate,
				EndDate:   &endDate,
			}
			body := OrderItemControllerGetRequest{
				TenantId: 1,
				StoreId:  1,
				Limit:    20,
				Page:     1,
				Filters: []*query.QueryFilter{
					{
						Column:    "created_at",
						Ascending: true,
					},
				},
				DateFilter: dateFilter,
			}

			byteBody, err := json.Marshal(&body)
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))

			orderItemServiceMock.Mock = &mock.Mock{}
			orderItemServiceMock.Mock.On("Get", body.TenantId, body.StoreId, body.Limit, body.Page, body.Filters, body.DateFilter).
				Return(expectedResponse.OrderItems, len(expectedResponse.OrderItems), nil)

			request = httptest.NewRequest("POST", fmt.Sprintf("/order_items/search/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			var responseBody OrderItemControllerGetResponse
			err = json.Unmarshal(byteResponseBody, &responseBody)
			assert.NoError(t, err)
			assert.NotNil(t, responseBody)
			assert.Equal(t, expectedResponse.TotalCount, responseBody.TotalCount)
			assert.Equal(t, expectedResponse.Limit, responseBody.Limit)
			assert.Equal(t, expectedResponse.RequestedByTenantId, responseBody.RequestedByTenantId)
		})
	})

	t.Run("Transactions", func(t *testing.T) {
		t.Run("NormalTransactions", func(t *testing.T) {
			// Add query
			expectedParams := &repository.CreateTransactionParams{
				PurchasedPrice: 30_000,
				TotalQuantity:  5,
				TotalAmount:    27_700,
				DiscountAmount: 1_300,
				SubTotal:       29_000, // 20_000 + 9_000

				Items: []*model.PurchasedItem{
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

			orderItemServiceMock.Mock = &mock.Mock{}
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

				"items": []*model.PurchasedItem{
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

	t.Run("FindById", func(t *testing.T) {
		t.Run("NormalFindById", func(t *testing.T) {
			baseURL := fmt.Sprintf("/order_items/details/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			params := url.Values{}
			params.Add("order_item_id", "1")
			parsedURL.RawQuery = params.Encode()

			expectedOrderItem := &model.OrderItem{
				Id:             1,
				Subtotal:       29_000,
				PurchasedPrice: 30_000,
				TotalQuantity:  5,
				TotalAmount:    27_700,
				DiscountAmount: 1_300,
				StoreId:        STORE_ID,
				TenantId:       createdTestTenant.Id,
				CreatedAt:      &now,
			}

			expectedPurchasedItemList := []*model.PurchasedItem{
				{
					Id:             1,
					Quantity:       2,
					PurchasedPrice: 10_000,
					DiscountAmount: 500,
					TotalAmount:    19_000, // (10_000 * 2) - (500 * 2)
					ItemId:         1,
					OrderItemId:    expectedOrderItem.Id,
				},
				{
					Id:             2,
					Quantity:       3,
					PurchasedPrice: 3_000,
					DiscountAmount: 100,
					TotalAmount:    8_700, // (3_000 * 3) - (100 * 3)
					ItemId:         2,
					OrderItemId:    expectedOrderItem.Id,
				},
			}

			orderItemServiceMock.Mock = &mock.Mock{}
			orderItemServiceMock.Mock.On("FindById", expectedOrderItem.Id, createdTestTenant.Id).Return(expectedOrderItem, expectedPurchasedItemList, nil)

			request = httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("ErrorResponseFromService", func(t *testing.T) {
			baseURL := fmt.Sprintf("/order_items/details/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			require.NoError(t, err)
			params := url.Values{}
			params.Add("order_item_id", "999")
			parsedURL.RawQuery = params.Encode()

			orderItemServiceMock.Mock = &mock.Mock{}
			orderItemServiceMock.Mock.On("FindById", 999, createdTestTenant.Id).Return(nil, nil, errors.New("Tenant id or Order item id Required !"))

			request = httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("OrderItemIdNotSpecify", func(t *testing.T) {
			baseURL := fmt.Sprintf("/order_items/details/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			require.NoError(t, err)
			//params := url.Values{}
			//params.Add("order_item_id", "1")
			//parsedURL.RawQuery = params.Encode()

			request = httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
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
