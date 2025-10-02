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

func TestWarehouseControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}
	testTimeout := int((time.Second * 5).Milliseconds())

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
	app.Get("/warehouses/active/:tenantId", tenantRestriction, warehouseController.GetActiveItem)
	app.Post("/warehouses/create_item/:tenantId", tenantRestriction, warehouseController.CreateItem)
	app.Post("/warehouses/find/:tenantId", tenantRestriction, warehouseController.FindById)
	app.Put("/warehouses/edit/:tenantId", tenantRestriction, warehouseController.Edit)
	app.Put("/warehouses/activate/:tenantId", tenantRestriction, warehouseController.SetActivate)

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

	t.Run("GetActiveItem", func(t *testing.T) {
		t.Run("NormalGetActiveItem", func(t *testing.T) {
			byteBody, err = json.Marshal(fiber.Map{
				"items": []*fiber.Map{
					{
						"item_name": "Test 1 item GetActiveItem NormalGetActiveItem",
						"stocks":    10,
					},
				},
			})
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Nil(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusOK, response.StatusCode)

			// It will get default param limit=5, page=1
			// build the URL
			baseURL := "/warehouses/active/" + strconv.Itoa(createdTenant.Id)
			parsedUrl, _ := url.Parse(baseURL)

			// Add query parameters
			query := parsedUrl.Query()
			query.Set("name_query", "NormalGetActiveItem") // or any value
			parsedUrl.RawQuery = query.Encode()

			// Create the request
			request := httptest.NewRequest("GET", parsedUrl.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.NoError(t, err)

			var getItemBody common.WebResponse
			responseBody, err := common.ReadBody(response.Body)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(responseBody), &getItemBody)
			require.NoError(t, err)

			// Here we take the required item only, since only 1 item created for this test scope,
			// then we safely to access the first element
			// Extract the created item
			dataMap, ok := getItemBody.Data.(map[string]interface{})
			require.True(t, ok)

			rawItems, ok := dataMap["items"].([]interface{})
			require.True(t, ok)
			require.NotEmpty(t, rawItems)

			// Marshal/unmarshal to proper type
			var items []*model.Item
			rawBytes, err := json.Marshal(rawItems)
			require.NoError(t, err)

			err = json.Unmarshal(rawBytes, &items)
			require.NoError(t, err)

			item := items[0]
			assert.Equal(t, "Test 1 item GetActiveItem NormalGetActiveItem", item.ItemName)

			// Clean up
			_, _, err = supabaseClient.From(repository.WarehouseTable).
				Delete("", "").
				Eq("item_id", strconv.Itoa(item.ItemId)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/GetActiveItem/NormalGetActiveItem")
		})
	})

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
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
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
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
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
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", someoneTenantId), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NotNil(t, response)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, responseBody, "Access denied to tenant. Current user is not associate with requested tenant.")
		})
	})

	t.Run("FindById", func(t *testing.T) {
		// Create 1 item and use for all current scope test
		byteBody, err = json.Marshal(fiber.Map{
			"items": []*fiber.Map{
				{
					"item_name": "Test 1 item FindById",
					"stocks":    10,
				},
			},
		})
		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		require.Equal(t, http.StatusOK, response.StatusCode)

		var createdItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &createdItemBody)
		require.NoError(t, err)

		// Here we take the required item only, since only 1 item created for this test scope,
		// then we safely to access the first element
		// Extract the created item
		dataMap, ok := createdItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawItems, ok := dataMap["items"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawItems)

		// Marshal/unmarshal to proper type
		var items []*model.Item
		rawBytes, err := json.Marshal(rawItems)
		require.NoError(t, err)

		err = json.Unmarshal(rawBytes, &items)
		require.NoError(t, err)

		item := items[0]

		t.Run("NormalFindById", func(t *testing.T) {
			// The test itself
			findRequestBody, err := json.Marshal(fiber.Map{
				"item_id": item.ItemId,
			})
			require.NoError(t, err)
			body = strings.NewReader(string(findRequestBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/find/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("NotFoundItemId", func(t *testing.T) {
			// The test itself
			findRequestBody, err := json.Marshal(fiber.Map{
				"item_id": nil,
			})
			require.NoError(t, err)
			body = strings.NewReader(string(findRequestBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/find/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("InvalidRequestBody", func(t *testing.T) {
			// The test itself
			findRequestBody, err := json.Marshal(fiber.Map{
				"item_id": "wrong type, should be int",
			})
			require.NoError(t, err)
			body = strings.NewReader(string(findRequestBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/find/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		// Clean up for FindById
		_, _, err = supabaseClient.From(repository.WarehouseTable).
			Delete("", "").
			Eq("item_id", fmt.Sprint(item.ItemId)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/FindById")
	})

	t.Run("Edit", func(t *testing.T) {
		// Create 2 item and use for all current scope test
		byteBody, err = json.Marshal(fiber.Map{
			"items": []*fiber.Map{
				{
					"item_name": "Test 1 item Edit",
					"stocks":    10,
				},
				{
					"item_name": "Test 2 item Edit",
					"stocks":    10,
				},
			},
		})
		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		require.Equal(t, http.StatusOK, response.StatusCode)

		var createdItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &createdItemBody)
		require.NoError(t, err)

		// Here we take the required item only, since only 1 item created for this test scope,
		// then we safely to access the first element
		// Extract the created item
		dataMap, ok := createdItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawItems, ok := dataMap["items"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawItems)

		// Marshal/unmarshal to proper type
		var items []*model.Item
		rawBytes, err := json.Marshal(rawItems)
		require.NoError(t, err)

		err = json.Unmarshal(rawBytes, &items)
		require.NoError(t, err)

		item1 := items[0]
		item2 := items[1]

		t.Run("NormalEdit", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": -3,
				"item": fiber.Map{
					"item_id":   item1.ItemId,
					"item_name": item1.ItemName + " edited",
				},
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/edit/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusAccepted, response.StatusCode)

			// Check
			var checkItem1 *model.Item
			_, err = supabaseClient.From(repository.WarehouseTable).
				Select("*", "", false).
				Eq("item_id", strconv.Itoa(item1.ItemId)).
				Single().
				ExecuteTo(&checkItem1)
			assert.NoError(t, err)
			assert.Equal(t, item1.Stocks-3, checkItem1.Stocks)
			assert.Equal(t, item1.ItemName+" edited", checkItem1.ItemName)
		})

		t.Run("InvalidQuantitiesByDecreasingTooMuchWhileStocksNotEnough", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": -999,
				"item": fiber.Map{
					"item_id":   item1.ItemId,
					"item_name": item1.ItemName,
				},
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/edit/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("InvalidQuantitiesByTooMuchIncreasingOrDecreasing", func(t *testing.T) {
			// Allowed increasing or decreasing quantities are between -999 and 999
			byteBody, err := json.Marshal(fiber.Map{
				"quantity": -1000,
				"item": fiber.Map{
					"item_id":   item1.ItemId,
					"item_name": item1.ItemName,
				},
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/edit/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = json.Marshal(fiber.Map{
				"quantity": 1000,
				"item": fiber.Map{
					"item_id":   item1.ItemId,
					"item_name": item1.ItemName,
				},
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/edit/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		// Clean up for Edit
		_, _, err = supabaseClient.From(repository.WarehouseTable).
			Delete("", "").
			In("item_id", []string{fmt.Sprint(item1.ItemId), fmt.Sprint(item2.ItemId)}).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/Edit")
	})

	t.Run("SetActivate", func(t *testing.T) {
		// Create 2 item and use for all current scope test
		byteBody, err = json.Marshal(fiber.Map{
			"items": []*fiber.Map{
				{
					"item_name": "Test 1 item SetActivate",
					"stocks":    10,
				},
			},
		})
		require.NoError(t, err)
		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/warehouses/create_item/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		require.Equal(t, http.StatusOK, response.StatusCode)

		var createdItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &createdItemBody)
		require.NoError(t, err)

		// Here we take the required item only, since only 1 item created for this test scope,
		// then we safely to access the first element
		// Extract the created item
		dataMap, ok := createdItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawItems, ok := dataMap["items"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawItems)

		// Marshal/unmarshal to proper type
		var items []*model.Item
		rawBytes, err := json.Marshal(rawItems)
		require.NoError(t, err)

		err = json.Unmarshal(rawBytes, &items)
		require.NoError(t, err)

		item1 := items[0]

		t.Run("NormalActivate", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"item_id":  item1.ItemId,
				"set_into": false,
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/activate/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusAccepted, response.StatusCode)
		})

		t.Run("NormalActivate", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"item_id":  item1.ItemId,
				"set_into": false,
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/activate/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusAccepted, response.StatusCode)
		})

		t.Run("InCaseSetIntoIsNull", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"item_id":  item1.ItemId,
				"set_into": nil, // By default when converted at warehouse_controller.SetActivate, GO will be transformed into false
			})
			require.NoError(t, err)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/warehouses/activate/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Equal(t, http.StatusAccepted, response.StatusCode)
		})

		// Clean up for SetActivate
		_, _, err = supabaseClient.From(repository.WarehouseTable).
			Delete("", "").
			Eq("item_id", fmt.Sprint(item1.ItemId)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl/CreateItem/SetActivate")
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
