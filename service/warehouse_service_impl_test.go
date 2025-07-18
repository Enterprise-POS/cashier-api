package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWarehouseServiceImpl(t *testing.T) {

	var warehouseRepo = &repository.WarehouseRepositoryMock{Mock: mock.Mock{}}
	var warehouseService = NewWarehouseServiceImpl(warehouseRepo)

	t.Run("NormalGet", func(t *testing.T) {
		now := time.Now()
		itemDummies := []*model.Item{
			{
				ItemId:    1,
				ItemName:  "Test 1",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			},
			{
				ItemId:    2,
				ItemName:  "Test 2",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			},
			{
				ItemId:    3,
				ItemName:  "Test 3",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			},
			{
				ItemId:    4,
				ItemName:  "Test 4",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			},
			{
				ItemId:    5,
				ItemName:  "Test 5",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			},
		}

		// Tell mock to return something
		// for test purpose set page=0
		// because at warehouseService.GetWarehouseItems the page will be auto subtracted
		tenantId, limit, page := 1, 5, 1
		warehouseRepo.Mock.On("Get", tenantId, limit, page-1).Return(itemDummies, 5, nil)

		result, count, err := warehouseService.GetWarehouseItems(tenantId, limit, page)

		// fmt.Println(result, count, err)
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, 0, count)
		assert.Equal(t, 5, count)
		for i, item := range result {
			assert.Equal(t, itemDummies[i].ItemId, item.ItemId)
			assert.Equal(t, itemDummies[i].TenantId, item.TenantId)
			assert.Equal(t, itemDummies[i].ItemName, item.ItemName)
			assert.Equal(t, itemDummies[i].Stocks, item.Stocks)
			assert.Equal(t, itemDummies[i].IsActive, item.IsActive)
		}

		tenantId, limit, page = 0, 5, 1
		errMessage := "(PGRST103) Requested range not satisfiable"
		warehouseRepo.Mock.On("Get", tenantId, limit, page-1).Return(nil, 0, errors.New(errMessage))

		result, count, err = warehouseService.GetWarehouseItems(0, 5, 1)
		assert.NotNil(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, count)
		assert.Equal(t, errMessage, err.Error())
	})
}
