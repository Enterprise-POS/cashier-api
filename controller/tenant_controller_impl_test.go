package controller

import (
	"bytes"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
	private function are at bottom of the file
	- createUser
	- createTenant
*/

func TestTenantControllerImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	supabaseClient := client.CreateSupabaseClient()

	// To create new user, User endpoint are needed
	userRepository := repository.NewUserRepositoryImpl(supabaseClient)
	userService := service.NewUserServiceImpl(userRepository)
	userController := NewUserControllerImpl(userService)

	tenantRepo := repository.NewTenantRepositoryImpl(supabaseClient)
	tenantService := service.NewTenantServiceImpl(tenantRepo)
	tenantController := NewTenantControllerImpl(tenantService)

	app := fiber.New()
	app.Post("/users/sign_up", userController.SignUpWithEmailAndPassword)
	app.Get("/tenants/:userId", middleware.ProtectedRoute, tenantController.GetTenantWithUser)
	app.Get("/tenants/members/:tenantId", middleware.ProtectedRoute, tenantController.GetTenantMembers)
	app.Post("/tenants/new", middleware.ProtectedRoute, tenantController.NewTenant)
	app.Post("/tenants/add_user", middleware.ProtectedRoute, tenantController.AddUserToTenant)
	app.Delete("/tenants/remove_user", middleware.ProtectedRoute, tenantController.RemoveUserFromTenant)

	type RequestBodyStructure struct {
		UserId      int `json:"user_id"` // To be add user
		TenantId    int `json:"tenant_id"`
		PerformerId int `json:"performer_id"`
	}

	t.Run("NewTenant", func(t *testing.T) {
		// Create dummy user
		dummyUser, enterprisePOSCookie := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		t.Run("NormalNewTenant", func(t *testing.T) {
			dummyTenant, response := createTenant(t, app, dummyUser.Name+" Group's", dummyUser.Id, enterprisePOSCookie)

			// The test itself
			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, "Created", responseBody)
			assert.Equal(t, http.StatusCreated, response.StatusCode)

			// Clean up
			// user_mtm_tenant
			_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
				Delete("", "").
				Eq("user_id", fmt.Sprint(dummyUser.Id)).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 1")

			// tenant
			_, _, err = supabaseClient.From(repository.TenantTable).
				Delete("", "").
				Eq("owner_user_id", fmt.Sprint(dummyTenant.OwnerUserId)).
				Eq("name", dummyTenant.Name).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 2")
		})

		t.Run("ForbiddenActionForCreateTenantWithOtherUserId", func(t *testing.T) {
			_, response := createTenant(t, app, dummyUser.Name+" Group's", 99, enterprisePOSCookie)

			// The test itself
			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			assert.Contains(t, responseBody, "Forbidden action detected ! Do not proceed")
		})

		// clean up
		// user
		_, _, err := supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", dummyUser.Email).
			Eq("name", dummyUser.Name).
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl")
	})

	t.Run("GetTenantWithUser", func(t *testing.T) {
		// Create dummy user
		dummyUser, enterprisePOSCookie := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		t.Run("NormalGet", func(t *testing.T) {
			// Create 1
			dummyTenant, _ := createTenant(t, app, dummyUser.Name+" Group's", dummyUser.Id, enterprisePOSCookie)

			// Create 2
			dummyTenant2, _ := createTenant(t, app, dummyUser.Name+" Group's 2", dummyUser.Id, enterprisePOSCookie)

			// The test itself
			request := httptest.NewRequest("GET", fmt.Sprintf("/tenants/%d", dummyUser.Id), nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))
			responseBody, err := common.ReadBody(response.Body)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, dummyTenant.Name)
			assert.Contains(t, responseBody, dummyTenant2.Name)

			// user_mtm_tenant
			_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
				Delete("", "").
				Eq("user_id", fmt.Sprint(dummyUser.Id)).
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 1")

			// tenant
			_, _, err = supabaseClient.From(repository.TenantTable).
				Delete("", "").
				Eq("owner_user_id", fmt.Sprint(dummyTenant.OwnerUserId)).
				// Eq("name", dummyTenant.Name). by commenting this line, allow to supabase to delete 2 tenants
				Execute()
			require.NoError(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl_NormalNewTenant 2")
		})

		// clean up
		// user
		_, _, err := supabaseClient.From(repository.UserTable).
			Delete("", "").
			Eq("email", dummyUser.Email).
			Eq("name", dummyUser.Name).
			Execute()

		require.Nil(t, err, "If this failed, then delete data at DB. error at TestTenantControllerImpl")
	})

	t.Run("AddUserToTenant", func(t *testing.T) {
		// Setup
		dummyUser, enterprisePOSCookie := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		// Create dummy tenant and add current dummyUser into this tenant
		dummyTenant, _ := createTenant(t, app, dummyUser.Name+" Group's", dummyUser.Id, enterprisePOSCookie)

		// Manually get created tenant
		var createdTenant *model.Tenant
		_, err := supabaseClient.From(repository.TenantTable).
			Select("*", "", false).
			Eq("name", dummyTenant.Name).
			Eq("owner_user_id", fmt.Sprint(dummyUser.Id)).
			Single().ExecuteTo(&createdTenant)
		require.NoError(t, err)

		// Create 2nd user, without tenant
		// Create dummy user
		dummyUser2, _ := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		t.Run("NormalAdd", func(t *testing.T) {
			// WARN: Current scope body is for current scope
			requestBody, err := json.Marshal(&RequestBodyStructure{
				UserId:      dummyUser2.Id, // to be added user
				TenantId:    createdTenant.Id,
				PerformerId: dummyTenant.OwnerUserId, // owner
			})
			require.NoError(t, err)
			request := httptest.NewRequest("POST", "/tenants/add_user", bytes.NewReader(requestBody))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))

			bodyBytes, err := io.ReadAll(response.Body)
			responseBody := string(bodyBytes)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, "success")
			assert.Contains(t, responseBody, fmt.Sprintf("\"added_user_id\":%d,", dummyUser2.Id))
			assert.Contains(t, responseBody, fmt.Sprintf("\"requested_by\":%d,", dummyTenant.OwnerUserId))
			assert.Contains(t, responseBody, fmt.Sprintf("\"target_tenant_id\":%d", createdTenant.Id))
		})

		t.Run("DuplicateKeyForInsertingTheSameUserId", func(t *testing.T) {
			// WARN: Current scope body is for current scope
			requestBody, err := json.Marshal(&RequestBodyStructure{
				UserId:      dummyUser2.Id, // to be added user
				TenantId:    createdTenant.Id,
				PerformerId: dummyTenant.OwnerUserId, // owner
			})
			require.NoError(t, err)
			request := httptest.NewRequest("POST", "/tenants/add_user", bytes.NewReader(requestBody))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))

			bodyBytes, err := io.ReadAll(response.Body)
			responseBody := string(bodyBytes)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
			assert.Contains(t, responseBody, "error")
			assert.Contains(t, responseBody, "duplicate key value violates unique constraint")
		})

		t.Run("IllegalActionByUsingAnotherUserId", func(t *testing.T) {
			requestBody, err := json.Marshal(&RequestBodyStructure{
				UserId:      dummyUser2.Id,
				TenantId:    createdTenant.Id,
				PerformerId: 99, // using someone id
			})

			request := httptest.NewRequest("POST", "/tenants/add_user", bytes.NewReader(requestBody))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))

			bodyBytes, err := io.ReadAll(response.Body)
			responseBody := string(bodyBytes)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			assert.Contains(t, responseBody, "error")
			assert.Contains(t, responseBody, "Forbidden action detected ! Do not proceed")
		})

		// Clean up
		// user_mtm_tenant
		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then delete data at DB. error at TestTenantControllerImpl/AddUserToTenant")

		// tenant
		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err)

		// user
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			In("id", []string{fmt.Sprint(dummyUser.Id), fmt.Sprint(dummyUser2.Id)}).
			Execute()
	})

	t.Run("RemoveUserFromTenant", func(t *testing.T) {
		// Setup
		// Create dummy user
		dummyUser, enterprisePOSCookie := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		// Create dummy tenant and add current dummyUser into this tenant
		dummyTenant := &model.Tenant{
			Name:        dummyUser.Name + " Group's",
			OwnerUserId: dummyUser.Id,
		}

		bodyBytes, err := json.Marshal(dummyTenant)
		require.NoError(t, err)

		request := httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
		request.AddCookie(enterprisePOSCookie)
		request.Header.Set("Content-Type", "application/json")
		_, err = app.Test(request, int(time.Second*5))
		require.NoError(t, err)

		// Manually get created tenant
		var createdTenant *model.Tenant
		_, err = supabaseClient.From(repository.TenantTable).
			Select("*", "", false).
			Eq("name", dummyTenant.Name).
			Eq("owner_user_id", fmt.Sprint(dummyUser.Id)).
			Single().ExecuteTo(&createdTenant)
		require.NoError(t, err)

		// Create 2nd user, without tenant
		// Create dummy user
		dummyUser2, _ := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		t.Run("NormalRemove", func(t *testing.T) {
			requestBody, err := json.Marshal(&RequestBodyStructure{
				UserId:      dummyUser2.Id,
				TenantId:    createdTenant.Id,
				PerformerId: dummyUser.Id, // using someone id
			})
			require.NoError(t, err)

			request := httptest.NewRequest("DELETE", "/tenants/remove_user", bytes.NewReader(requestBody))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))

			require.NoError(t, err)
			bodyBytes, err := io.ReadAll(response.Body)
			responseBody := string(bodyBytes)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, "success")
			assert.Contains(t, responseBody, "Removed from tenant")
		})

		t.Run("RemoveOwner", func(t *testing.T) {
			// Current tenant will be archived
			requestBody, err := json.Marshal(&RequestBodyStructure{
				UserId:      dummyUser.Id, // the owner
				TenantId:    createdTenant.Id,
				PerformerId: dummyUser.Id, // using someone id
			})
			require.NoError(t, err)

			// Test itself
			// WARN: even we say delete and it's from the owner, this perform soft delete
			request := httptest.NewRequest("DELETE", "/tenants/remove_user", bytes.NewReader(requestBody))
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, int(time.Second*5))

			require.NoError(t, err)
			bodyBytes, err := io.ReadAll(response.Body)
			responseBody := string(bodyBytes)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, "success")
			assert.Contains(t, responseBody, "Current tenant will be archived")
		})

		// Clean up
		// user_mtm_tenant
		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then delete data at DB. error at TestTenantControllerImpl/AddUserToTenant")

		// tenant
		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err)

		// user
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			In("id", []string{fmt.Sprint(dummyUser.Id), fmt.Sprint(dummyUser2.Id)}).
			Execute()
	})

	t.Run("GetTenantMembers", func(t *testing.T) {
		// Setup
		// User
		// Create 1st user by calling user controller
		dummyUser, enterprisePOSCookie := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		// Create 2nd user, without tenant
		dummyUser2, enterprisePOSCookie2 := createUser(t, app, strings.ReplaceAll(uuid.NewString(), "-", ""), "12345678")

		// Tenant
		// Create tenant and add current dummyUser into this tenant
		dummyTenant, response := createTenant(t, app, dummyUser.Name+" Group's", dummyUser.Id, enterprisePOSCookie)

		// Manually get created tenant
		var createdTenant *model.Tenant
		_, err := supabaseClient.From(repository.TenantTable).
			Select("*", "", false).
			Eq("name", dummyTenant.Name).
			Eq("owner_user_id", fmt.Sprint(dummyUser.Id)).
			Single().ExecuteTo(&createdTenant)
		require.NoError(t, err)

		// Interaction
		// Invite 2nd user into tenant
		type addUserRequestBody struct {
			UserId      int `json:"user_id"` // To be add user
			TenantId    int `json:"tenant_id"`
			PerformerId int `json:"performer_id"`
		}
		bodyBytes, err := json.Marshal(&addUserRequestBody{UserId: dummyUser2.Id, TenantId: createdTenant.Id, PerformerId: dummyUser.Id})
		request := httptest.NewRequest("POST", "/tenants/add_user", bytes.NewReader(bodyBytes))
		request.AddCookie(enterprisePOSCookie)
		request.Header.Set("Content-Type", "application/json")
		response, err = app.Test(request, int(time.Second*5))
		require.NoError(t, err)

		// here is the test
		t.Run("NormalGetTenantMembers", func(t *testing.T) {
			request = httptest.NewRequest("GET", fmt.Sprintf("/tenants/members/%d", createdTenant.Id), nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))
			require.NoError(t, err)

			// The test itself
			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, strconv.Itoa(dummyUser.Id))
			assert.Contains(t, responseBody, strconv.Itoa(dummyUser2.Id))
		})

		t.Run("NonOwnerRequestingGetTenantMembers", func(t *testing.T) {
			request = httptest.NewRequest("GET", fmt.Sprintf("/tenants/members/%d", createdTenant.Id), nil)
			request.AddCookie(enterprisePOSCookie2) // The cookie is from user 2
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))
			require.NoError(t, err)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Contains(t, responseBody, strconv.Itoa(dummyUser.Id))
			assert.Contains(t, responseBody, strconv.Itoa(dummyUser2.Id))
		})

		t.Run("ForbiddenActionByRequestingNotRegisteredInCurrentTenant", func(t *testing.T) {
			request = httptest.NewRequest("GET", fmt.Sprintf("/tenants/members/%d", 1), nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))
			require.NoError(t, err)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, response.StatusCode)
			assert.Contains(t, responseBody, "Forbidden action !")
		})

		t.Run("BadRequestBecauseParamsNotInt", func(t *testing.T) {
			request = httptest.NewRequest("GET", fmt.Sprintf("/tenants/members/%s", "something"), nil)
			request.AddCookie(enterprisePOSCookie)
			request.Header.Set("Content-Type", "application/json")
			response, err = app.Test(request, int(time.Second*5))
			require.NoError(t, err)

			responseBody, err := common.ReadBody(response.Body)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
			assert.Contains(t, responseBody, "Something gone wrong !")
		})

		// Clean up
		// user_mtm_tenant
		_, _, err = supabaseClient.From(repository.UserMtmTenantTable).
			Delete("", "").
			Eq("tenant_id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err, "If this fail, then delete data at DB. error at TestTenantControllerImpl/AddUserToTenant")

		// tenant
		_, _, err = supabaseClient.From(repository.TenantTable).
			Delete("", "").
			Eq("id", fmt.Sprint(createdTenant.Id)).
			Execute()
		require.NoError(t, err)

		// user
		_, _, err = supabaseClient.From(repository.UserTable).
			Delete("", "").
			In("id", []string{fmt.Sprint(dummyUser.Id), fmt.Sprint(dummyUser2.Id)}).
			Execute()
		require.NoError(t, err)
	})
}

func createUser(t *testing.T, app *fiber.App, name string, pass string) (*model.User, *http.Cookie) {
	// Create dummy user
	userIdentity := name
	userPassword := pass
	dummyUser := &model.User{
		Email: "test_tenant_controller_" + userIdentity + "@gmail.com",
		Name:  "test tenant" + userIdentity,
	}
	body := strings.NewReader(fmt.Sprintf(`{
				"email": "%s",
				"password": "%s",
				"name": "%s"
			}`, dummyUser.Email, userPassword, dummyUser.Name))

	signUpRequest := &common.NewRequest{
		Url:       "/users/sign_up",
		Method:    "POST",
		Body:      body,
		TimeoutMs: 0,
	}
	_, response, err := signUpRequest.RunRequest(app)

	require.Nil(t, err)
	require.Equal(t, 201, response.StatusCode)

	var enterprisePOSCookie *http.Cookie
	for _, c := range response.Cookies() {
		if c.Name == constant.EnterprisePOS {
			enterprisePOSCookie = c
		}
	}
	require.NotNil(t, enterprisePOSCookie)

	// Manually get the jwt payload from cookie
	claims := jwt.MapClaims{}
	token, err := common.ClaimJWT(enterprisePOSCookie.Value, &claims)
	require.NotEqual(t, "", claims["sub"])
	sub, ok := claims["sub"].(float64)
	require.True(t, ok)
	require.True(t, token.Valid)
	dummyUser.Id = int(sub)

	return dummyUser, enterprisePOSCookie
}

func createTenant(t *testing.T, app *fiber.App, name string, ownerUserId int, cookie *http.Cookie) (*model.Tenant, *http.Response) {
	dummyTenant := &model.Tenant{
		Name:        name,
		OwnerUserId: ownerUserId,
	}

	bodyBytes, err := json.Marshal(dummyTenant)
	require.NoError(t, err)

	request := httptest.NewRequest("POST", "/tenants/new", bytes.NewReader(bodyBytes))
	request.AddCookie(cookie)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, int(time.Second*5))

	return dummyTenant, response
}
