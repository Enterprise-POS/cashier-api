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

func TestCategoryServiceImpl(t *testing.T) {
	categoryRepository := repository.NewCategoryRepositoryMock(&mock.Mock{}).(*repository.CategoryRepositoryMock)
	categoryService := NewCategoryServiceImpl(categoryRepository)
	now := time.Now()

	t.Run("Get", func(t *testing.T) {
		tenantId := 1

		t.Run("NormalGet", func(t *testing.T) {
			expectedCategory := []*model.Category{
				{
					Id:           1,
					CategoryName: "Fruit",
					TenantId:     tenantId,
					CreatedAt:    &now,
				},
				{
					Id:           2,
					CategoryName: "Milk",
					TenantId:     tenantId,
					CreatedAt:    &now,
				},
				{
					Id:           3,
					CategoryName: "Italian_Food",
					TenantId:     tenantId,
					CreatedAt:    &now,
				},
			}

			limit, page := 5, 1
			categoryRepository.Mock.On("Get", tenantId, page-1, limit, "").Return(expectedCategory, len(expectedCategory), nil)
			categories, count, err := categoryService.Get(tenantId, page, limit, "")
			assert.NoError(t, err)
			assert.Equal(t, len(expectedCategory), count)
			assert.NotNil(t, categories)
			for i, category := range categories {
				assert.Equal(t, expectedCategory[i].Id, category.Id)
				assert.Equal(t, expectedCategory[i].CategoryName, category.CategoryName)
				assert.Equal(t, expectedCategory[i].TenantId, category.TenantId)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// tenant id
			limit, page := 5, 1
			categories, count, err := categoryService.Get(0, page, limit, "")
			assert.Error(t, err)
			assert.Equal(t, "Invalid tenant id", err.Error())
			assert.Nil(t, categories)
			assert.Equal(t, 0, count)

			// limit
			limit, page = 0, 1
			categories, count, err = categoryService.Get(tenantId, page, limit, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categories)
			assert.Equal(t, fmt.Sprintf("limit could not less then 1 (limit >= 1). Given limit %d", limit), err.Error())

			// page
			limit, page = 5, 0
			categories, count, err = categoryService.Get(tenantId, page, limit, "")
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categories)
			assert.Equal(t, fmt.Sprintf("page could not less then 1 (page >= 1). Given page %d", page), err.Error())
		})
	})

	t.Run("Create", func(t *testing.T) {
		tenantId := 1 // Mock tenant id

		t.Run("NormalCreate", func(t *testing.T) {
			categories := []string{
				"Fruit",
				"Milk",
				"Italian_Food",
				"3 Best Food",
			}

			expectedCategories := []*model.Category{
				{
					CategoryName: "Fruit",
					TenantId:     tenantId,
				},
				{
					CategoryName: "Milk",
					TenantId:     tenantId,
				},
				{
					CategoryName: "Italian_Food",
					TenantId:     tenantId,
				},
				{
					CategoryName: "3 Best Food",
					TenantId:     tenantId,
				},
			}

			// Instead of calling real CategoryRepository, CategoryRepositoryMock will replace real repository
			categoryRepository.Mock.
				// Add required parameter for CategoryRepository.Create ex: service.Repository.Create(tenantId, categories)
				On("Create", tenantId, expectedCategories).
				// What CategoryRepository.Create want to return
				Return(expectedCategories, nil)

			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.NoError(t, err)
			assert.Equal(t, len(categories), len(createdCategories))
		})

		t.Run("InvalidRequiredParameter", func(t *testing.T) {
			categories := []string{}

			createdCategories, err := categoryService.Create(0, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
			assert.Equal(t, "Tenant Id is not valid", err.Error())

			categories = []string{}
			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
			assert.Equal(t, "Please fill at least 1 category", err.Error())
		})

		t.Run("CreateWithInvalidName", func(t *testing.T) {
			// Invalid characters
			categories := []string{
				"Fruit@",
			}

			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			// Category name more than 15 characters
			categories = []string{
				"Fruit But With Long Category Name",
			}

			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			// Empty string
			categories = []string{
				"",
			}

			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
		})

		t.Run("InvalidCategoryBody", func(t *testing.T) {
			// Field that should not be filled
			// Id
			// createdAt
			// tenantId

			categories := []string{"Fruit But With Long Category Name"}
			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
		})

		t.Run("DuplicateCategoryName", func(t *testing.T) {
			// Duplicate category name handled by postgreSQL constraints
			categories := []string{"Fruits", "Fruits"}
			expectedCategories := []*model.Category{
				{
					CategoryName: "Fruits",
					TenantId:     tenantId,
				},
				{
					CategoryName: "Fruits",
					TenantId:     tenantId,
				},
			}

			categoryRepository.Mock = &mock.Mock{}
			categoryRepository.Mock.On("Create", tenantId, expectedCategories).Return(nil, errors.New("(23505)"))
			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
		})
	})

	t.Run("Register", func(t *testing.T) {
		t.Run("NormalRegister", func(t *testing.T) {
			tobeRegisters := []*model.CategoryMtmWarehouse{
				{
					CategoryId: 1,
					ItemId:     1,
				},
				{
					CategoryId: 1,
					ItemId:     2,
				},
				{
					CategoryId: 2,
					ItemId:     1,
				},
			}
			categoryRepository.Mock.On("Register", tobeRegisters).Return(nil)
			err := categoryService.Register(tobeRegisters)
			assert.NoError(t, err)
		})

		t.Run("GiveParameterEmptySlice", func(t *testing.T) {
			err := categoryService.Register([]*model.CategoryMtmWarehouse{})
			assert.Error(t, err)
		})

		t.Run("InvalidRegisterBody", func(t *testing.T) {
			// categoryId
			tobeRegisters := []*model.CategoryMtmWarehouse{
				{
					// CategoryId: 1,
					ItemId: 1,
				},
			}
			err := categoryService.Register(tobeRegisters)
			assert.Error(t, err)

			// itemId
			tobeRegisters = []*model.CategoryMtmWarehouse{
				{
					CategoryId: 1,
					// ItemId:     1,
				},
			}
			err = categoryService.Register(tobeRegisters)
			assert.Error(t, err)
		})

		t.Run("DuplicateItem", func(t *testing.T) {
			tobeRegisters := []*model.CategoryMtmWarehouse{
				{
					CategoryId: 1,
					ItemId:     1,
				},
				{
					CategoryId: 1,
					ItemId:     1,
				},
			}

			categoryRepository.Mock.On("Register", tobeRegisters).Return(errors.New("(23505)"))
			err := categoryService.Register(tobeRegisters)
			assert.Equal(t, "Error, Current items with category already added", err.Error())
		})
	})

	t.Run("Unregister", func(t *testing.T) {
		t.Run("NormalUnregister", func(t *testing.T) {
			toUnregister := &model.CategoryMtmWarehouse{
				CategoryId: 1,
				ItemId:     1,
			}

			categoryRepository.Mock.On("Unregister", toUnregister).Return(nil)
			err := categoryService.Unregister(toUnregister)
			assert.NoError(t, err)
		})

		t.Run("InvalidToUnregisterRequest", func(t *testing.T) {
			toUnregister := &model.CategoryMtmWarehouse{
				// CategoryId: 1,
				ItemId: 1,
			}
			err := categoryService.Unregister(toUnregister)
			assert.Error(t, err)

			toUnregister = &model.CategoryMtmWarehouse{
				CategoryId: 1,
				// ItemId:     1,
			}
			err = categoryService.Unregister(toUnregister)
			assert.Error(t, err)
		})

		t.Run("IllegalActionNoDataDeleted", func(t *testing.T) {
			toUnregister := &model.CategoryMtmWarehouse{
				CategoryId: 1,
				ItemId:     1,
			}

			categoryRepository.Mock = &mock.Mock{}
			categoryRepository.Mock.On("Unregister", toUnregister).Return(fmt.Errorf("Warning ! Handled error, no data deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId))
			err := categoryService.Unregister(toUnregister)
			assert.Equal(t, fmt.Sprintf("Warning ! Handled error, no data deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId), err.Error())
		})
	})

	t.Run("EditItemCategory", func(t *testing.T) {
		tenantId := 1 // Non existence tenant id but valid
		t.Run("NormalEditItemCategory", func(t *testing.T) {
			tobeEditItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 2,
				ItemId:     1,
			}

			categoryRepository.Mock.On("EditItemCategory", tenantId, tobeEditItemCategory).Return(nil)
			err := categoryService.EditItemCategory(tenantId, tobeEditItemCategory)
			assert.NoError(t, err)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			tobeEditItemCategory := &model.CategoryMtmWarehouse{
				CategoryId: 1,
				ItemId:     0,
			}

			err := categoryService.EditItemCategory(tenantId, tobeEditItemCategory)
			assert.Error(t, err)

			tobeEditItemCategory = &model.CategoryMtmWarehouse{
				CategoryId: 0,
				ItemId:     1,
			}
			err = categoryService.EditItemCategory(tenantId, tobeEditItemCategory)
			assert.Error(t, err)
		})
	})

	t.Run("Update", func(t *testing.T) {
		tenantId := 1
		categoryId := 1
		tobeChangeCategoryName := "New Name"

		t.Run("NormalUpdate", func(t *testing.T) {
			expectedUpdatedCategory := &model.Category{
				Id:           categoryId,
				CategoryName: tobeChangeCategoryName,
				TenantId:     tenantId,
				CreatedAt:    &now,
			}

			categoryRepository.Mock.On("Update", tenantId, categoryId, tobeChangeCategoryName).Return(expectedUpdatedCategory, nil)
			updatedCategory, err := categoryService.Update(tenantId, categoryId, tobeChangeCategoryName)
			assert.NoError(t, err)
			assert.Equal(t, expectedUpdatedCategory.CategoryName, updatedCategory.CategoryName)
			assert.Equal(t, expectedUpdatedCategory.TenantId, updatedCategory.TenantId)
			assert.Equal(t, expectedUpdatedCategory.Id, updatedCategory.Id)
		})

		t.Run("InvalidParameterValue", func(t *testing.T) {
			// Invalid tenant id
			invalidTenantId := 0
			updatedCategory, err := categoryService.Update(invalidTenantId, categoryId, tobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)

			invalidTenantId = -1
			updatedCategory, err = categoryService.Update(invalidTenantId, categoryId, tobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)

			// Invalid category id
			invalidCategoryId := 0
			updatedCategory, err = categoryService.Update(tenantId, invalidCategoryId, tobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)

			invalidCategoryId = -1
			updatedCategory, err = categoryService.Update(tenantId, invalidCategoryId, tobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)
			assert.Equal(t, fmt.Sprintf("Invalid tenant id or category id: tenant id: %d category id: %d", tenantId, invalidCategoryId), err.Error())

			// Invalid tobe change category name
			invalidTobeChangeCategoryName := "@ something"
			updatedCategory, err = categoryService.Update(tenantId, invalidCategoryId, invalidTobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)

			invalidTobeChangeCategoryName = "Something that too long" // max 15 characters
			updatedCategory, err = categoryService.Update(tenantId, categoryId, invalidTobeChangeCategoryName)
			assert.Error(t, err)
			assert.Nil(t, updatedCategory)
			assert.Equal(t, fmt.Sprintf("Current category name is not allowed: %s", invalidTobeChangeCategoryName), err.Error())
		})
	})

	t.Run("GetCategoryWithItems", func(t *testing.T) {
		tenantId := 1
		limit, page := 5, 1

		t.Run("NormalGetCategoryWithItems", func(t *testing.T) {
			expectedCategoryWithItems := []*model.CategoryWithItem{
				{
					CategoryId:   1,
					CategoryName: "Fruits",
					ItemId:       1,
					ItemName:     "Apple",
					Stocks:       10,
				},
				{
					CategoryId:   1,
					CategoryName: "Fruits",
					ItemId:       2,
					ItemName:     "Banana",
					Stocks:       10,
				},
			}

			categoryRepository.Mock.On("GetCategoryWithItems", tenantId, page-1, limit).Return(expectedCategoryWithItems, len(expectedCategoryWithItems), nil)
			categoryWithItems, count, err := categoryService.GetCategoryWithItems(tenantId, page, limit)
			assert.NoError(t, err)
			assert.Equal(t, count, len(categoryWithItems))
			assert.NotNil(t, categoryWithItems)
			for i, categoryWithItem := range categoryWithItems {
				assert.Equal(t, expectedCategoryWithItems[i].CategoryId, categoryWithItem.CategoryId)
				assert.Equal(t, expectedCategoryWithItems[i].CategoryName, categoryWithItem.CategoryName)
				assert.Equal(t, expectedCategoryWithItems[i].ItemId, categoryWithItem.ItemId)
				assert.Equal(t, expectedCategoryWithItems[i].ItemName, categoryWithItem.ItemName)
				assert.Equal(t, expectedCategoryWithItems[i].Stocks, categoryWithItem.Stocks)
			}
		})

		t.Run("ReturnNothing", func(t *testing.T) {
			notExistPage := 999
			categoryRepository.Mock = &mock.Mock{}
			categoryRepository.Mock.On("GetCategoryWithItems", tenantId, notExistPage-1, limit).Return(nil, 0, errors.New("(PGRST103)"))
			categoryWithItems, count, err := categoryService.GetCategoryWithItems(tenantId, notExistPage, limit)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categoryWithItems)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// tenant id
			categoryWithItems, count, err := categoryService.GetCategoryWithItems(0, page, limit)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categoryWithItems)

			// limit
			categoryWithItems, count, err = categoryService.GetCategoryWithItems(tenantId, page, 0)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categoryWithItems)

			// page
			categoryWithItems, count, err = categoryService.GetCategoryWithItems(tenantId, 0, limit)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categoryWithItems)
		})
	})

	t.Run("GetItemsByCategoryId", func(t *testing.T) {
		tenantId := 1
		categoryId := 1
		limit, page := 5, 1

		t.Run("NormalGetItemsByCategoryId", func(t *testing.T) {
			// Category id must the same.
			expectedCategoryWithItems := []*model.CategoryWithItem{
				{
					CategoryId:   categoryId,
					CategoryName: "Fruits",
					ItemId:       1,
					ItemName:     "Apple",
					Stocks:       10,
				},
				{
					CategoryId:   categoryId,
					CategoryName: "Fruits",
					ItemId:       2,
					ItemName:     "Banana",
					Stocks:       10,
				},
			}

			categoryRepository.Mock.On("GetItemsByCategoryId", tenantId, categoryId, limit, page-1).Return(expectedCategoryWithItems, len(expectedCategoryWithItems), nil)
			categoryWithItems, count, err := categoryService.GetItemsByCategoryId(tenantId, categoryId, limit, page)
			assert.NoError(t, err)
			assert.Greater(t, count, 0)
			assert.Equal(t, len(expectedCategoryWithItems), len(categoryWithItems))
			assert.NotNil(t, categoryWithItems)
			for i, categoryWithItem := range categoryWithItems {
				assert.Equal(t, expectedCategoryWithItems[i].CategoryId, categoryWithItem.CategoryId)
				assert.Equal(t, expectedCategoryWithItems[i].CategoryName, categoryWithItem.CategoryName)
				assert.Equal(t, expectedCategoryWithItems[i].ItemId, categoryWithItem.ItemId)
				assert.Equal(t, expectedCategoryWithItems[i].ItemName, categoryWithItem.ItemName)
				assert.Equal(t, expectedCategoryWithItems[i].Stocks, categoryWithItem.Stocks)
			}
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// tenantId
			invalidTenantId := 0
			categoryWithItems, count, err := categoryService.GetItemsByCategoryId(invalidTenantId, categoryId, limit, page)
			assert.Error(t, err)
			assert.Nil(t, categoryWithItems)
			assert.Equal(t, 0, count)

			// categoryId
			invalidCategoryId := 0
			categoryWithItems, count, err = categoryService.GetItemsByCategoryId(tenantId, invalidCategoryId, limit, page)
			assert.Error(t, err)
			assert.Nil(t, categoryWithItems)
			assert.Equal(t, 0, count)

			// limit
			invalidLimit := 0
			categoryWithItems, count, err = categoryService.GetItemsByCategoryId(tenantId, categoryId, invalidLimit, page)
			assert.Error(t, err)
			assert.Nil(t, categoryWithItems)
			assert.Equal(t, 0, count)

			// page
			invalidPage := 0
			categoryWithItems, count, err = categoryService.GetItemsByCategoryId(tenantId, categoryId, limit, invalidPage)
			assert.Error(t, err)
			assert.Nil(t, categoryWithItems)
			assert.Equal(t, 0, count)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("NormalDelete", func(t *testing.T) {
			category := &model.Category{
				Id:       1,
				TenantId: 1,

				// Not Required
				CategoryName: "Test Category Name",

				// Not Required
				CreatedAt: &now,
			}

			categoryRepository.Mock.On("Delete", category).Return(nil)
			err := categoryService.Delete(category)
			assert.NoError(t, err)
		})

		t.Run("NoDataDeleted", func(t *testing.T) {
			category := &model.Category{
				Id:           1,
				CategoryName: "Test Category Name",
				CreatedAt:    &now,
				TenantId:     1,
			}

			categoryRepository.Mock = &mock.Mock{}
			categoryRepository.Mock.On("Delete", category).Return(fmt.Errorf("Warning ! Handled error, no data deleted from categoryId: %d, tenantId: %d", category.Id, category.TenantId))
			err := categoryService.Delete(category)
			assert.Error(t, err)
		})

		t.Run("InvalidCategoryBody", func(t *testing.T) {
			invalidCategory := &model.Category{
				// Id: 1, // Invalid
				CategoryName: "Test Category Name",
				CreatedAt:    &now,
				TenantId:     1,
			}
			err := categoryService.Delete(invalidCategory)
			assert.Error(t, err)

			invalidCategory = &model.Category{
				Id:           1, // Invalid
				CategoryName: "Test Category Name",
				CreatedAt:    &now,
				// TenantId:     1,
			}
			err = categoryService.Delete(invalidCategory)
			assert.Error(t, err)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			// Since pointer parameter, nil is allowed by GO
			err := categoryService.Delete(nil)
			assert.Error(t, err)
		})
	})
}
