package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWarehouseServiceImpl(t *testing.T) {

	var warehouseRepo = &repository.WarehouseRepositoryMock{Mock: mock.Mock{}}
	var warehouseService = NewWarehouseServiceImpl(warehouseRepo)

	t.Run("Get", func(t *testing.T) {
		now := time.Now()

		t.Run("NormalGet", func(t *testing.T) {
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

			result, count, err := warehouseService.GetWarehouseItems(tenantId, limit, page, "")

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

			result, count, err = warehouseService.GetWarehouseItems(0, 5, 1, "")
			assert.NotNil(t, err)
			assert.Nil(t, result)
			assert.Equal(t, 0, count)
			assert.Equal(t, errMessage, err.Error())
		})
	})

	t.Run("GetActiveItem", func(t *testing.T) {
		now := time.Now()

		t.Run("Normal", func(t *testing.T) {
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
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("Get", tenantId, limit, page-1).Return(itemDummies, 5, nil)

			result, count, err := warehouseService.GetWarehouseItems(tenantId, limit, page, "")

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

			result, count, err = warehouseService.GetWarehouseItems(0, 5, 1, "")
			assert.NotNil(t, err)
			assert.Nil(t, result)
			assert.Equal(t, 0, count)
			assert.Equal(t, errMessage, err.Error())
		})
	})

	t.Run("CreateItem", func(t *testing.T) {
		now := time.Now()

		t.Run("NormalCreate", func(t *testing.T) {
			parameterItemDummies := []*model.Item{
				{
					ItemName: "Jasmine",
					Stocks:   10,
					TenantId: 1,
				},
				{
					ItemName: "O'neil",
					Stocks:   10,
					TenantId: 1,
				},
				{
					ItemName: "puffer fish",
					Stocks:   10,
					TenantId: 1,
				},
			}
			expectedItemDummies := []*model.Item{
				{
					ItemId:    1,
					ItemName:  "Jasmine",
					Stocks:    10,
					TenantId:  1,
					IsActive:  true,
					CreatedAt: &now,
				},
				{
					ItemId:    2,
					ItemName:  "O'neil",
					Stocks:    10,
					TenantId:  1,
					IsActive:  true,
					CreatedAt: &now,
				},
				{
					ItemId:    3,
					ItemName:  "puffer fish",
					Stocks:    10,
					TenantId:  1,
					IsActive:  true,
					CreatedAt: &now,
				},
			}

			warehouseRepo.Mock.On("CreateItem", parameterItemDummies).Return(expectedItemDummies, nil)
			items, err := warehouseService.CreateItem(parameterItemDummies)
			assert.NoError(t, err)
			assert.NotNil(t, items)
			assert.Equal(t, len(parameterItemDummies), len(items))
		})

		t.Run("NotAllowedItemName", func(t *testing.T) {
			notAllowedItemName := []*model.Item{
				{
					ItemName: "123Tea",
					Stocks:   10,
					TenantId: 1,
				},
			}

			// Re apply .Mock, otherwise the return will be collided with before Mock
			items, err := warehouseService.CreateItem(notAllowedItemName)
			assert.Error(t, err)
			assert.Nil(t, items)

			notAllowedItemName = []*model.Item{
				{
					ItemName: "'Tea",
					Stocks:   10,
					TenantId: 1,
				},
			}

			items, err = warehouseService.CreateItem(notAllowedItemName)
			assert.Error(t, err)
			assert.Nil(t, items)

			notAllowedItemName = []*model.Item{
				{
					ItemName: "Tea!",
					Stocks:   10,
					TenantId: 1,
				},
			}

			items, err = warehouseService.CreateItem(notAllowedItemName)
			assert.Error(t, err)
			assert.Nil(t, items)

			notAllowedItemName = []*model.Item{
				{
					ItemName: "",
					Stocks:   10,
					TenantId: 1,
				},
			}

			items, err = warehouseService.CreateItem(notAllowedItemName)
			assert.Error(t, err)
			assert.Nil(t, items)
		})

		t.Run("IllegalInput", func(t *testing.T) {
			illegalInput := []*model.Item{
				{
					ItemId:   1, // Item id not allowed
					ItemName: "Jasmine Tea",
					Stocks:   10,
					TenantId: 1,
				},
			}

			items, err := warehouseService.CreateItem(illegalInput)
			assert.Error(t, err)
			assert.Nil(t, items)
		})

		t.Run("EmptyTenant", func(t *testing.T) {
			illegalInput := []*model.Item{
				{
					ItemName: "Jasmine Tea",
					Stocks:   10,
					// TenantId: 1, // Empty TenantId is not allowed
				},
			}

			items, err := warehouseService.CreateItem(illegalInput)
			assert.Error(t, err)
			assert.Nil(t, items)
		})
	})

	t.Run("FindById", func(t *testing.T) {
		now := time.Now()

		t.Run("NormalFindById", func(t *testing.T) {
			expectedItem := &model.Item{
				ItemId:    1,
				ItemName:  "Test item 1",
				Stocks:    10,
				TenantId:  1,
				IsActive:  true,
				CreatedAt: &now,
			}
			warehouseRepo.Mock.On("FindById", expectedItem.ItemId, expectedItem.TenantId).Return(expectedItem, nil)
			items, err := warehouseService.FindById(expectedItem.ItemId, expectedItem.TenantId)
			assert.NoError(t, err)
			assert.Equal(t, expectedItem.ItemId, items.ItemId)
		})

		t.Run("InCaseTenantIdNotFound", func(t *testing.T) {
			itemId := 1
			tenantId := 1
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("FindById", itemId, tenantId).Return(nil, errors.New("(PGRST116) JSON object requested, multiple (or no) rows returned"))
			items, err := warehouseService.FindById(itemId, tenantId)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), fmt.Sprintf("Item not found for current requested item id. Item Id: %d", itemId))
			assert.Nil(t, items)

			itemId = 1
			tenantId = 0
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("FindById", itemId, tenantId).Return(nil, errors.New("(PGRST116) JSON object requested, multiple (or no) rows returned"))
			items, err = warehouseService.FindById(itemId, tenantId)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), fmt.Sprintf("Item not found for current requested item id. Item Id: %d", itemId))
			assert.Nil(t, items)
		})
	})

	t.Run("Edit", func(t *testing.T) {
		now := time.Now()

		t.Run("NormalEdit", func(t *testing.T) {
			/*
				dummyItem := &model.Item{
					ItemId:    1,
					ItemName:  "Test item 1",
					Stocks:    10,
					TenantId:  1,
					IsActive:  true,
					CreatedAt: &now,
				}
			*/
			editedItem := &model.Item{
				ItemId:    1,
				ItemName:  "Test item 1 edited",
				Stocks:    7,
				TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			warehouseRepo.Mock.On("Edit", -3, editedItem).Return(nil)
			err := warehouseService.Edit(-3, editedItem)
			assert.NoError(t, err)
		})

		t.Run("CurrentItemNotFoundAtWarehouse", func(t *testing.T) {
			editedItem := &model.Item{
				ItemId:    1,
				ItemName:  "Test item 1 edited",
				Stocks:    7,
				TenantId:  999,
				IsActive:  false,
				CreatedAt: &now,
			}
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("Edit", -3, editedItem).Return(errors.New("[ERROR] Fatal error, current item from store never exist at warehouse"))
			err := warehouseService.Edit(-3, editedItem)
			assert.Error(t, err)
			assert.Equal(t, "Fatal error, current item from store never exist at warehouse", err.Error())

			editedItem = &model.Item{
				ItemId:    999,
				ItemName:  "Test item 1 edited",
				Stocks:    7,
				TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("Edit", -3, editedItem).Return(errors.New("[ERROR] Fatal error, current item from store never exist at warehouse"))
			err = warehouseService.Edit(-3, editedItem)
			assert.Error(t, err)
			assert.Equal(t, "Fatal error, current item from store never exist at warehouse", err.Error())
		})

		t.Run("InvalidEditName", func(t *testing.T) {
			editedItem := &model.Item{
				ItemId:    1,
				ItemName:  "Test item 1 (edited)",
				Stocks:    7,
				TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			err := warehouseService.Edit(-3, editedItem)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("Could not use this item name: %s", editedItem.ItemName))
		})

		t.Run("EmptyTenantId", func(t *testing.T) {
			editedItem := &model.Item{
				// ItemId:    1,
				ItemName:  "Test item 1 edited",
				Stocks:    7,
				TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			err := warehouseService.Edit(-3, editedItem)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Item ID could not be empty or filled with 0 quantity / -quantity is not allowed")
		})

		t.Run("EmptyTenantId", func(t *testing.T) {
			editedItem := &model.Item{
				ItemId:   1,
				ItemName: "Test item 1 edited",
				Stocks:   7,
				// TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			err := warehouseService.Edit(-3, editedItem)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Required tenant id is empty or filled with 0 quantity / -quantity is not allowed")
		})

		t.Run("IncreasingOrDecreasingTooMuch", func(t *testing.T) {
			editedItem := &model.Item{
				ItemId:    1,
				ItemName:  "Test item 1 edited",
				Stocks:    7,
				TenantId:  1,
				IsActive:  false,
				CreatedAt: &now,
			}
			err := warehouseService.Edit(-1000, editedItem)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "You can only increase an item's quantity up to 999 or decrease by -999")

			err = warehouseService.Edit(1000, editedItem)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "You can only increase an item's quantity up to 999 or decrease by -999")
		})
	})

	t.Run("SetActivate", func(t *testing.T) {
		t.Run("NormalSetActivate", func(t *testing.T) {
			tenantId := 1
			itemId := 1
			setInto := false

			warehouseRepo.Mock.On("SetActivate", tenantId, itemId, setInto).Return(nil)
			err := warehouseService.SetActivate(tenantId, itemId, setInto)
			assert.NoError(t, err)
		})

		t.Run("InvalidItemId", func(t *testing.T) {
			tenantId := 1
			itemId := 0
			setInto := false
			err := warehouseService.SetActivate(tenantId, itemId, setInto)
			assert.Error(t, err)
		})

		t.Run("InvalidTenantId", func(t *testing.T) {
			tenantId := 0
			itemId := 1
			setInto := false
			err := warehouseService.SetActivate(tenantId, itemId, setInto)
			assert.Error(t, err)
		})

		t.Run("NotExistItemAtWarehouse", func(t *testing.T) {
			itemId := 99
			tenantId := 99
			setInto := false

			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("SetActivate", itemId, tenantId, setInto).Return(errors.New("PGRST116"))
			err := warehouseService.SetActivate(tenantId, itemId, setInto)
			assert.Error(t, err)
		})
	})

	t.Run("FindCompleteById", func(t *testing.T) {
		tenantId := 1
		itemId := 1
		t.Run("NormalFindCompleteById", func(t *testing.T) {

			expectedItem := &model.CategoryWithItem{
				CategoryId:   0,
				ItemId:       itemId,
				CategoryName: "",
				TotalCount:   0,
				Stocks:       10,
				ItemName:     "Test item 1 edited",
			}
			warehouseRepo.Mock.On("FindCompleteById", itemId, tenantId).Return(expectedItem, nil)
			item, err := warehouseService.FindCompleteById(itemId, tenantId)
			assert.NoError(t, err)
			assert.Equal(t, expectedItem.ItemId, item.ItemId)
			assert.Equal(t, expectedItem.ItemName, item.ItemName)
		})

		t.Run("NoDataFound", func(t *testing.T) {
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("FindCompleteById", itemId, tenantId).Return(nil, errors.New("NO_DATA_FOUND"))
			item, err := warehouseService.FindCompleteById(itemId, tenantId)
			assert.Error(t, err)
			assert.Nil(t, item)
			assert.Equal(t, "No data return or non exist data", err.Error())
		})

		t.Run("CardinalityViolation", func(t *testing.T) {
			warehouseRepo.Mock = mock.Mock{}
			warehouseRepo.Mock.On("FindCompleteById", itemId, tenantId).Return(nil, errors.New("CARDINALITY_VIOLATION"))
			item, err := warehouseService.FindCompleteById(itemId, tenantId)
			assert.Error(t, err)
			assert.Nil(t, item)
			assert.Equal(t, "Fatal error ! Current item is not valid, duplicate assigning this item category values may cause this error", err.Error())
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			item, err := warehouseService.FindCompleteById(0, tenantId)
			assert.Error(t, err)
			assert.Nil(t, item)
			assert.Equal(t, "Item ID could not be empty or fill <= 0", err.Error())

			item, err = warehouseService.FindCompleteById(itemId, 0)
			assert.Error(t, err)
			assert.Nil(t, item)
			assert.Equal(t, "Required tenant id is empty or fill <= 0", err.Error())
		})
	})
}
