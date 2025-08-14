package controller

import (
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TenantControllerImpl struct {
	Service service.TenantService
}

func NewTenantControllerImpl(service service.TenantService) TenantController {
	return &TenantControllerImpl{
		Service: service,
	}
}

// GetTenantWithUser implements TenantController.
func (controller *TenantControllerImpl) GetTenantWithUser(ctx *fiber.Ctx) error {
	p_userId := ctx.Params("userId")
	userId, err := strconv.Atoi(p_userId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! params userId is not int"))
	}

	sub := ctx.Locals("sub")
	id, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(401, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	data, err := controller.Service.GetTenantWithUser(userId, id)
	if err != nil {
		if err.Error() == "[TenantService:GetTenantWithUser:1]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Forbidden action detected ! Do not proceed"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by": userId,
			"tenants":      data,
		}))
}

// NewTenant implements TenantController.
func (controller *TenantControllerImpl) NewTenant(ctx *fiber.Ctx) error {
	var body struct {
		Name        string `json:"name"`
		OwnerUserId int    `json:"owner_user_id"`
		// IsActive    bool       `json:"is_active"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	sub := ctx.Locals("sub")
	id, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(401, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	newTenant := &model.Tenant{
		Name:        body.Name,
		OwnerUserId: body.OwnerUserId,
		IsActive:    true, // manually specify because it the first time tenant createc
	}

	err = controller.Service.NewTenant(newTenant, id)
	if err != nil {
		if err.Error() == "[TenantService:NewTenant:1]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, " Forbidden action detected ! Do not proceed"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

// RemoveUserFromTenant implements TenantController.
func (controller *TenantControllerImpl) RemoveUserFromTenant(ctx *fiber.Ctx) error {
	var body struct {
		UserId      int `json:"user_id"` // To be remove user
		TenantId    int `json:"tenant_id"`
		PerformerId int `json:"performer_id"` // Tell the identity who's perform this request/action
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	sub := ctx.Locals("sub")
	userId, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(401, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	toBeRemoveUser := &model.UserMtmTenant{
		UserId:   body.UserId,
		TenantId: body.TenantId,
	}

	response, err := controller.Service.RemoveUserFromTenant(toBeRemoveUser, body.PerformerId, userId)
	if err != nil {
		if err.Error() == "[TenantService:RemoveUserFromTenant]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Forbidden action detected ! Do not proceed"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by": body.PerformerId,
			"message":      strings.Replace(response, "[SUCCESS] ", "", 1),
		}))
}

// AddUserToTenant implements TenantController.
func (controller *TenantControllerImpl) AddUserToTenant(ctx *fiber.Ctx) error {
	var body struct {
		UserId      int `json:"user_id"` // To be add user
		TenantId    int `json:"tenant_id"`
		PerformerId int `json:"performer_id"`
	}

	err := ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! The request body is malformed"))
	}

	sub := ctx.Locals("sub")
	userId, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	createdUserUserMtmTenant, err := controller.Service.AddUserToTenant(body.UserId, body.TenantId, body.PerformerId, userId)
	if err != nil {
		if err.Error() == "[TenantService:AddUserToTenant]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Forbidden action detected ! Do not proceed"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by":     body.PerformerId,
			"added_user_id":    createdUserUserMtmTenant.UserId,
			"target_tenant_id": createdUserUserMtmTenant.TenantId,
		}))
}

// GetTenantMembers implements TenantController.
func (controller *TenantControllerImpl) GetTenantMembers(ctx *fiber.Ctx) error {
	p_tenantId := ctx.Params("tenantId")
	tenantId, err := strconv.Atoi(p_tenantId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! params tenant id is not int"))
	}

	if tenantId == 0 {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Something gone wrong ! Please check for tenant id"))
	}

	sub := ctx.Locals("sub")
	userId, ok := sub.(int)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, "Unexpected behavior ! could not get the id"))
	}

	members, err := controller.Service.GetTenantMembers(tenantId, userId)
	if err != nil {
		if err.Error() == "[TenantService:GetTenantMembers]" {
			return ctx.Status(fiber.StatusForbidden).
				JSON(common.NewWebResponseError(403, common.StatusError, "Forbidden action ! Requesting tenant data that not registered in current requested tenant"))
		}

		return ctx.Status(fiber.StatusBadRequest).
			JSON(common.NewWebResponseError(400, common.StatusError, err.Error()))
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(200, common.StatusSuccess, fiber.Map{
			"requested_by":     userId,
			"requested_tenant": tenantId,
			"members":          members,
		}))
}
