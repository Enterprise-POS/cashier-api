package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
)

type WarehouseServiceImpl struct {
	Repository repository.WarehouseRepository
}

func NewWarehouseServiceImpl(repository repository.WarehouseRepository) WarehouseService {
	return &WarehouseServiceImpl{Repository: repository}
}

func (service *WarehouseServiceImpl) GetWarehouseItems(tenantId, limit, page int) ([]*model.Item, int, error) {
	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// By default SQL will be start from 0 index, if page 1 then page have to subtracted by 1 (page = 0)
	return service.Repository.Get(tenantId, limit, page-1)
}

// CreateItem implements WarehouseService.
func (service *WarehouseServiceImpl) CreateItem(items []*model.Item) ([]*model.Item, error) {
	isError := false
	errorString := "Some item have an error, please check the input: \n"

	/*
		Allowed condition
		Jasmine
		Jasmine Tea
		Tea123
		O'Neil
		Green Tea 2
		Apple'
		A B C
		neal
	*/
	itemRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9' ]*$`)
	for _, item := range items {
		if !itemRegex.MatchString(item.ItemName) {
			isError = true
			errorString += fmt.Sprintf("Could not use this item name: %s\n", item.ItemName)
		}

		if item.ItemId != 0 {
			isError = true
			errorString += fmt.Sprintf("Illegal input from item name: %s, with illegal item id: %d", item.ItemName, item.ItemId)
		}

		// Item tenant id should not be null
		if item.TenantId == 0 {
			isError = true
			errorString += fmt.Sprintf("Tenant not provided for item name: %s\n", item.ItemName)
		}

		if item.Stocks < 0 {
			item.Stocks = 0
		}

		item.IsActive = true
	}

	// While scanning, if 1 item get is invalid, then all operation will fail
	if isError {
		return nil, errors.New(errorString)
	}

	createdItem, err := service.Repository.CreateItem(items)
	if err != nil {
		return nil, err
	}

	return createdItem, nil
}
