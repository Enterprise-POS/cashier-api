package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type CategoryServiceImpl struct {
	Repository        repository.CategoryRepository
	CategoryNameRegex *regexp.Regexp
}

func NewCategoryServiceImpl(repository repository.CategoryRepository) CategoryService {
	return &CategoryServiceImpl{
		Repository:        repository,
		CategoryNameRegex: regexp.MustCompile(`^[a-zA-Z0-9_ ]{1,15}$`),
	}
}

// Create implements CategoryService.
func (service *CategoryServiceImpl) Create(tenantId int, categoryNames []string) ([]*model.Category, error) {
	// 0 means, usually null but GO does not allow null so instead null will get 0
	if tenantId == 0 {
		return nil, errors.New("Tenant Id is not valid")
	}

	if len(categoryNames) == 0 {
		return nil, errors.New("Please fill at least 1 category")
	}

	var categories []*model.Category
	for _, categoryName := range categoryNames {
		// Check for name. Only allowed up to 15 characters
		if !service.CategoryNameRegex.MatchString(categoryName) {
			return nil, fmt.Errorf("Current category name is not allowed: %s", categoryName)
		}

		// Fill category tenant id manually (required)
		categories = append(categories, &model.Category{CategoryName: categoryName, TenantId: tenantId})
	}

	// If all categories is valid then access repository
	createdCategory, err := service.Repository.Create(tenantId, categories)
	if err != nil {
		// 23505: unique_violation
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		if strings.Contains(err.Error(), "(23505)") {
			return nil, errors.New("Something gone wrong. Duplicate category detected")
		}

		return nil, err
	}

	return createdCategory, nil
}

// Get implements CategoryService.
func (service *CategoryServiceImpl) Get(tenantId int, page int, limit int) ([]*model.Category, int, error) {
	// 0 means, usually null but GO does not allow null so instead null will get 0
	if tenantId < 1 {
		return nil, 0, errors.New("Invalid tenant id")
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// By default SQL will be start from 0 index, if page 1 then page have to subtracted by 1 (page = 0)
	categories, count, err := service.Repository.Get(tenantId, page-1, limit)
	if err != nil {
		return nil, 0, err
	}

	return categories, count, nil
}

// Register implements CategoryService.
func (service *CategoryServiceImpl) Register(tobeRegisters []*model.CategoryMtmWarehouse) error {
	if len(tobeRegisters) == 0 {
		return errors.New("Invalid request. Fill at least 1 category and item to be add")
	}

	// categoryId
	// itemId
	for _, tobeRegister := range tobeRegisters {
		if tobeRegister.ItemId < 1 || tobeRegister.CategoryId < 1 {
			return fmt.Errorf("Required item id or category id is not valid. item id: %d, category id: %d", tobeRegister.ItemId, tobeRegister.CategoryId)
		}
	}

	err := service.Repository.Register(tobeRegisters)
	if err != nil {
		if strings.Contains(err.Error(), "(23505)") {
			return errors.New("Error, Current items with category already added")
		}

		if strings.Contains(err.Error(), "(23503)") {
			return errors.New("Forbidden action ! non exist category id or item id")
		}

		return err
	}

	return nil
}

// Unregister implements CategoryService.
func (service *CategoryServiceImpl) Unregister(toUnregister *model.CategoryMtmWarehouse) error {
	// controller will guaranteed parameter not nil

	if toUnregister.CategoryId < 1 || toUnregister.ItemId < 1 {
		return fmt.Errorf("Invalid category id or item id: category id: %d item id: %d", toUnregister.CategoryId, toUnregister.ItemId)
	}

	// Repository will only take category_id and item_id, other properties will be ignored
	err := service.Repository.Unregister(toUnregister)
	if err != nil {
		return err
	}

	return nil
}

// Update implements CategoryService.
func (service *CategoryServiceImpl) Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error) {
	// only 1 category allowed to change /per request
	if tenantId < 1 || categoryId < 1 {
		return nil, fmt.Errorf("Invalid tenant id or category id: tenant id: %d category id: %d", tenantId, categoryId)
	}

	if !service.CategoryNameRegex.MatchString(tobeChangeCategoryName) {
		return nil, fmt.Errorf("Current category name is not allowed: %s", tobeChangeCategoryName)
	}

	updatedCategory, err := service.Repository.Update(tenantId, categoryId, tobeChangeCategoryName)
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST116)") {
			return nil, fmt.Errorf("Nothing is updated from category_id: %d, tenant_id: %d", categoryId, tenantId)
		}
		return nil, err
	}

	return updatedCategory, nil
}

// GetCategoryWithItems implements CategoryService.
func (service *CategoryServiceImpl) GetCategoryWithItems(tenantId int, page int, limit int, doCount bool) ([]*model.CategoryWithItem, int, error) {
	if tenantId < 1 {
		return nil, 0, fmt.Errorf("Fatal Error, Invalid tenant id, tenant id: %d", tenantId)
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	categoryWithItems, count, err := service.Repository.GetCategoryWithItems(tenantId, page-1, limit, doCount)
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST103)") {
			return nil, 0, errors.New("Requested range not satisfiable")
		}

		return nil, 0, err
	}

	return categoryWithItems, count, nil
}

// GetItemsByCategoryId implements CategoryService.
func (service *CategoryServiceImpl) GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, error) {
	if tenantId < 1 {
		return nil, fmt.Errorf("Fatal Error, Invalid tenant id, tenant id: %d", tenantId)
	}

	if categoryId < 1 {
		return nil, fmt.Errorf("Fatal Error, Invalid category id, category id: %d", categoryId)
	}

	if limit < 1 {
		return nil, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	categoryWithItems, err := service.Repository.GetItemsByCategoryId(tenantId, categoryId, limit, page)
	if err != nil {
		return nil, err
	}

	return categoryWithItems, nil
}

// Delete implements CategoryService.
func (service *CategoryServiceImpl) Delete(category *model.Category) error {
	if category == nil {
		return errors.New("Category should not empty")
	}

	// Tenant Id, category id
	if category.TenantId < 1 {
		return fmt.Errorf("Fatal Error, Invalid tenant id, tenant id: %d", category.TenantId)
	}

	if category.Id < 1 {
		return fmt.Errorf("Fatal Error, Invalid category id, category id: %d", category.Id)
	}

	err := service.Repository.Delete(category)
	if err != nil {
		return err
	}

	return nil
}
