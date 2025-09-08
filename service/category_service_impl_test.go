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

	t.Run("Get", func(t *testing.T) {
		tenantId := 1
		now := time.Now()

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
			categoryRepository.Mock.On("Get", tenantId, page, limit).Return(expectedCategory, len(expectedCategory), nil)
			categories, count, err := categoryService.Get(tenantId, page, limit)
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
			categories, count, err := categoryService.Get(0, page, limit)
			assert.Error(t, err)
			assert.Equal(t, "Invalid tenant id", err.Error())
			assert.Nil(t, categories)
			assert.Equal(t, 0, count)

			// limit
			limit, page = 0, 1
			categories, count, err = categoryService.Get(tenantId, page, limit)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categories)
			assert.Equal(t, fmt.Sprintf("limit could not less then 1 (limit >= 1). Given limit %d", limit), err.Error())

			// page
			limit, page = 5, 0
			categories, count, err = categoryService.Get(tenantId, page, limit)
			assert.Error(t, err)
			assert.Equal(t, 0, count)
			assert.Nil(t, categories)
			assert.Equal(t, fmt.Sprintf("page could not less then 1 (page >= 1). Given page %d", page), err.Error())
		})
	})

	t.Run("Create", func(t *testing.T) {
		tenantId := 1 // Mock tenant id

		t.Run("NormalCreate", func(t *testing.T) {
			categories := []*model.Category{
				{
					CategoryName: "Fruit",
				},
				{
					CategoryName: "Milk",
				},
				{
					CategoryName: "Italian_Food",
				},
				{
					CategoryName: "3 Best Food",
				},
			}

			// Instead of calling real CategoryRepository, CategoryRepositoryMock will replace real repository
			categoryRepository.Mock.
				// Add required parameter for CategoryRepository.Create ex: service.Repository.Create(tenantId, categories)
				On("Create", tenantId, categories).
				// What CategoryRepository.Create want to return
				Return(categories, nil)

			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.NoError(t, err)
			assert.Equal(t, len(categories), len(createdCategories))
		})

		t.Run("InvalidRequiredParameter", func(t *testing.T) {
			categories := []*model.Category{}

			createdCategories, err := categoryService.Create(0, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
			assert.Equal(t, "Tenant Id is not valid", err.Error())

			categories = []*model.Category{}
			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
			assert.Equal(t, "Please fill at least 1 category", err.Error())
		})

		t.Run("CreateWithInvalidName", func(t *testing.T) {
			// Invalid characters
			categories := []*model.Category{
				{
					CategoryName: "Fruit@",
				},
			}

			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			// Category name more than 15 characters
			categories = []*model.Category{
				{
					CategoryName: "Fruit But With Long Category Name",
				},
			}

			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			// Empty string
			categories = []*model.Category{
				{
					CategoryName: "",
				},
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
			categories := []*model.Category{
				{
					Id:           1,
					CategoryName: "Fruits",
				},
			}

			createdCategories, err := categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			now := time.Now()
			categories = []*model.Category{
				{
					CategoryName: "Fruit But With Long Category Name",
					CreatedAt:    &now,
				},
			}

			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)

			categories = []*model.Category{
				{
					CategoryName: "Fruit But With Long Category Name",
					TenantId:     1,
				},
			}
			createdCategories, err = categoryService.Create(tenantId, categories)
			assert.Error(t, err)
			assert.Nil(t, createdCategories)
		})

		t.Run("DuplicateCategoryName", func(t *testing.T) {
			// Duplicate category name handled by postgreSQL constraints
			categories := []*model.Category{
				{
					CategoryName: "Fruits",
				},
				{
					CategoryName: "Fruits",
				},
			}

			categoryRepository.Mock = &mock.Mock{}
			categoryRepository.Mock.On("Create", tenantId, categories).Return(nil, errors.New("(23505)"))
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

			// id specified
			tobeRegisters = []*model.CategoryMtmWarehouse{
				{
					Id:         1, // not allowed
					CategoryId: 1,
					ItemId:     1,
				},
			}
			err = categoryService.Register(tobeRegisters)
			assert.Error(t, err)

			// createdAt specified
			now := time.Now()
			tobeRegisters = []*model.CategoryMtmWarehouse{
				{
					CategoryId: 1,
					ItemId:     1,
					CreatedAt:  &now, // Not allowed
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
}
