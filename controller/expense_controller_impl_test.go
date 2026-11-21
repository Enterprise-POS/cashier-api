package controller

import (
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
)

func TestExpenseControllerImpl(t *testing.T) {
	if os.Getenv(constant.JWT_S) == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	// Whole app runs on 1 transaction and rolls back, nothing is left at DB.
	// Requests are sequential (app.Test), so sharing the transaction is safe here
	tx := client.CreateGormClient().Begin()
	t.Cleanup(func() { tx.Rollback() })

	testTimeout := int((time.Second * 10).Milliseconds())
	app := fiber.New()

	userController := NewUserControllerImpl(service.NewUserServiceImpl(repository.NewUserRepositoryImpl(tx)))
	app.Post("/users/sign_in", userController.SignInWithEmailAndPassword)
	app.Use(middleware.ProtectedRoute)

	expenseController := NewExpenseControllerImpl(service.NewExpenseServiceImpl(repository.NewExpenseRepositoryImpl(tx)))
	tenantRestriction := middleware.RestrictByTenant(tx)
	app.Get("/expenses/details/:tenantId", tenantRestriction, expenseController.FindById)
	app.Post("/expenses/search/:tenantId", tenantRestriction, expenseController.Get)
	app.Post("/expenses/:tenantId", tenantRestriction, expenseController.Create)
	app.Put("/expenses/:tenantId", tenantRestriction, expenseController.Edit)
	app.Delete("/expenses/:tenantId", tenantRestriction, expenseController.Delete)
	app.Get("/expense_categories/:tenantId", tenantRestriction, expenseController.GetCategories)
	app.Post("/expense_categories/:tenantId", tenantRestriction, expenseController.CreateCategory)
	app.Put("/expense_categories/:tenantId", tenantRestriction, expenseController.EditCategory)
	app.Delete("/expense_categories/:tenantId", tenantRestriction, expenseController.DeleteCategory)

	// Member user + tenant + store, and another tenant the user is NOT a member of
	uniqueIdentity := strings.ReplaceAll(uuid.NewString(), "-", "")
	testUser := createUser(tx, &model.UserRegisterForm{
		Name:     "Test_ExpenseController" + uniqueIdentity,
		Email:    uniqueIdentity + "@gmail.com",
		Password: "$2a$10$V6ZP0rm./adZ9kryl3mYf.MB9IY80Y8ZCjtKslUEPWoH.9PCsX7vK", // 12345678
	})
	testTenant := createTenant(tx, &model.Tenant{Name: testUser.Name + "'Group", OwnerUserId: testUser.Id, IsActive: true})
	testStore := createStore(tx, testTenant.Id, "Expense Controller Store")

	otherUser := createUser(tx, &model.UserRegisterForm{Name: "Other" + uniqueIdentity, Email: "other" + uniqueIdentity + "@gmail.com", Password: "x"})
	otherTenant := createTenant(tx, &model.Tenant{Name: otherUser.Name + "'Group", OwnerUserId: otherUser.Id, IsActive: true})

	byteBody, err := json.Marshal(fiber.Map{"email": testUser.Email, "password": "12345678"})
	require.NoError(t, err)
	request := httptest.NewRequest("POST", "/users/sign_in", strings.NewReader(string(byteBody)))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request, testTimeout)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	cookie := extractEnterprisePOSCookie(response.Cookies())
	require.NotNil(t, cookie)

	// send returns status code + decoded "data"
	send := func(t *testing.T, method string, url string, body any) (int, map[string]any) {
		t.Helper()
		var reader io.Reader
		if body != nil {
			byteBody, err := json.Marshal(body)
			require.NoError(t, err)
			reader = strings.NewReader(string(byteBody))
		}
		request := httptest.NewRequest(method, url, reader)
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(cookie)
		response, err := app.Test(request, testTimeout)
		require.NoError(t, err)

		var decoded struct {
			Data map[string]any `json:"data"`
		}
		byteResponse, _ := io.ReadAll(response.Body)
		_ = json.Unmarshal(byteResponse, &decoded)
		return response.StatusCode, decoded.Data
	}
	expenseUrl := fmt.Sprint("/expenses/", testTenant.Id)
	categoryUrl := fmt.Sprint("/expense_categories/", testTenant.Id)

	var categoryId, storeExpenseId int

	t.Run("CreateCategory", func(t *testing.T) {
		status, data := send(t, "POST", categoryUrl, fiber.Map{"name": "Electricity"})
		require.Equal(t, http.StatusOK, status)
		categoryId = int(data["created_expense_category"].(map[string]any)["id"].(float64))

		status, _ = send(t, "POST", categoryUrl, fiber.Map{"name": ""})
		assert.Equal(t, http.StatusBadRequest, status)

		status, data = send(t, "GET", categoryUrl, nil)
		require.Equal(t, http.StatusOK, status)
		assert.Len(t, data["expense_categories"], 1)
	})

	t.Run("Create", func(t *testing.T) {
		status, data := send(t, "POST", expenseUrl, fiber.Map{
			"store_id": testStore.Id, "expense_category_id": categoryId,
			"amount": 150_000, "description": "Electricity November", "expense_date": 1763683200, "payment_type": "QRIS",
		})
		require.Equal(t, http.StatusOK, status)
		created := data["created_expense"].(map[string]any)
		storeExpenseId = int(created["id"].(float64))
		assert.Equal(t, "Electricity", created["expense_category_name"])
		assert.Equal(t, testUser.Name, created["responsible_user_name_snapshot"])
		assert.Equal(t, float64(testUser.Id), created["responsible_user_id"])

		// Whole business, defaults
		status, data = send(t, "POST", expenseUrl, fiber.Map{"amount": 20_000})
		require.Equal(t, http.StatusOK, status)
		created = data["created_expense"].(map[string]any)
		assert.Nil(t, created["store_id"])
		assert.Equal(t, "CASH", created["payment_type"])

		status, _ = send(t, "POST", expenseUrl, fiber.Map{"amount": 0})
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("Search", func(t *testing.T) {
		status, data := send(t, "POST", fmt.Sprint("/expenses/search/", testTenant.Id),
			fiber.Map{"limit": 10, "page": 1})
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, float64(2), data["total_count"])
		assert.Equal(t, float64(150_000+20_000), data["sum_amount"])

		status, data = send(t, "POST", fmt.Sprint("/expenses/search/", testTenant.Id),
			fiber.Map{"limit": 10, "page": 1, "expense_category_id": -1})
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, float64(1), data["total_count"])

		status, _ = send(t, "POST", fmt.Sprint("/expenses/search/", testTenant.Id),
			fiber.Map{"limit": 10, "page": 1, "filters": []fiber.Map{{"column": "tenant_id", "ascending": true}}})
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("EditAndDetails", func(t *testing.T) {
		status, data := send(t, "PUT", expenseUrl, fiber.Map{"expense_id": storeExpenseId, "amount": 175_000, "store_id": nil})
		require.Equal(t, http.StatusOK, status)
		edited := data["edited_expense"].(map[string]any)
		assert.Equal(t, float64(175_000), edited["amount"])
		assert.Nil(t, edited["store_id"])
		assert.Nil(t, edited["expense_category_id"]) // free edit: omitted means cleared

		status, data = send(t, "GET", fmt.Sprintf("/expenses/details/%d?expense_id=%d", testTenant.Id, storeExpenseId), nil)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, float64(175_000), data["expense"].(map[string]any)["amount"])
	})

	t.Run("OtherTenantForbidden", func(t *testing.T) {
		status, _ := send(t, "POST", fmt.Sprint("/expenses/search/", otherTenant.Id), fiber.Map{"limit": 10, "page": 1})
		assert.Equal(t, http.StatusForbidden, status)
		status, _ = send(t, "POST", fmt.Sprint("/expenses/", otherTenant.Id), fiber.Map{"amount": 1})
		assert.Equal(t, http.StatusForbidden, status)
	})

	t.Run("Delete", func(t *testing.T) {
		status, _ := send(t, "DELETE", expenseUrl, fiber.Map{"expense_id": storeExpenseId})
		assert.Equal(t, http.StatusNoContent, status)

		status, _ = send(t, "GET", fmt.Sprintf("/expenses/details/%d?expense_id=%d", testTenant.Id, storeExpenseId), nil)
		assert.Equal(t, http.StatusBadRequest, status)

		status, _ = send(t, "DELETE", categoryUrl, fiber.Map{"expense_category_id": categoryId})
		assert.Equal(t, http.StatusNoContent, status)
	})
}
