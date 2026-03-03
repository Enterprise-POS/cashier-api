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
	gormClient := client.CreateGormClient()

	userRepository := repository.NewUserRepositoryImpl(gormClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)

	app.Use(middleware.ProtectedRoute)

	storeRepository := repository.NewStoreRepositoryImpl(gormClient)
	storeService := service.NewStoreServiceImpl(storeRepository)
	storeController := NewStoreControllerImpl(storeService)

	// User
	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	testUser := &model.UserRegisterForm{
		Name:     "Test_StoreController" + uniqueIdentity,
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
	tenantRestriction := middleware.RestrictByTenant(gormClient) // User only allowed to access associated tenant
	app.Post("/stores/:tenantId", tenantRestriction, storeController.Create)
	app.Get("/stores/:tenantId", tenantRestriction, storeController.GetAll)
	app.Put("/stores/:tenantId", tenantRestriction, storeController.Edit)
	app.Put("/stores/set_activate/:tenantId", tenantRestriction, storeController.SetActivate)

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			testStoreName := "Test Store Controller"
			byteBody, err := json.Marshal(fiber.Map{"name": testStoreName})
			require.NoError(t, err)

			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
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
				result := gormClient.
					Where("id = ? AND tenant_id = ?", responseBody.Data.CreatedStore.Id, createdTestTenant.Id).
					Delete(&model.Store{})
				require.NoError(t, result.Error, "If this fail, then immediately delete the data from TestStoreControllerImpl/Create/NormalCreate")
				require.Equal(t, int64(1), result.RowsAffected)
			})
		})

		t.Run("WrongRequestBodyDataType", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{"name": 1}) // Should be string
			require.NoError(t, err)

			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
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
		testStoreNames := []string{"Test Store Controller Get All1", "Test Store Controller Get All2"}
		var createdTestStores []*model.Store

		for _, storeName := range testStoreNames {
			byteBody, err := json.Marshal(fiber.Map{"name": storeName})
			require.NoError(t, err)

			request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			var responseBody struct {
				Data struct {
					CreatedStore *model.Store `json:"created_store"`
				} `json:"data"`
			}
			err = json.Unmarshal(byteResponseBody, &responseBody)
			require.NoError(t, err)
			createdTestStores = append(createdTestStores, responseBody.Data.CreatedStore)
		}

		t.Run("NormalGetAll", func(t *testing.T) {
			baseURL := fmt.Sprintf("/stores/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			require.NoError(t, err)

			params := url.Values{}
			params.Add("limit", "2")
			params.Add("page", "1")
			params.Add("include_non_active", "true")
			parsedURL.RawQuery = params.Encode()

			request := httptest.NewRequest("GET", parsedURL.String(), nil)
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
					Stores []*model.Store `json:"stores"`
				} `json:"data"`
			}
			err = json.Unmarshal(byteResponseBody, &responseBody)
			require.NoError(t, err)

			for _, store := range responseBody.Data.Stores {
				assert.Contains(t, testStoreNames, store.Name)
				assert.True(t, store.IsActive)
				assert.NotNil(t, store.CreatedAt)
				assert.NotEqual(t, 0, store.Id)
			}
		})

		t.Run("WrongRequestDataType", func(t *testing.T) {
			baseURL := fmt.Sprintf("/stores/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			require.NoError(t, err)

			params := url.Values{}
			params.Add("limit", "impossible number")
			params.Add("page", "1")
			params.Add("include_non_active", "true")
			parsedURL.RawQuery = params.Encode()

			request := httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("RequestDataIncomplete", func(t *testing.T) {
			baseURL := fmt.Sprintf("/stores/%d", createdTestTenant.Id)
			parsedURL, err := url.Parse(baseURL)
			require.NoError(t, err)

			params := url.Values{}
			params.Add("limit", "0")
			params.Add("page", "1")
			params.Add("include_non_active", "true")
			parsedURL.RawQuery = params.Encode()

			request := httptest.NewRequest("GET", parsedURL.String(), nil)
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			byteResponseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(byteResponseBody), "limit could not less than 1")
		})

		t.Cleanup(func() {
			result := gormClient.
				Where("id", []int{createdTestStores[0].Id, createdTestStores[1].Id}).
				Delete(&model.Store{})
			require.NoError(t, result.Error, "If this fail, then immediately delete the data from TestStoreControllerImpl/GetAll")
			require.Equal(t, int64(2), result.RowsAffected)
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		// Setup: create a store via the API
		byteBody, err := json.Marshal(fiber.Map{"name": "Test Store Controller SetActivate"})
		require.NoError(t, err)

		request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		var setupResponseBody struct {
			Data struct {
				CreatedStore *model.Store `json:"created_store"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(byteResponseBody, &setupResponseBody))
		createdTestStore := setupResponseBody.Data.CreatedStore

		t.Run("NormalSetActivate", func(t *testing.T) {
			byteBody, err := json.Marshal(fiber.Map{
				"store_id": createdTestStore.Id,
				"set_into": false,
			})
			require.NoError(t, err)

			request := httptest.NewRequest("PUT", fmt.Sprint("/stores/set_activate/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusAccepted, response.StatusCode)

			// Verify via GORM that is_active is now false
			var testStore model.Store
			err = gormClient.Where("id", createdTestStore.Id).First(&testStore).Error
			assert.NoError(t, err)
			assert.False(t, testStore.IsActive)
		})

		t.Cleanup(func() {
			result := gormClient.Where("id", createdTestStore.Id).Delete(&model.Store{})
			require.NoError(t, result.Error, "If this fail, then immediately delete the data from TestStoreControllerImpl/SetActivate")
		})
	})

	t.Run("Edit", func(t *testing.T) {
		// Setup: create a store via the API
		byteBody, err := json.Marshal(fiber.Map{"name": "Test Store Controller Edit"})
		require.NoError(t, err)

		request := httptest.NewRequest("POST", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(enterprisePOSCookie)
		response, err := app.Test(request, testTimeout)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		var setupResponseBody struct {
			Data struct {
				CreatedStore *model.Store `json:"created_store"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(byteResponseBody, &setupResponseBody))
		createdTestStore := setupResponseBody.Data.CreatedStore

		t.Run("NormalEdit", func(t *testing.T) {
			editedName := "Test Store Controller Normal Edit"
			byteBody, err := json.Marshal(fiber.Map{
				"store_id": createdTestStore.Id,
				"name":     editedName,
			})
			require.NoError(t, err)

			request := httptest.NewRequest("PUT", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, http.StatusOK, response.StatusCode)

			// Verify via GORM that name is updated
			var testStore model.Store
			err = gormClient.Where("id = ?", createdTestStore.Id).First(&testStore).Error
			assert.NoError(t, err)
			assert.Equal(t, editedName, testStore.Name)
		})

		t.Run("WrongInputType", func(t *testing.T) {
			// name should be string, not int
			byteBody, err := json.Marshal(fiber.Map{
				"store_id": createdTestStore.Id,
				"name":     1,
			})
			require.NoError(t, err)

			request := httptest.NewRequest("PUT", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)

			// store_id should be int, not string
			byteBody, err = json.Marshal(fiber.Map{
				"store_id": "1",
				"name":     "Test Store Controller Normal Edit",
			})
			require.NoError(t, err)

			request = httptest.NewRequest("PUT", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err = app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Run("NotSpecifySomeField", func(t *testing.T) {
			// store_id omitted — Go defaults to 0, should return 400
			byteBody, err := json.Marshal(fiber.Map{
				"name": "Test Store Controller Normal Edit",
			})
			require.NoError(t, err)

			request := httptest.NewRequest("PUT", fmt.Sprint("/stores/", createdTestTenant.Id), strings.NewReader(string(byteBody)))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(enterprisePOSCookie)
			response, err := app.Test(request, testTimeout)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		})

		t.Cleanup(func() {
			result := gormClient.Where("id", createdTestStore.Id).Delete(&model.Store{})
			require.NoError(t, result.Error, "If this fail, then immediately delete the data from TestStoreControllerImpl/Edit")
		})
	})

	t.Cleanup(func() {
		err := gormClient.Where("user_id", createdTestUser.Id).Where("tenant_id", createdTestTenant.Id).Delete(&model.UserMtmTenant{}).Error
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (1)")

		err = gormClient.Delete(createdTestTenant).Error
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (2)")

		err = gormClient.Delete(createdTestUser).Error
		require.NoError(t, err, "If this fail, then immediately delete the data from TestStoreControllerImpl (3)")
	})
}
