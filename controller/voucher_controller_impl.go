package controller

import (
	"cashier-api/exception"
	common "cashier-api/helper"
	"cashier-api/model"
	"cashier-api/service"
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
