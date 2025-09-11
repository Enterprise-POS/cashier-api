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

	// create items
	// createdItems, err := createItems(supabase, []*model.Item{
	// 	{
	// 		ItemName: "Apple",
	// 		Stocks:   10,
	// 		TenantId: createdTenant.Id,
	// 	},
	// })
	// require.NoError(t, err)

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

	t.Cleanup(func() {
		_, _, err = supabase.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("user_id", fmt.Sprint(createdTestUser.Id)).
			Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl (1)")
		_, _, err = supabase.From(repository.TenantTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl (2)")
		_, _, err = supabase.From(repository.UserTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTestUser.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestWarehouseControllerImpl (3)")
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
