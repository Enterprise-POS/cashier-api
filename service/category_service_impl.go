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
	Repository repository.CategoryRepository
}

func NewCategoryServiceImpl(repository repository.CategoryRepository) CategoryService {
	return &CategoryServiceImpl{
		Repository: repository,
	}
}

// Create implements CategoryService.
func (service *CategoryServiceImpl) Create(tenantId int, categories []*model.Category) ([]*model.Category, error) {
	// 0 means, usually null but GO does not allow null so instead null will get 0
	if tenantId == 0 {
		return nil, errors.New("Tenant Id is not valid")
	}

	if len(categories) == 0 {
		return nil, errors.New("Please fill at least 1 category")
	}

	for _, category := range categories {
		// Check for name. Only allowed up to 15 characters
		var itemNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_ ]{1,15}$`)
		if !itemNameRegex.MatchString(category.CategoryName) {
			return nil, fmt.Errorf("Current category name is not allowed: %s", category.CategoryName)
		}

		// category.id must be 0
		if category.Id != 0 {
			return nil, errors.New("Invalid / not allowed category structure (id)")
		}

		if category.CreatedAt != nil {
			return nil, errors.New("Invalid / not allowed category structure (created at)")
		}

		if category.TenantId != 0 {
			return nil, errors.New("Invalid / not allowed category structure (tenant id)")
		}

		// Fill category tenant id manually (required)
		category.TenantId = tenantId
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

	categories, count, err := service.Repository.Get(tenantId, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return categories, count, err
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

		// User does not allowed to specify id
		if tobeRegister.Id != 0 {
			return fmt.Errorf("Fatal error ! Id should not not be specify. invalid id: %d", tobeRegister.Id)
		}

		if tobeRegister.CreatedAt != nil {
			return fmt.Errorf("Fatal error ! Created at should not not be specify. invalid id: %d", tobeRegister.Id)
		}
	}

	err := service.Repository.Register(tobeRegisters)
	if err != nil {
		if strings.Contains(err.Error(), "(23505)") {
			return fmt.Errorf("Error, Current items with category already added")
		}

		return err
	}

	return nil
}
