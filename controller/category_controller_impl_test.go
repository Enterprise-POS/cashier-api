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
	app.Post("/categories/create/:tenantId", tenantRestriction, categoryController.Create)
	app.Post("/categories/register/:tenantId", tenantRestriction, categoryController.Register)
	app.Get("/categories/:tenantId", tenantRestriction, categoryController.Get)

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
