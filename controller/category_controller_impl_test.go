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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestCategoryControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}
	testTimeout := int((time.Second * 5).Milliseconds())
	app := fiber.New()
	supabase := client.CreateSupabaseClient()
	userRepository := repository.NewUserRepositoryImpl(supabase)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	// protected only login user
	app.Use(middleware.ProtectedRoute)
	categoryRepository := repository.NewCategoryRepositoryImpl(supabase)
	categoryService := service.NewCategoryServiceImpl(categoryRepository)
	categoryController := NewCategoryControllerImpl(categoryService)

	tenantRestriction := middleware.RestrictByTenant(supabase) // User only allowed to access associated tenant
	app.Post("/categories/items_by_category_id/:tenantId", tenantRestriction, categoryController.GetItemsByCategoryId)
	app.Post("/categories/category_with_items/:tenantId", tenantRestriction, categoryController.GetCategoryWithItems)
	app.Post("/categories/create/:tenantId", tenantRestriction, categoryController.Create)
	app.Post("/categories/register/:tenantId", tenantRestriction, categoryController.Register)
	app.Get("/categories/:tenantId", tenantRestriction, categoryController.Get)
	app.Put("/categories/update/:tenantId", tenantRestriction, categoryController.Update)
	app.Delete("/categories/unregister/:tenantId", tenantRestriction, categoryController.Unregister)
	app.Delete("/categories/:tenantId", tenantRestriction, categoryController.Delete)

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

	t.Run("Get", func(t *testing.T) {
		dummyCategories := []*model.Category{
			{
				CategoryName: "Category Get",
				TenantId:     createdTenant.Id,
			},
		}

		var expectedDummyCategories []*model.Category
		_, err = supabase.From(repository.CategoryTable).
			Insert(dummyCategories, false, "", "representation", "").
			ExecuteTo(&expectedDummyCategories)
		require.NoError(t, err)

		t.Run("NormalGet", func(t *testing.T) {
			request := httptest.NewRequest("GET", fmt.Sprintf("/categories/%d", createdTenant.Id), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), dummyCategories[0].CategoryName)
		})

		t.Run("UsingNameQuery", func(t *testing.T) {
			// Create dummy category only for this scope
			dummyCategories := []*model.Category{
				{
					CategoryName: "CategoryGet ZZZ", // Only 15 characters
					TenantId:     createdTenant.Id,
				},
			}

			var expectedDummyCategories []*model.Category
			_, err = supabase.From(repository.CategoryTable).
				Insert(dummyCategories, false, "", "representation", "").
				ExecuteTo(&expectedDummyCategories)
			require.NoError(t, err)

			baseURL := fmt.Sprintf("/categories/%d", createdTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			params := url.Values{}
			params.Add("limit", "1") // Make sure only 1 applied
			params.Add("name_query", dummyCategories[0].CategoryName)

			parsedURL.RawQuery = params.Encode()

			request := httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), dummyCategories[0].CategoryName)

			t.Cleanup(func() {
				_, _, err := supabase.From(repository.CategoryTable).
					Delete("", "").
					Eq("id", fmt.Sprint(expectedDummyCategories[0].Id)).
					Execute()
				require.NoError(t, err)
			})
		})

		t.Run("OverlapRange", func(t *testing.T) {
			page, limit := "2", "5"
			url, err := url.Parse(fmt.Sprintf("/categories/%d", createdTenant.Id))
			require.NoError(t, err)

			query := url.Query()
			query.Set("page", page)
			query.Set("limit", limit)
			url.RawQuery = query.Encode()

			request := httptest.NewRequest("GET", url.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), "Requested range not satisfiable")
		})

		t.Cleanup(func() {
			_, _, err := supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(expectedDummyCategories[0].Id)).
				Execute()
			require.NoError(t, err)
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		dummyCategories := []*model.Category{
			{
				CategoryName: "GItemsByCtgry",
				TenantId:     createdTenant.Id,
			},
		}

		var categories []*model.Category
		_, err = supabase.From(repository.CategoryTable).
			Insert(dummyCategories, false, "", "representation", "").
			ExecuteTo(&categories)
		require.NoError(t, err)

		testItems := []*model.Item{
			{
				ItemName: "Test 1 Get Category With Items",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
			{
				ItemName: "Test 2 Get Category With Items",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
			{
				ItemName: "Test 3 Get Category With Items",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
		}

		createdItems, err := createItems(supabase, testItems)
		require.NoError(t, err)
		require.Len(t, testItems, len(testItems))

		// Register
		requestBody := fiber.Map{
			"tobe_registers": []*model.CategoryMtmWarehouse{
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[0].ItemId,
				},
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[1].ItemId,
				},
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[2].ItemId,
				},
			},
		}
		byteBody, err = json.Marshal(&requestBody)
		require.NoError(t, err)

		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		require.Nil(t, err)
		require.NotNil(t, response)
		require.Equal(t, http.StatusCreated, response.StatusCode)

		byteBody, err = io.ReadAll(response.Body)
		require.NoError(t, err)
		require.Equal(t, "Created", string(byteBody))

		t.Run("NormalGetCategoryWithItems", func(t *testing.T) {
			page, limit := 1, 5
			requestBody := fiber.Map{
				"page":  page,
				"limit": limit,
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/category_with_items/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("PageOverflow", func(t *testing.T) {
			page, limit := 100, 5
			requestBody := fiber.Map{
				"page":  page,
				"limit": limit,
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/category_with_items/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)

			// Will return empty slice if page is overflow at RPC
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Cleanup(func() {
			// It will delete also category_mtm_warehouse
			_, _, err = supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(categories[0].Id)).
				Execute()
			require.NoError(t, err)

			// Items
			_, _, err = supabase.From(repository.WarehouseTable).
				Delete("", "").
				In("item_id", []string{fmt.Sprint(createdItems[0].ItemId), fmt.Sprint(createdItems[1].ItemId), fmt.Sprint(createdItems[2].ItemId)}).
				Execute()
		})
	})

	t.Run("GetItemsByCategory", func(t *testing.T) {
		dummyCategories := []*model.Category{
			{
				CategoryName: "GItemsByCtgry",
				TenantId:     createdTenant.Id,
			},
		}

		var categories []*model.Category
		_, err = supabase.From(repository.CategoryTable).
			Insert(dummyCategories, false, "", "representation", "").
			ExecuteTo(&categories)
		require.NoError(t, err)

		testItems := []*model.Item{
			{
				ItemName: "Test 1 Get Items By Category",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
			{
				ItemName: "Test 2 Get Items By Category",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
			{
				ItemName: "Test 3 Get Items By Category",
				Stocks:   10,
				IsActive: true,
				TenantId: createdTenant.Id,
			},
		}

		createdItems, err := createItems(supabase, testItems)
		require.NoError(t, err)
		require.Len(t, testItems, len(testItems))

		// Register
		requestBody := fiber.Map{
			"tobe_registers": []*model.CategoryMtmWarehouse{
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[0].ItemId,
				},
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[1].ItemId,
				},
				{
					CategoryId: categories[0].Id,
					ItemId:     createdItems[2].ItemId,
				},
			},
		}
		byteBody, err = json.Marshal(&requestBody)
		require.NoError(t, err)

		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		require.Nil(t, err)
		require.NotNil(t, response)
		require.Equal(t, http.StatusCreated, response.StatusCode)

		byteBody, err = io.ReadAll(response.Body)
		require.NoError(t, err)
		require.Equal(t, "Created", string(byteBody))

		t.Run("NormalGetItemsByCategory", func(t *testing.T) {
			page, limit := 1, 5
			requestBody := fiber.Map{
				"category_id": categories[0].Id,
				"page":        page,
				"limit":       limit,
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/items_by_category_id/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Run("PageOverflow", func(t *testing.T) {
			page, limit := 100, 5
			requestBody := fiber.Map{
				"category_id": categories[0].Id,
				"page":        page,
				"limit":       limit,
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/items_by_category_id/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)

			// Will return empty slice if page is overflow at RPC
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})

		t.Cleanup(func() {
			// It will delete also category_mtm_warehouse
			_, _, err = supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(categories[0].Id)).
				Execute()
			require.NoError(t, err)

			// Items
			_, _, err = supabase.From(repository.WarehouseTable).
				Delete("", "").
				In("item_id", []string{fmt.Sprint(createdItems[0].ItemId), fmt.Sprint(createdItems[1].ItemId), fmt.Sprint(createdItems[2].ItemId)}).
				Execute()
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			requestBody := fiber.Map{
				"categories": []string{"Fast Food", "Japanese Food"},
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/create/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			t.Cleanup(func() {
				var getItemBody common.WebResponse
				responseBody, err := common.ReadBody(response.Body)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(responseBody), &getItemBody)
				require.NoError(t, err)

				dataMap, ok := getItemBody.Data.(map[string]interface{})
				require.True(t, ok)

				rawCategories, ok := dataMap["categories"].([]interface{})
				require.True(t, ok)
				require.NotEmpty(t, rawCategories)

				// Marshal/unmarshal to proper type
				var categories []*model.Category
				rawBytes, err := json.Marshal(rawCategories)
				require.NoError(t, err)

				err = json.Unmarshal(rawBytes, &categories)
				require.NoError(t, err)

				_, count, err := supabase.From(repository.CategoryTable).
					Delete("", "exact").
					In("id", []string{fmt.Sprint(categories[0].Id), fmt.Sprint(categories[1].Id)}).
					Execute()
				require.NoError(t, err)
				require.Equal(t, 2, int(count))
			})
		})

		t.Run("EmptyRequest", func(t *testing.T) {
			requestBody := fiber.Map{
				"categories": []string{},
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/create/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode) // 400
		})

		t.Run("SameCategoryName", func(t *testing.T) {
			requestBody := fiber.Map{
				"categories": []string{"Same Category", "Same Category"},
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/create/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode) // 400

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), "Something gone wrong. Duplicate category detected")
		})
	})

	t.Run("Register", func(t *testing.T) {
		// Create category for current Register scope only
		dummyCategories := []string{"Test Register"}

		// Create new category for register
		requestBody := fiber.Map{
			"categories": dummyCategories,
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

		// Get the id
		var getItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &getItemBody)
		require.NoError(t, err)

		dataMap, ok := getItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawCategories, ok := dataMap["categories"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawCategories)

		// Marshal/unmarshal to proper type
		var categories []*model.Category
		rawBytes, err := json.Marshal(&rawCategories)
		require.NoError(t, err)

		err = json.Unmarshal(rawBytes, &categories)
		require.NoError(t, err)

		dummyItems := []*model.Item{
			{
				ItemName: "Test Category Register 1",
				Stocks:   10,
				TenantId: createdTenant.Id,
				IsActive: true,
			},
		}

		items, err := createItems(supabase, dummyItems)
		require.NoError(t, err)

		t.Run("NormalRegister", func(t *testing.T) {
			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{
					{
						"category_id": categories[0].Id,
						"item_id":     items[0].ItemId,
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusCreated, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Equal(t, "Created", string(byteBody))
		})

		t.Run("InvalidEmptyTobeRegisters", func(t *testing.T) {
			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{}, // Invalid, Empty/len = 0 is not allowed
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), "Invalid request. Fill at least 1 category and item to be add")
		})

		t.Run("InvalidTobeRegistersBody", func(t *testing.T) {
			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{
					{
						"category_id": 0,
						"item_id":     items[0].ItemId,
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), fmt.Sprintf("Required item id or category id is not valid. item id: %d, category id: %d", items[0].ItemId, 0))

			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{
					{
						"category_id": categories[0].Id,
						"item_id":     0,
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), fmt.Sprintf("Required item id or category id is not valid. item id: %d, category id: %d", 0, categories[0].Id))
		})

		t.Run("NotExistCategoryId", func(t *testing.T) {
			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{
					{
						"category_id": 1,   // Not valid
						"item_id":     999, // Not valid / not even exist
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), "Forbidden action ! non exist category id or item id")
		})

		t.Cleanup(func() {
			// category_mtm_warehouse
			_, _, err := supabase.From(repository.CategoryMtmWarehouseTable).
				Delete("", "").
				Eq("category_id", fmt.Sprint(categories[0].Id)).
				Eq("item_id", fmt.Sprint(items[0].ItemId)).
				Execute()
			require.NoError(t, err)

			// category
			_, _, err = supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(categories[0].Id)).
				Execute()
			require.NoError(t, err)

			// warehouse item
			_, _, err = supabase.From(repository.WarehouseTable).
				Delete("", "").
				Eq("item_id", fmt.Sprint(items[0].ItemId)).
				Execute()
			require.NoError(t, err)
		})
	})

	t.Run("Unregister", func(t *testing.T) {
		// Create category for current Register scope only
		dummyCategories := []string{"Test Unregister"}

		// Create new category for register
		requestBody := fiber.Map{
			"categories": dummyCategories,
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

		// Get the id
		var getItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &getItemBody)
		require.NoError(t, err)

		dataMap, ok := getItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawCategories, ok := dataMap["categories"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawCategories)

		// Marshal/unmarshal to proper type
		var categories []*model.Category
		rawBytes, err := json.Marshal(&rawCategories)
		require.NoError(t, err)

		err = json.Unmarshal(rawBytes, &categories)
		require.NoError(t, err)

		dummyItems := []*model.Item{
			{
				ItemName: "Test Category Register 1",
				Stocks:   10,
				TenantId: createdTenant.Id,
				IsActive: true,
			},
		}

		items, err := createItems(supabase, dummyItems)
		require.NoError(t, err)

		t.Run("NormalUnregister", func(t *testing.T) {
			// Use register path to register item first
			requestBody = fiber.Map{
				"tobe_registers": []fiber.Map{
					{
						"category_id": categories[0].Id,
						"item_id":     items[0].ItemId,
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			require.Nil(t, err)
			require.NotNil(t, response)
			require.Equal(t, http.StatusCreated, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			require.Equal(t, "Created", string(byteBody))

			// The test itself
			requestBody := &model.CategoryMtmWarehouse{
				CategoryId: categories[0].Id,
				ItemId:     items[0].ItemId,
			}
			byteBody, err = json.Marshal(requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("DELETE", fmt.Sprintf("/categories/unregister/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusNoContent, response.StatusCode)
		})

		t.Run("NonExistCategoryIdOrItemId", func(t *testing.T) {
			// Use register path to register item first
			requestBody = fiber.Map{
				"tobe_registers": []*model.CategoryMtmWarehouse{
					{
						CategoryId: 999,
						ItemId:     999,
					},
				},
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), "Forbidden action ! non exist category id or item id")
		})

		t.Cleanup(func() {
			// category_mtm_warehouse
			// Don't need to delete category_mtm_warehouse because already deleted by
			// unregister method

			// category
			_, _, err = supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(categories[0].Id)).
				Execute()
			require.NoError(t, err)

			// warehouse item
			_, _, err = supabase.From(repository.WarehouseTable).
				Delete("", "").
				Eq("item_id", fmt.Sprint(items[0].ItemId)).
				Execute()
			require.NoError(t, err)
		})
	})

	t.Run("Update", func(t *testing.T) {
		// Update category itself not warehouse item
		// Create category for current Register scope only
		dummyCategories := []string{"Test Update"}

		// Create new category for register
		requestBody := fiber.Map{
			"categories": dummyCategories,
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

		// Get the id
		var getItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &getItemBody)
		require.NoError(t, err)

		dataMap, ok := getItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawCategories, ok := dataMap["categories"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawCategories)

		rawBytes, err := json.Marshal(&rawCategories)
		require.NoError(t, err)

		// Marshal/unmarshal to proper type
		var categories []*model.Category
		err = json.Unmarshal(rawBytes, &categories)
		require.NoError(t, err)

		t.Run("NormalUpdate", func(t *testing.T) {
			requestBody := fiber.Map{
				"category_id":   categories[0].Id,
				"category_name": "Updated ctg", // Tobe changed category name / will update name
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/categories/update/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), requestBody["category_name"])
		})

		t.Run("NonExistCategoryIdOrTenantId", func(t *testing.T) {
			requestBody := fiber.Map{
				"category_id":   999,
				"category_name": "Updated ctg", // Tobe changed category name / will update name
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/categories/update/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), fmt.Sprintf("Nothing is updated from category_id: %d, tenant_id: %d", 999, createdTenant.Id))
		})

		t.Run("InvalidUpdateCategoryName", func(t *testing.T) {
			// Too long for category name
			requestBody := fiber.Map{
				"category_id":   categories[0].Id,
				"category_name": "Too long name for categories", // Tobe changed category name / will update name
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/categories/update/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), fmt.Sprintf("Current category name is not allowed: %s", requestBody["category_name"]))

			// Invalid characters
			requestBody = fiber.Map{
				"category_id":   categories[0].Id,
				"category_name": "Inval!d", // Tobe changed category name / will update name
			}
			byteBody, err = json.Marshal(&requestBody)
			require.NoError(t, err)

			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("PUT", fmt.Sprintf("/categories/update/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteBody, err = io.ReadAll(response.Body)
			assert.Contains(t, string(byteBody), fmt.Sprintf("Current category name is not allowed: %s", requestBody["category_name"]))
		})

		t.Cleanup(func() {
			// Category
			_, _, err = supabase.From(repository.CategoryTable).
				Delete("", "").
				Eq("id", fmt.Sprint(categories[0].Id)).
				Execute()
			require.NoError(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		// Update category itself not warehouse item
		// Create category for current Register scope only
		dummyCategories := []string{"Test Delete"}

		// Create new category for register
		requestBody := fiber.Map{
			"categories": dummyCategories,
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

		// Get the id
		var getItemBody common.WebResponse
		responseBody, err := common.ReadBody(response.Body)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(responseBody), &getItemBody)
		require.NoError(t, err)

		dataMap, ok := getItemBody.Data.(map[string]interface{})
		require.True(t, ok)

		rawCategories, ok := dataMap["categories"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, rawCategories)

		rawBytes, err := json.Marshal(&rawCategories)
		require.NoError(t, err)

		// Marshal/unmarshal to proper type
		var categories []*model.Category
		err = json.Unmarshal(rawBytes, &categories)
		require.NoError(t, err)

		// Add warehouse items to current category
		dummyItems := []*model.Item{
			{
				ItemName: "Test Category Delete 1",
				Stocks:   10,
				TenantId: createdTenant.Id,
				IsActive: true,
			},
			{
				ItemName: "Test Category Delete 2",
				Stocks:   10,
				TenantId: createdTenant.Id,
				IsActive: true,
			},
		}

		items, err := createItems(supabase, dummyItems)
		require.NoError(t, err)
		requestBody = fiber.Map{
			"tobe_registers": []fiber.Map{
				{
					"category_id": categories[0].Id, // Must be the same otherwise test will be fail
					"item_id":     items[0].ItemId,
				},
				{
					"category_id": categories[0].Id, // Must be the same otherwise test will be fail
					"item_id":     items[1].ItemId,
				},
			},
		}
		byteBody, err = json.Marshal(&requestBody)
		require.NoError(t, err)

		// Register created items into current category
		body = strings.NewReader(string(byteBody))
		request = httptest.NewRequest("POST", fmt.Sprintf("/categories/register/%d", createdTenant.Id), body)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err = app.Test(request, testTimeout)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusCreated, response.StatusCode)

		byteBody, err = io.ReadAll(response.Body)
		assert.Equal(t, "Created", string(byteBody))

		t.Run("NormalDelete", func(t *testing.T) {
			requestBody = fiber.Map{
				"category_id": categories[0].Id,
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("DELETE", fmt.Sprintf("/categories/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusNoContent, response.StatusCode)
			byteBody, err = io.ReadAll(response.Body)

			// Check if category actually deleted
			var testCategory []*model.Category
			_, err = supabase.From(repository.CategoryTable).
				Select("*", "exact", false).
				Eq("id", fmt.Sprint(categories[0].Id)).
				ExecuteTo(&testCategory)
			assert.NoError(t, err)
			assert.Len(t, testCategory, 0)
		})

		t.Run("NoCategoryDeleted", func(t *testing.T) {
			/*
				If only run current test scope, test might be fail
				due to warehouse items not yet deleted
			*/
			requestBody = fiber.Map{
				"category_id": 1, // Non exist category for current tenant
			}
			byteBody, err = json.Marshal(&requestBody)
			body = strings.NewReader(string(byteBody))
			request = httptest.NewRequest("DELETE", fmt.Sprintf("/categories/%d", createdTenant.Id), body)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Cleanup(func() {
			// Warehouse items
			_, _, err = supabase.From(repository.WarehouseTable).
				Delete("", "").
				In("item_id", []string{fmt.Sprint(items[0].ItemId), fmt.Sprint(items[1].ItemId)}).
				Execute()
			require.NoError(t, err)
		})
	})

	t.Cleanup(func() {
		_, _, err = supabase.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("user_id", fmt.Sprint(createdTestUser.Id)).
			Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestCategoryControllerImpl (1)")
		_, _, err = supabase.From(repository.TenantTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestCategoryControllerImpl (2)")
		_, _, err = supabase.From(repository.UserTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTestUser.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestCategoryControllerImpl (3)")
	})
}

func createItems(supabase *supabase.Client, items []*model.Item) ([]*model.Item, error) {
	var results []*model.Item
	_, err := supabase.From(repository.WarehouseTable).Insert(items, false, "", "", "").ExecuteTo(&results)
	if err != nil {
		return nil, err
	} else {
		return results, nil
	}
}
