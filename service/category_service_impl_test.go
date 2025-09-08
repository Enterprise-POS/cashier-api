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

func TestCategoryServiceImpl(t *testing.T) {
	categoryRepository := repository.NewCategoryRepositoryMock(&mock.Mock{}).(*repository.CategoryRepositoryMock)
	categoryService := NewCategoryServiceImpl(categoryRepository)

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
}
