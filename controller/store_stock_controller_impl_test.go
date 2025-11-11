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
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase-community/supabase-go"
)

func TestStoreStockControllerImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	//SETUP//
	supabaseClient := client.CreateSupabaseClient()
	testTimeout := int((time.Second * 5).Milliseconds())
	app := fiber.New()

	//IMPLEMENTATION//
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)

	storeRepository := repository.NewStoreRepositoryImpl(supabaseClient)
	storeService := service.NewStoreServiceImpl(storeRepository)
	storeController := NewStoreControllerImpl(storeService)

	storeStockRepository := repository.NewStoreStockRepositoryImpl(supabaseClient)
	storeStockService := service.NewStoreStockServiceImpl(storeStockRepository)
	storeStockController := NewStoreStockControllerImpl(storeStockService)

	warehouseRepository := repository.NewWarehouseRepositoryImpl(supabaseClient)

	//ROUTE//
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	// These 2 protection are required
	app.Use(middleware.ProtectedRoute)
	tenantRestriction := middleware.RestrictByTenant(supabaseClient) // User only allowed to access associated tenant

	app.Get("/stores/:tenantId", tenantRestriction, storeController.GetAll)
	app.Post("/stores/:tenantId", tenantRestriction, storeController.Create)
	app.Put("/stores/set_activate/:tenantId", tenantRestriction, storeController.SetActivate)

	/*
		Required to test store_stock controller
		- user
		- tenant (will connect by user_mtm_tenant)
		- warehouse
		- store (the test itself)
		- cookie to simulate real request
	*/
	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	expectedUser := &model.UserRegisterForm{
		Name:     "TestStoreStockControllerImpl Test User",
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

	// Store
	testStoreName := fmt.Sprintf("TestStore By %s", createdTestUser.Name)
	createdTestStore := createStore(
		supabaseClient,
		createdTestTenant.Id,
		testStoreName,
	)
	require.Equal(t, testStoreName, createdTestStore.Name)

	//ROUTE// Put it here just to make it easier to check the route
	app.Get("/store_stocks/:tenantId", tenantRestriction, storeStockController.Get)
	app.Get("/store_stocks/v2/:tenantId", tenantRestriction, storeStockController.GetV2)
	app.Put("/store_stocks/edit/:tenantId", tenantRestriction, storeStockController.Edit)
	app.Put("/store_stocks/transfer_to_store_stock/:tenantId", tenantRestriction, storeStockController.TransferStockToStoreStock)
	app.Put("/store_stocks/transfer_to_warehouse/:tenantId", tenantRestriction, storeStockController.TransferStockToWarehouse)

	t.Run("Get", func(t *testing.T) {
		// Create the warehouse item first
		expectedItems := []*model.Item{
			{
				ItemName: "Test StoreStock TransferStockToStoreStock 1 Get",
				Stocks:   10,
				TenantId: createdTestTenant.Id,
				IsActive: true,
			},
		}
		createdTestItems, err := warehouseRepository.CreateItem(expectedItems)
		require.NoError(t, err)
		require.Equal(t, expectedItems[0].ItemName, createdTestItems[0].ItemName)
		itemToTransfer := createdTestItems[0]

		byteBody, err := json.Marshal(fiber.Map{
			"quantity": 5,
			"item_id":  itemToTransfer.ItemId,
			"store_id": createdTestStore.Id,
		})
		requestBody := strings.NewReader(string(byteBody))
		request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/transfer_to_store_stock/%d", createdTestTenant.Id), requestBody)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusAccepted, response.StatusCode)

		page := 1
		// limit := 5 By default will be set to 5

		t.Run("NormalGet", func(t *testing.T) {
			baseURL := fmt.Sprintf("/store_stocks/%d", createdTestTenant.Id)
			u, err := url.Parse(baseURL)
			require.NoError(t, err)

			// Add query
			query := u.Query()
			query.Set("page", strconv.Itoa(page))
			query.Set("store_id", strconv.Itoa(createdTestStore.Id))
			// query.Set("limit", limit)
			u.RawQuery = query.Encode()

			request := httptest.NewRequest("GET", u.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("RequestWithoutSpecifyingLimitAndPage", func(t *testing.T) {
			baseURL := fmt.Sprintf("/store_stocks/%d", createdTestTenant.Id)
			u, err := url.Parse(baseURL)

			query := u.Query()
			query.Set("store_id", strconv.Itoa(createdTestStore.Id))
			u.RawQuery = query.Encode()

			request := httptest.NewRequest("GET", u.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("StoreIdParamNotSpecified", func(t *testing.T) {
			request := httptest.NewRequest("GET", fmt.Sprintf("/store_stocks/%d", createdTestTenant.Id), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Cleanup(func() {
			// store_stock
			_, _, err = supabaseClient.From(repository.StoreStockTable).
				Delete("", "").
				Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
				Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/NormalTransferStockToStoreStock")
		})
	})

	t.Run("GetV2", func(t *testing.T) {
		// Create the warehouse item first
		expectedItems := []*model.Item{
			{
				ItemName: "Test StoreStock TransferStockToStoreStock 1 Get",
				Stocks:   10,
				TenantId: createdTestTenant.Id,
				IsActive: true,
			},
		}
		createdTestItems, err := warehouseRepository.CreateItem(expectedItems)
		require.NoError(t, err)
		require.Equal(t, expectedItems[0].ItemName, createdTestItems[0].ItemName)
		itemToTransfer := createdTestItems[0]

		byteBody, err := json.Marshal(fiber.Map{
			"quantity": 5,
			"item_id":  itemToTransfer.ItemId,
			"store_id": createdTestStore.Id,
		})
		requestBody := strings.NewReader(string(byteBody))
		request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/transfer_to_store_stock/%d", createdTestTenant.Id), requestBody)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusAccepted, response.StatusCode)

		page := 1
		limit := 5

		t.Run("NormalGet", func(t *testing.T) {
			baseURL := fmt.Sprintf("/store_stocks/v2/%d", createdTestTenant.Id)
			u, err := url.Parse(baseURL)
			require.NoError(t, err)

			// Add query
			query := u.Query()
			query.Set("page", strconv.Itoa(page))
			query.Set("store_id", strconv.Itoa(createdTestStore.Id))
			query.Set("limit", strconv.Itoa(limit))
			u.RawQuery = query.Encode()

			request := httptest.NewRequest("GET", u.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.Contains(t, string(byteResponseBody), createdTestItems[0].ItemName)
		})

		t.Run("StoreIdParamNotSpecified", func(t *testing.T) {
			request := httptest.NewRequest("GET", fmt.Sprintf("/store_stocks/v2/%d", createdTestTenant.Id), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Cleanup(func() {
			// store_stock
			_, _, err = supabaseClient.From(repository.StoreStockTable).
				Delete("", "").
				Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
				Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/NormalTransferStockToStoreStock")
		})
	})

	t.Run("Edit", func(t *testing.T) {
		// Create the warehouse item first
		expectedItems := []*model.Item{
			{
				ItemName: "Test StoreStock Edit 1",
				Stocks:   10,
				TenantId: createdTestTenant.Id,
				IsActive: true,
			},
		}
		createdTestItems, err := warehouseRepository.CreateItem(expectedItems)
		require.NoError(t, err)
		require.Equal(t, expectedItems[0].ItemName, createdTestItems[0].ItemName)
		itemToTransfer := createdTestItems[0]

		byteBody, err := json.Marshal(fiber.Map{
			"quantity": 5,
			"item_id":  itemToTransfer.ItemId,
			"store_id": createdTestStore.Id,
		})
		requestBody := strings.NewReader(string(byteBody))
		request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/transfer_to_store_stock/%d", createdTestTenant.Id), requestBody)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusAccepted, response.StatusCode)

		var transferredItems *model.StoreStock
		_, err = supabaseClient.From(repository.StoreStockTable).
			Select("*", "", false).
			Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
			Eq("store_id", strconv.Itoa(createdTestStore.Id)).
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Single().ExecuteTo(&transferredItems)
		require.NoError(t, err)
		require.NotNil(t, transferredItems)

		t.Run("NormalEdit", func(t *testing.T) {
			// Because we when transferring for the first time, We don't know the id so
			// get manually the transferred item

			// Here we edit the createdItems
			// Id will be already defined here
			var transferredItemsCopy model.StoreStock = *transferredItems
			transferredItemsCopy.Price = 9999

			byteBody, _ = json.Marshal(&transferredItemsCopy)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/edit/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusAccepted, response.StatusCode)

			// Check manually
			var testEditedItems *model.StoreStock
			_, err = supabaseClient.From(repository.StoreStockTable).
				Select("*", "", false).
				Eq("id", strconv.Itoa(transferredItemsCopy.Id)).
				Single().ExecuteTo(&testEditedItems)
			assert.NoError(t, err)
			assert.NotNil(t, testEditedItems)
			assert.Equal(t, transferredItemsCopy.Price, testEditedItems.Price)
		})

		t.Run("WrongStructure", func(t *testing.T) {
			wrongStructure := map[string]interface{}{
				"id": "nothing",
			}
			byteBody, _ = json.Marshal(&wrongStructure)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/edit/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("NonExistData", func(t *testing.T) {
			var transferredItemsCopy model.StoreStock = *transferredItems
			transferredItemsCopy.Id = 9999 // Non exist for this test user
			byteBody, _ = json.Marshal(&transferredItemsCopy)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/edit/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Cleanup(func() {
			// store_stock
			_, _, err = supabaseClient.From(repository.StoreStockTable).
				Delete("", "").
				Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
				Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/NormalTransferStockToStoreStock")
		})
	})

	t.Run("TransferStockToStoreStock", func(t *testing.T) {
		// Create the warehouse item first
		expectedItems := []*model.Item{
			{
				ItemName: "Test StoreStock TransferStockToStoreStock 1 TFToStoreStock",
				Stocks:   10,
				TenantId: createdTestTenant.Id,
				IsActive: true,
			},
		}
		createdTestItems, err := warehouseRepository.CreateItem(expectedItems)
		require.NoError(t, err)
		require.Equal(t, expectedItems[0].ItemName, createdTestItems[0].ItemName)
		itemToTransfer := createdTestItems[0]

		t.Run("NormalTransferStockToStoreStock", func(t *testing.T) {
			// The test itself
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": 5,
				"item_id":  itemToTransfer.ItemId,
				"store_id": createdTestStore.Id,
			})
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprintf("/store_stocks/transfer_to_store_stock/%d", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusAccepted, response.StatusCode)

			// Check if the data really counted 5
			var testValue *model.StoreStock
			_, err = supabaseClient.From(repository.StoreStockTable).
				Select("*", "exact", false).
				Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
				Single().
				ExecuteTo(&testValue)
			assert.NoError(t, err)
			assert.Equal(t, 5, testValue.Stocks)
		})

		t.Run("NotEnoughStock", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": 99,
				"item_id":  itemToTransfer.ItemId,
				"store_id": createdTestStore.Id,
			})
			body := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprint("/store_stocks/transfer_to_store_stock/", createdTestTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("ItemNeverExistAtWarehouse", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": 99,
				"item_id":  1, // This will never exist at current created warehouse
				"store_id": createdTestStore.Id,
			})
			body := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprint("/store_stocks/transfer_to_store_stock/", createdTestTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(byteBody), "[ERROR]")
		})

		t.Cleanup(func() {
			// store_stock
			_, _, err = supabaseClient.From(repository.StoreStockTable).
				Delete("", "").
				Eq("item_id", strconv.Itoa(itemToTransfer.ItemId)).
				Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/NormalTransferStockToStoreStock")
		})
	})

	t.Run("TransferStockToWarehouse", func(t *testing.T) {

	})

	t.Cleanup(func() {
		_, _, err = supabaseClient.From(repository.StoreTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestStore.Id)).
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreStockControllerImpl (1)")

		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("user_id", strconv.Itoa(createdTestUser.Id)).
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreStockControllerImpl (2)")

		_, _, err = supabaseClient.From(repository.WarehouseTable).
			Delete("", "").
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (3)")

		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreStockControllerImpl (4)")

		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestUser.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreStockControllerImpl (5)")
	})
}

func createStore(client *supabase.Client, tenantId int, storeName string) *model.Store {
	var result *model.Store
	_, err := client.From(repository.StoreTable).
		Insert(&model.Store{
			TenantId: tenantId,
			Name:     storeName,
		}, false, "", "representation", "").
		Single().
		ExecuteTo(&result)
	if err != nil {
		panic(fmt.Sprintf("[DEV] Could not create store. Reason: %s", err.Error()))
	}

	return result
}
