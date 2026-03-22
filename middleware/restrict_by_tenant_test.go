package middleware

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestrictByTenant(t *testing.T) {
	gormClient := client.CreateGormClient()
	testTimeout := int((time.Second * 5).Milliseconds())

	dummyUser := &model.User{
		Name:  "Test_RestrictByTenant_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		Email: strings.ReplaceAll(uuid.NewString(), "-", "") + "@test.com",
	}
	require.NoError(t, gormClient.Create(dummyUser).Error)

	dummyTenant := &model.Tenant{
		Name:        "Test_RestrictByTenant_Tenant_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		OwnerUserId: dummyUser.Id,
		IsActive:    true,
	}
	require.NoError(t, gormClient.Create(dummyTenant).Error)

	require.NoError(t, gormClient.Create(&model.UserMtmTenant{
		UserId:   dummyUser.Id,
		TenantId: dummyTenant.Id,
	}).Error)

	// Unrelated tenant — user has no association with this one
	otherTenant := &model.Tenant{
		Name:        "Test_RestrictByTenant_Other_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		OwnerUserId: dummyUser.Id,
		IsActive:    true,
	}
	require.NoError(t, gormClient.Create(otherTenant).Error)

	// Helper: builds app that simulates JWT middleware by injecting dummyUser.Id into ctx.Locals("sub")
	newApp := func() *fiber.App {
		app := fiber.New()
		app.Use(func(ctx *fiber.Ctx) error {
			ctx.Locals("sub", dummyUser.Id)
			return ctx.Next()
		})
		app.Get("/test/:tenantId", RestrictByTenant(gormClient), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(fiber.StatusOK)
		})
		return app
	}

	t.Run("AuthorizedUser", func(t *testing.T) {
		// User is associated with dummyTenant — should pass through to 200
		req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", dummyTenant.Id), nil)
		resp, err := newApp().Test(req, testTimeout)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UnauthorizedTenant", func(t *testing.T) {
		// User has no association with otherTenant — should return 403.
		// so it matches any row in user_mtm_tenant and always grants access.
		req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", otherTenant.Id), nil)
		resp, err := newApp().Test(req, testTimeout)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("MissingTenantId", func(t *testing.T) {
		// Route has no :tenantId param — Params("tenantId") returns "" — should return 400
		app := fiber.New()
		app.Use(func(ctx *fiber.Ctx) error {
			ctx.Locals("sub", dummyUser.Id)
			return ctx.Next()
		})
		app.Get("/test", RestrictByTenant(gormClient), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req, testTimeout)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NonIntTenantId", func(t *testing.T) {
		// tenantId cannot be parsed as int — should return 400
		req := httptest.NewRequest("GET", "/test/abc", nil)
		resp, err := newApp().Test(req, testTimeout)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidSubLocals", func(t *testing.T) {
		// ctx.Locals("sub") is a string, not int — type assertion fails — should return 400
		app := fiber.New()
		app.Use(func(ctx *fiber.Ctx) error {
			ctx.Locals("sub", "not-an-int")
			return ctx.Next()
		})
		app.Get("/test/:tenantId", RestrictByTenant(gormClient), func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", fmt.Sprintf("/test/%d", dummyTenant.Id), nil)
		resp, err := app.Test(req, testTimeout)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Cleanup(func() {
		gormClient.Where("user_id", dummyUser.Id).Where("tenant_id", dummyTenant.Id).
			Delete(&model.UserMtmTenant{})
		gormClient.Where("id", []int{dummyTenant.Id, otherTenant.Id}).
			Delete(&model.Tenant{})
		gormClient.Where("id", dummyUser.Id).
			Delete(&model.User{})
	})
}
