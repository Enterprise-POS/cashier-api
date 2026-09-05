package controller

import (
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPurchasedItemControllerImpl(t *testing.T) {
	//SETUP// — no database, no auth middleware, no network. Purely controller <-> mocked service.
	testTimeout := int(5000) // 5s, in milliseconds
	app := fiber.New()

	purchasedItemServiceMock := service.NewPurchasedItemServiceMock(&mock.Mock{}).(*service.PurchasedItemServiceMock)
	purchasedItemController := NewPurchasedItemControllerImpl(purchasedItemServiceMock)

	// Route registered directly — the handler only reads ctx.Params("tenantId")
	// and the JSON body, so no auth/tenant-restriction middleware is needed here.
	app.Post("/purchased_items/logs/:tenantId", purchasedItemController.PurchasedItemListLogs)

	const TENANT_ID = 42
	const STORE_ID = 1

	t.Run("NormalGet", func(t *testing.T) {
		expectedLogs := []*model.PurchasedItem{
			{
				Id:                 1,
				ItemId:             10,
				OrderItemId:        1,
				Quantity:           2,
				StorePriceSnapshot: 10_000,
				DiscountAmount:     0,
				TotalAmount:        20_000,
				ItemNameSnapshot:   "Apple Snapshot",
				BasePriceSnapshot:  100,
			},
			{
				Id:                 2,
				ItemId:             20,
				OrderItemId:        1,
				Quantity:           1,
				StorePriceSnapshot: 20_000,
				DiscountAmount:     0,
				TotalAmount:        20_000,
				ItemNameSnapshot:   "Peach Snapshot",
				BasePriceSnapshot:  100,
			},
		}

		startDate := int64(1600000000)
		endDate := int64(1700000000)
		dateFilter := &query.DateFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		reqBody := PurchasedItemListLogsRequest{
			StoreId: STORE_ID,
			ItemIds: []int{10, 20},
			Limit:   20,
			Page:    1,
			Filters: []query.QueryFilter{
				{Column: query.CreatedAtColumn, Ascending: true},
			},
			DateFilter: dateFilter,
		}

		byteBody, err := json.Marshal(&reqBody)
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))

		purchasedItemServiceMock.Mock = &mock.Mock{}
		purchasedItemServiceMock.Mock.On(
			"PurchasedItemListLogs",
			TENANT_ID,
			reqBody.StoreId,
			reqBody.ItemIds,
			reqBody.Limit,
			reqBody.Page,
			reqBody.DateFilter,
			reqBody.Filters,
		).Return(expectedLogs, len(expectedLogs), nil)

		request := httptest.NewRequest("POST", fmt.Sprintf("/purchased_items/logs/%d", TENANT_ID), requestBody)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		var envelope common.WebResponse
		require.NoError(t, json.Unmarshal(byteResponseBody, &envelope))

		dataBytes, err := json.Marshal(envelope.Data) // re-marshal the map back to JSON
		require.NoError(t, err)

		var responseBody PurchasedItemListLogsResponse
		require.NoError(t, json.Unmarshal(dataBytes, &responseBody))

		assert.Equal(t, len(expectedLogs), responseBody.TotalCount)
		assert.Equal(t, reqBody.Limit, responseBody.Limit)
		assert.Equal(t, reqBody.Page, responseBody.Page)
		assert.Equal(t, TENANT_ID, responseBody.RequestedByTenantId)
		assert.Len(t, responseBody.Logs, len(expectedLogs))
		purchasedItemServiceMock.Mock.AssertExpectations(t)
	})

	t.Run("DefaultsAndNilFiltersPassThrough", func(t *testing.T) {
		// No item_ids, no date_filter, no filters — request body omits them entirely.
		reqBody := PurchasedItemListLogsRequest{
			StoreId: STORE_ID,
			Limit:   10,
			Page:    1,
		}

		byteBody, err := json.Marshal(&reqBody)
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))

		purchasedItemServiceMock.Mock = &mock.Mock{}
		purchasedItemServiceMock.Mock.On(
			"PurchasedItemListLogs",
			TENANT_ID,
			reqBody.StoreId,
			[]int(nil),
			reqBody.Limit,
			reqBody.Page,
			(*query.DateFilter)(nil),
			[]query.QueryFilter(nil),
		).Return([]*model.PurchasedItem{}, 0, nil)

		request := httptest.NewRequest("POST", fmt.Sprintf("/purchased_items/logs/%d", TENANT_ID), requestBody)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		purchasedItemServiceMock.Mock.AssertExpectations(t)
	})

	t.Run("RequestWithInvalidBody", func(t *testing.T) {
		errorParams := fiber.Map{
			"store_id": "not_an_int", // Should be int
			"limit":    10,
			"page":     1,
		}

		byteBody, err := json.Marshal(errorParams)
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))

		purchasedItemServiceMock.Mock = &mock.Mock{}

		request := httptest.NewRequest("POST", fmt.Sprintf("/purchased_items/logs/%d", TENANT_ID), requestBody)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(byteResponseBody), "Something gone wrong ! The request body is malformed")
		purchasedItemServiceMock.Mock.AssertExpectations(t) // no expectations set, so nothing should have been called
	})

	t.Run("ServiceReturnsError", func(t *testing.T) {
		reqBody := PurchasedItemListLogsRequest{
			StoreId: -1, // Value itself doesn't matter here — the mock decides the outcome
			Limit:   10,
			Page:    1,
		}

		byteBody, err := json.Marshal(&reqBody)
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))

		purchasedItemServiceMock.Mock = &mock.Mock{}
		purchasedItemServiceMock.Mock.On(
			"PurchasedItemListLogs",
			TENANT_ID,
			reqBody.StoreId,
			[]int(nil),
			reqBody.Limit,
			reqBody.Page,
			(*query.DateFilter)(nil),
			[]query.QueryFilter(nil),
		).Return(nil, 0, errors.New("Given store id value is not allowed. storeId: -1"))

		request := httptest.NewRequest("POST", fmt.Sprintf("/purchased_items/logs/%d", TENANT_ID), requestBody)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)

		byteResponseBody, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(byteResponseBody), "Given store id value is not allowed")
		purchasedItemServiceMock.Mock.AssertExpectations(t)
	})

	t.Run("NonNumericTenantIdInPath", func(t *testing.T) {
		// The controller does `strconv.Atoi(ctx.Params("tenantId"))` and ignores the
		// error, so a non-numeric path segment silently becomes tenantId = 0 and the
		// request still reaches the service. There is no middleware here to reject it
		// first (that's RestrictByTenant's job in the real app, not this controller).
		reqBody := PurchasedItemListLogsRequest{
			StoreId: STORE_ID,
			Limit:   10,
			Page:    1,
		}

		byteBody, err := json.Marshal(&reqBody)
		require.NoError(t, err)
		requestBody := strings.NewReader(string(byteBody))

		purchasedItemServiceMock.Mock = &mock.Mock{}
		purchasedItemServiceMock.Mock.On(
			"PurchasedItemListLogs",
			0, // tenantId defaults to zero-value on parse failure
			reqBody.StoreId,
			[]int(nil),
			reqBody.Limit,
			reqBody.Page,
			(*query.DateFilter)(nil),
			[]query.QueryFilter(nil),
		).Return(nil, 0, errors.New("Tenant id is Required !"))

		request := httptest.NewRequest("POST", "/purchased_items/logs/not_a_number", requestBody)
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request, testTimeout)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		purchasedItemServiceMock.Mock.AssertExpectations(t)
	})
}
