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
	"github.com/stretchr/testify/require"
)

type voucherControllerServiceStub struct {
	createdVoucher  *model.Voucher
	createErr       error
	receivedVoucher *model.Voucher
}

func (stub *voucherControllerServiceStub) CreateVoucher(voucher *model.Voucher) (*model.Voucher, error) {
	stub.receivedVoucher = voucher
	return stub.createdVoucher, stub.createErr
}

func (stub *voucherControllerServiceStub) DeleteVoucher(voucherId, tenantId int) error {
	return nil
}

func (stub *voucherControllerServiceStub) UseVoucher(tenantId int, code string, items []*model.Item) (int, int, int, error) {
	return 0, 0, 0, nil
}

func TestVoucherControllerCreate(t *testing.T) {
	t.Run("UsesTenantFromRouteAndReturnsCreatedVoucher", func(t *testing.T) {
		createdVoucher := &model.Voucher{Id: 12, Code: "SAVE", TenantId: 7}
		serviceStub := &voucherControllerServiceStub{createdVoucher: createdVoucher}
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(serviceStub).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE","tenant_id":999}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		assert.Equal(t, fiber.StatusCreated, response.StatusCode)
		require.NotNil(t, serviceStub.receivedVoucher)
		assert.Equal(t, 7, serviceStub.receivedVoucher.TenantId)

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
		serviceStub := &voucherControllerServiceStub{
			createErr: &service.VoucherServiceError{
				Kind:    service.VoucherServiceErrorInternal,
				Message: "Unable to complete the voucher request. Please try again later.",
			},
		}
		app := fiber.New()
		app.Post("/vouchers/:tenantId", NewVoucherControllerImpl(serviceStub).Create)

		request := httptest.NewRequest(http.MethodPost, "/vouchers/7", strings.NewReader(`{"code":"SAVE"}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		assert.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
		var responseBody struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.NewDecoder(response.Body).Decode(&responseBody))
		assert.Equal(t, "Unable to complete the voucher request. Please try again later.", responseBody.Message)
		assert.NotContains(t, responseBody.Message, "database")
	})
}
