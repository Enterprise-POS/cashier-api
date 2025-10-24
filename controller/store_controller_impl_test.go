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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	/*
		Required to test store controller
		- user
		- tenant (will connect by user_mtm_tenant)
		- store (the test itself)
		- cookie to simulate real request
	*/
	testTimeout := int((time.Second * 5).Milliseconds())
	app := fiber.New()
	supabaseClient := client.CreateSupabaseClient()
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	app.Use(middleware.ProtectedRoute)

	storeRepository := repository.NewStoreRepositoryImpl(supabaseClient)
	storeService := service.NewStoreServiceImpl(storeRepository)
	storeController := NewStoreControllerImpl(storeService)

	// User
	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	testUser := &model.UserRegisterForm{
		Name:     "Test_StoreController" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		Email:    uniqueIdentity + "@gmail.com",
		Password: "$2a$10$V6ZP0rm./adZ9kryl3mYf.MB9IY80Y8ZCjtKslUEPWoH.9PCsX7vK",
	}
	createdTestUser := createUser(supabaseClient, testUser)

	// Tenant
	testTenant := &model.Tenant{
		Name:        createdTestUser.Name + "'Group",
		OwnerUserId: createdTestUser.Id,
		IsActive:    true}
	createdTestTenant := createTenant(supabaseClient, testTenant)

	byteBody, err := json.Marshal(fiber.Map{
		"email":    createdTestUser.Email,
		"password": "12345678",
	})
	require.NoError(t, err)

	body := strings.NewReader(string(byteBody))
	request := httptest.NewRequest("POST", "/users/sign_in", body)
	request.Header.Set("Content-Type", "application/json")

	// Cookie
	response, err := app.Test(request, testTimeout)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusOK, response.StatusCode)
	var enterprisePOSCookie *http.Cookie = extractEnterprisePOSCookie(response.Cookies())
	require.NotNil(t, enterprisePOSCookie)

	//ROUTE//
	tenantRestriction := middleware.RestrictByTenant(supabaseClient) // User only allowed to access associated tenant
	app.Post("/stores/:tenantId", tenantRestriction, storeController.Create)
	app.Get("/stores/:tenantId", tenantRestriction, storeController.GetAll)
	app.Put("/stores/set_activate/:tenantId", tenantRestriction, storeController.SetActivate)

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCrate", func(t *testing.T) {
			testStoreName := "Test Store Controller"
			byteBody, err := json.Marshal(fiber.Map{
				"name": testStoreName,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			var responseBody struct {
				Code   int    `json:"code"`
				Status string `json:"status"`
				Data   struct {
					CreatedStore *model.Store `json:"created_store"`
				} `json:"data"`
			}
			err = json.Unmarshal(byteResponseBody, &responseBody)
			require.NoError(t, err)
			assert.Equal(t, testStoreName, responseBody.Data.CreatedStore.Name)
			assert.NotEqual(t, 0, responseBody.Data.CreatedStore.Id)
			assert.NotNil(t, responseBody.Data.CreatedStore.CreatedAt)
			assert.Equal(t, http.StatusOK, responseBody.Code)
			assert.Equal(t, common.StatusSuccess, responseBody.Status)

			t.Cleanup(func() {
				_, _, err = supabaseClient.From(repository.StoreTable).
					Delete("", "").
					Eq("id", strconv.Itoa(responseBody.Data.CreatedStore.Id)).
					Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
					Execute()
				require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (1)")
			})
		})

		// Wrong json
		t.Run("WrongRequestBodyDataType", func(t *testing.T) {
			testStoreName := 1 // Should be string
			byteBody, err := json.Marshal(fiber.Map{
				"name": testStoreName,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(byteResponseBody), "Something gone wrong ! The request body is malformed")
		})
	})

	t.Run("GetAll", func(t *testing.T) {
		// A setup for this test scope
		testStoreNames := []string{"Test Store Controller Get All1", "Test Store Controller Get All2"}
		var createdTestStores = []*model.Store{}
		for _, storeName := range testStoreNames {
			byteBody, err := json.Marshal(fiber.Map{
				"name": storeName,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			fmt.Println(string(byteResponseBody))

			var responseBody struct {
				Code   int    `json:"code"`
				Status string `json:"status"`
				Data   struct {
					CreatedStore *model.Store `json:"created_store"`
				} `json:"data"`
			}
			err = json.Unmarshal(byteResponseBody, &responseBody)
			createdTestStores = append(createdTestStores, responseBody.Data.CreatedStore)
		}

		t.Run("NormalGetAll", func(t *testing.T) {
			page := 1
			limit := 2
			includeNonActive := true
			byteBody, err := json.Marshal(fiber.Map{
				"page":               page,
				"limit":              limit,
				"include_non_active": includeNonActive,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("GET", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			fmt.Println(string(byteResponseBody))

			var responseBody struct {
				Code   int    `json:"code"`
				Status string `json:"status"`
				Data   struct {
					Stores []*model.Store `json:"stores"` // We only take the stores property
				} `json:"data"`
			}
			err = json.Unmarshal(byteResponseBody, &responseBody)
			for _, store := range responseBody.Data.Stores {
				assert.Contains(t, testStoreNames, store.Name)
				assert.True(t, store.IsActive)
				assert.NotNil(t, store.CreatedAt)
				assert.NotEqual(t, 0, store.Id)
			}
		})

		t.Run("WrongRequestDataType", func(t *testing.T) {
			page := 1
			limit := "2"
			includeNonActive := true
			byteBody, err := json.Marshal(fiber.Map{
				"page":               page,
				"limit":              limit,
				"include_non_active": includeNonActive,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("GET", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("RequestDataIncomplete", func(t *testing.T) {
			page := 1
			includeNonActive := true
			byteBody, err := json.Marshal(fiber.Map{
				"page": page,
				// "limit":              limit,
				"include_non_active": includeNonActive,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("GET", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			// By default Go will convert the limit into 0 because limit not provided
			// fmt.Println(string(byteResponseBody))
			assert.Contains(t, string(byteResponseBody), "limit could not less than 1")
		})

		t.Cleanup(func() {
			_, _, err := supabaseClient.From(repository.StoreTable).
				Delete("", "").
				In("id", []string{strconv.Itoa(createdTestStores[0].Id), strconv.Itoa(createdTestStores[1].Id)}).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/GetAll (1)")
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		// Setup
		testStoreName := "Test Store Controller SetActivate"
		byteBody, err := json.Marshal(fiber.Map{
			"name": testStoreName,
		})
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))
		request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), requestBody)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		var responseBody struct {
			Code   int    `json:"code"`
			Status string `json:"status"`
			Data   struct {
				CreatedStore *model.Store `json:"created_store"`
			} `json:"data"`
		}
		err = json.Unmarshal(byteResponseBody, &responseBody)
		createdTestStore := responseBody.Data.CreatedStore

		t.Run("NormalSetActivate", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"store_id": createdTestStore.Id,
				"set_into": false,
			})
			require.NoError(t, err)
			requestBody := strings.NewReader(string(byteBody))
			request := httptest.NewRequest("PUT", fmt.Sprint("/stores/set_activate/", createdTestTenant.Id), requestBody)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusAccepted, response.StatusCode)

			var testStore *model.Store
			_, err = supabaseClient.From(repository.StoreTable).
				Select("*", "", false).
				Eq("id", strconv.Itoa(createdTestStore.Id)).
				Single().
				ExecuteTo(&testStore)
			assert.NoError(t, err)
			assert.False(t, testStore.IsActive)
		})

		t.Cleanup(func() {
			_, _, err := supabaseClient.From(repository.StoreTable).
				Delete("", "").
				Eq("id", strconv.Itoa(createdTestStore.Id)).
				Execute()
			require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl/SetActivate (1)")
		})
	})

	t.Cleanup(func() {
		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("user_id", strconv.Itoa(createdTestUser.Id)).
			Eq("tenant_id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (1)")

		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (2)")

		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("id", strconv.Itoa(createdTestUser.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (3)")
	})
}
