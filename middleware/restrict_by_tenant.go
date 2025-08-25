package middleware

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"strconv"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

/*
Always put this middleware after protected_route middleware
*/
func RestrictByTenant(client *supabase.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get user id (from JWT middleware -> ctx.Locals("sub"))
		sub := ctx.Locals("sub")
		userId, ok := sub.(int)
		if !ok {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! could not get the id"))
		}

		// Convert tenantId from route param
		tenantId := ctx.Params("tenantId")
		if tenantId == "" {
			return ctx.Status(fiber.StatusBadRequest).
				JSON(common.NewWebResponseError(400, common.StatusError, "Missing tenant_id at the parameter"))
		}

		// Check if relation exists in user_mtm_tenant
		// supabase returns error when no rows found
		var exist model.UserMtmTenant
		_, err := client.From("user_mtm_tenant").
			Select("user_id, tenant_id", "", false).
			Eq("user_id", strconv.Itoa(userId)).
			Eq("tenant_id", tenantId).
			Single().
			ExecuteTo(&exist)

		if err != nil {
			log.Warnf("Forbidden action detected. Current user is not associate with requested tenant. From userId: %d, requesting for tenantId: %s", userId, tenantId)
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Access denied to tenant. Current user is not associate with requested tenant."))
		}

		// ✅ Authorized
		return ctx.Next()
	}
}
