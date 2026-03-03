package middleware

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"strconv"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

/*
Always put this middleware after protected_route middleware
*/
func RestrictByTenant(client *gorm.DB) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get user id (from JWT middleware -> ctx.Locals("sub"))
		sub := ctx.Locals("sub")
		userId, ok := sub.(int)
		if !ok {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! could not get the id"))
		}

		// Convert tenantId from route param
		paramTenantId := ctx.Params("tenantId")
		if paramTenantId == "" {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "Missing tenant_id at the parameter"))
		}

		// Here we don't need the int tenantId, but if there is no error then the next handler is guaranteed to be int
		// Try to see example for warehouse.CreateItem
		_, err := strconv.Atoi(paramTenantId)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "TenantId is not int"))
		}

		// Check if relation exists in user_mtm_tenant
		// supabase returns error when no rows found
		var exist model.UserMtmTenant
		err = client.Select("user_id, tenant_id").Where("user_id", userId).Where("tenant_id", paramTenantId).Take(&exist).Error

		if err != nil {
			log.Warnf("Forbidden action detected. Current user is not associate with requested tenant. From userId: %d, requesting for tenantId: %s", userId, paramTenantId)
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Access denied to tenant. Current user is not associate with requested tenant."))
		}

		// ✅ Authorized
		return ctx.Next()
	}
}
