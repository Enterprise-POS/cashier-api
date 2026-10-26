package controller

import (
	"cashier-api/model"
	"cashier-api/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVoucherControllerCreate(t *testing.T) {
	t.Run("UsesTenantFromRouteAndReturnsCreatedVoucher", func(t *testing.T) {
		createdVoucher := &model.Voucher{Id: 12, Code: "SAVE", TenantId: 7}
		serviceMock := new(mock.Mock)
		serviceMock.On("CreateVoucher", mock.MatchedBy(func(voucher *model.Voucher) bool {
			return voucher.Code == "SAVE" && voucher.TenantId == 7
		})).Return(createdVoucher, nil).Once()
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(service.NewVoucherServiceMock(serviceMock)).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE","tenant_id":999}`))
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

	t.Run("ReturnsSafeMessageForDatabaseFailure", func(t *testing.T) {
		serviceMock := new(mock.Mock)
		serviceMock.On("CreateVoucher", mock.AnythingOfType("*model.Voucher")).
			Return(nil, &service.VoucherServiceError{
				Kind:    service.VoucherServiceErrorInternal,
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
