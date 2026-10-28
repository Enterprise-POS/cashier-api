package controller

import (
	"cashier-api/exception"
	"cashier-api/model"
	"cashier-api/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVoucherControllerCreate(t *testing.T) {
	t.Run("UsesTenantFromRouteAndReturnsCreatedVoucher", func(t *testing.T) {
		createdVoucher := &model.Voucher{Id: 12, Code: "SAVE", TenantId: 7}
		// Unix seconds for 2025-01-01T00:00:00Z and 2025-01-02T00:00:00Z.
		startAtTimestamp := int64(1735689600)
		endAtTimestamp := int64(1735776000)
		serviceMock := new(mock.Mock)
		serviceMock.On("CreateVoucher", mock.MatchedBy(func(voucher *model.Voucher) bool {
			return voucher.Code == "SAVE" &&
				voucher.TenantId == 7 &&
				voucher.StartAt != nil && voucher.StartAt.Equal(time.Unix(startAtTimestamp, 0).UTC()) &&
				voucher.EndAt != nil && voucher.EndAt.Equal(time.Unix(endAtTimestamp, 0).UTC())
		})).Return(createdVoucher, nil).Once()
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(service.NewVoucherServiceMock(serviceMock)).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE","tenant_id":999,"start_at":1735689600,"end_at":1735776000}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()
		assert.Equal(t, fiber.StatusCreated, response.StatusCode)
		assert.Equal(t, fiber.StatusCreated, response.StatusCode)
		serviceMock.AssertExpectations(t)

		var responseBody struct {
			Data struct {
				Voucher model.Voucher `json:"voucher"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(response.Body).Decode(&responseBody))
		assert.Equal(t, 12, responseBody.Data.Voucher.Id)
		assert.Equal(t, "SAVE", responseBody.Data.Voucher.Code)
	})

	t.Run("RejectsTimeStringsForStartAndEndAt", func(t *testing.T) {
		serviceMock := new(mock.Mock)
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(service.NewVoucherServiceMock(serviceMock)).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE","start_at":"2025-01-01T00:00:00Z"}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		serviceMock.AssertNotCalled(t, "CreateVoucher", mock.Anything)
	})

	t.Run("ReturnsSafeMessageForDatabaseFailure", func(t *testing.T) {
		serviceMock := new(mock.Mock)
		serviceMock.On("CreateVoucher", mock.AnythingOfType("*model.Voucher")).
			Return(nil, &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorInternal,
				Message: "Unable to complete the voucher request. Please try again later.",
			}).Once()
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(service.NewVoucherServiceMock(serviceMock)).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE"}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		serviceMock.AssertExpectations(t)
		assert.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		var responseBody struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.NewDecoder(response.Body).Decode(&responseBody))
		assert.Equal(t, "Unable to complete the voucher request. Please try again later.", responseBody.Message)
		assert.NotContains(t, responseBody.Message, "database")
	})
}
