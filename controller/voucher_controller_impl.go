package controller

import (
	"bytes"
	"cashier-api/exception"
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	log "github.com/sirupsen/logrus"
)

type VoucherControllerImpl struct {
	Service service.VoucherService
}

func NewVoucherControllerImpl(voucherService service.VoucherService) VoucherController {
	return &VoucherControllerImpl{Service: voucherService}
}

func (controller *VoucherControllerImpl) Get(ctx fiber.Ctx) error {
	tenantId, err := voucherTenantID(ctx)
	if err != nil {
		return voucherRequestError(ctx, "get", err, "Invalid tenant ID.")
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "5"))
	if err != nil || limit < 1 {
		return voucherRequestError(ctx, "get", errors.New("limit must be greater than zero"), "Invalid limit.")
	}
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return voucherRequestError(ctx, "get", errors.New("page must be greater than zero"), "Invalid page.")
	}

	queryFilters := common.ParseQueryFilterParam(ctx.Query("sort", "created_at:desc"))
	vouchers, count, err := controller.Service.GetVouchers(
		tenantId,
		limit,
		page,
		ctx.Query("name_query", ""),
		queryFilters,
	)
	if err != nil {
		return voucherServiceErrorResponse(ctx, "get", tenantId, err)
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(fiber.StatusOK, common.StatusSuccess, fiber.Map{
			"count":    count,
			"vouchers": vouchers,
		}))
}

func (controller *VoucherControllerImpl) Create(ctx fiber.Ctx) error {
	tenantId, err := voucherTenantID(ctx)
	if err != nil {
		return voucherRequestError(ctx, "create", err, "Invalid tenant ID.")
	}

	var body struct {
		model.Voucher
		// Voucher dates are Unix timestamps in seconds, like query.DateFilter.
		StartAt *int64 `json:"start_at"`
		EndAt   *int64 `json:"end_at"`
	}
	if err := ctx.Bind().Body(&body); err != nil {
		return voucherRequestError(ctx, "create", err, "The request body is malformed.")
	}
	voucher := body.Voucher
	if body.StartAt != nil {
		startAt := time.Unix(*body.StartAt, 0).UTC()
		voucher.StartAt = &startAt
	}
	if body.EndAt != nil {
		endAt := time.Unix(*body.EndAt, 0).UTC()
		voucher.EndAt = &endAt
	}
	voucher.TenantId = tenantId

	createdVoucher, err := controller.Service.CreateVoucher(&voucher)
	if err != nil {
		return voucherServiceErrorResponse(ctx, "create", tenantId, err)
	}

	return ctx.Status(fiber.StatusCreated).
		JSON(common.NewWebResponse(fiber.StatusCreated, common.StatusSuccess, fiber.Map{
			"voucher": createdVoucher,
		}))
}

func (controller *VoucherControllerImpl) Edit(ctx fiber.Ctx) error {
	tenantId, err := voucherTenantID(ctx)
	if err != nil {
		return voucherRequestError(ctx, "edit", err, "Invalid tenant ID.")
	}
	voucherId, err := strconv.Atoi(ctx.Params("voucherId"))
	if err != nil || voucherId < 1 {
		if err == nil {
			err = errors.New("voucher ID must be greater than zero")
		}
		return voucherRequestError(ctx, "edit", err, "Invalid voucher ID.")
	}

	var body map[string]json.RawMessage
	if err := ctx.Bind().Body(&body); err != nil || body == nil {
		if err == nil {
			err = errors.New("request body must be a JSON object")
		}
		return voucherRequestError(ctx, "edit", err, "The request body is malformed.")
	}

	updates := &model.VoucherUpdate{}
	if rawValue, exists := body["name"]; exists {
		var name string
		if err := json.Unmarshal(rawValue, &name); err != nil || bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			return voucherRequestError(ctx, "edit", errors.New("name must be a string"), "Invalid voucher name.")
		}
		updates.Name = &name
	}
	if rawValue, exists := body["code"]; exists {
		var code string
		if err := json.Unmarshal(rawValue, &code); err != nil || bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			return voucherRequestError(ctx, "edit", errors.New("code must be a string"), "Invalid voucher code.")
		}
		updates.Code = &code
	}
	if rawValue, exists := body["start_at"]; exists {
		updates.StartAtSet = true
		if !bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			var timestamp int64
			if err := json.Unmarshal(rawValue, &timestamp); err != nil {
				return voucherRequestError(ctx, "edit", err, "Invalid start_at; expected a Unix timestamp in seconds or null.")
			}
			startAt := time.Unix(timestamp, 0).UTC()
			updates.StartAt = &startAt
		}
	}
	if rawValue, exists := body["end_at"]; exists {
		updates.EndAtSet = true
		if !bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			var timestamp int64
			if err := json.Unmarshal(rawValue, &timestamp); err != nil {
				return voucherRequestError(ctx, "edit", err, "Invalid end_at; expected a Unix timestamp in seconds or null.")
			}
			endAt := time.Unix(timestamp, 0).UTC()
			updates.EndAt = &endAt
		}
	}

	updatedVoucher, err := controller.Service.EditVoucher(voucherId, tenantId, updates)
	if err != nil {
		return voucherServiceErrorResponse(ctx, "edit", tenantId, err)
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(fiber.StatusOK, common.StatusSuccess, fiber.Map{
			"voucher": updatedVoucher,
		}))
}

func (controller *VoucherControllerImpl) Delete(ctx fiber.Ctx) error {
	tenantId, err := voucherTenantID(ctx)
	if err != nil {
		return voucherRequestError(ctx, "delete", err, "Invalid tenant ID.")
	}
	voucherId, err := strconv.Atoi(ctx.Params("voucherId"))
	if err != nil || voucherId < 1 {
		if err == nil {
			err = errors.New("voucher ID must be greater than zero")
		}
		return voucherRequestError(ctx, "delete", err, "Invalid voucher ID.")
	}

	if err := controller.Service.DeleteVoucher(voucherId, tenantId); err != nil {
		return voucherServiceErrorResponse(ctx, "delete", tenantId, err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (controller *VoucherControllerImpl) Use(ctx fiber.Ctx) error {
	tenantId, err := voucherTenantID(ctx)
	if err != nil {
		return voucherRequestError(ctx, "use", err, "Invalid tenant ID.")
	}

	var body struct {
		Code  string        `json:"code"`
		Items []*model.Item `json:"items"`
	}
	if err := ctx.Bind().Body(&body); err != nil {
		return voucherRequestError(ctx, "use", err, "The request body is malformed.")
	}

	beforeDiscountPrice, discountPrice, afterDiscountPrice, err := controller.Service.UseVoucher(tenantId, body.Code, body.Items)
	if err != nil {
		return voucherServiceErrorResponse(ctx, "use", tenantId, err)
	}

	return ctx.Status(fiber.StatusOK).
		JSON(common.NewWebResponse(fiber.StatusOK, common.StatusSuccess, fiber.Map{
			"before_discount_price": beforeDiscountPrice,
			"discount_price":        discountPrice,
			"after_discount_price":  afterDiscountPrice,
		}))
}

func voucherTenantID(ctx fiber.Ctx) (int, error) {
	tenantId, err := strconv.Atoi(ctx.Params("tenantId"))
	if err != nil {
		return 0, err
	}
	if tenantId < 1 {
		return 0, errors.New("tenant ID must be greater than zero")
	}
	return tenantId, nil
}

func voucherRequestError(ctx fiber.Ctx, operation string, err error, message string) error {
	log.WithFields(log.Fields{
		"component": "voucher_controller",
		"operation": operation,
		"tenant_id": ctx.Params("tenantId"),
	}).WithError(err).Warn("Invalid voucher request")

	return ctx.Status(fiber.StatusBadRequest).
		JSON(common.NewWebResponseError(fiber.StatusBadRequest, common.StatusError, message))
}

func voucherServiceErrorResponse(ctx fiber.Ctx, operation string, tenantId int, err error) error {
	statusCode := fiber.StatusBadRequest
	message := err.Error()
	logLevel := log.WarnLevel

	var voucherError *exception.VoucherServiceError
	if errors.As(err, &voucherError) {
		switch voucherError.Kind {
		case exception.VoucherServiceErrorNotFound:
			statusCode = fiber.StatusNotFound
		case exception.VoucherServiceErrorConflict:
			statusCode = fiber.StatusConflict
		case exception.VoucherServiceErrorUnavailable:
			statusCode = fiber.StatusServiceUnavailable
			logLevel = log.ErrorLevel
		case exception.VoucherServiceErrorInternal:
			statusCode = fiber.StatusInternalServerError
			logLevel = log.ErrorLevel
		}
	}

	logger := log.WithFields(log.Fields{
		"component": "voucher_controller",
		"operation": operation,
		"status":    statusCode,
		"tenant_id": tenantId,
	})
	if logLevel == log.ErrorLevel {
		logger.WithError(err).Error("Voucher request failed")
	} else {
		logger.WithError(err).Warn("Voucher request failed")
	}

	return ctx.Status(statusCode).
		JSON(common.NewWebResponseError(statusCode, common.StatusError, message))
}
