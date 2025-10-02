package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type WarehouseServiceImpl struct {
	Repository repository.WarehouseRepository
}

func NewWarehouseServiceImpl(repository repository.WarehouseRepository) WarehouseService {
	return &WarehouseServiceImpl{Repository: repository}
}

func (service *WarehouseServiceImpl) GetWarehouseItems(tenantId, limit, page int, nameQuery string) ([]*model.Item, int, error) {
	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// By default SQL will be start from 0 index, if page 1 then page have to subtracted by 1 (page = 0)
	return service.Repository.Get(tenantId, limit, page-1, nameQuery)
}

// GetActiveItem implements WarehouseService.
func (service *WarehouseServiceImpl) GetActiveItem(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {
	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}

	// By default SQL will be start from 0 index, if page 1 then page have to subtracted by 1 (page = 0)
	return service.Repository.GetActiveItem(tenantId, limit, page-1, nameQuery)
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
	itemRegex := regexp.MustCompile(`^[\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z][\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z0-9' ]*$`)
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

// FindById implements WarehouseService.
func (service *WarehouseServiceImpl) FindById(itemId, tenantId int) (*model.Item, error) {
	// Parameter is impossible to be Nil, so no need to check.
	// If in the case user request invalid itemId then Repository will handle it by return nothing
	requestedItem, err := service.Repository.FindById(itemId, tenantId)
	if err != nil {
		if strings.Contains(err.Error(), "PGRST116") {
			return nil, fmt.Errorf("Item not found for current requested item id. Item Id: %d", itemId)
		}
		return nil, err
	}

	return requestedItem, nil
}

// Edit implements WarehouseService.
func (service *WarehouseServiceImpl) Edit(quantity int, item *model.Item) error {
	// item.Stocks == 0, is allowed

	if item.ItemId < 1 {
		return errors.New("Item ID could not be empty or filled with 0 quantity / -quantity is not allowed")
	}
	if item.TenantId < 1 {
		return errors.New("Required tenant id is empty or filled with 0 quantity / -quantity is not allowed")
	}
	itemRegex := regexp.MustCompile(`^[\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z][\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z0-9' ]*$`)
	if !itemRegex.MatchString(item.ItemName) {
		return fmt.Errorf("Could not use this item name: %s\n", item.ItemName)
	}

	if quantity > 999 || quantity < -999 {
		return errors.New("You can only increase an item's quantity up to 999 or decrease by -999")
	}

	err := service.Repository.Edit(quantity, item)
	if err != nil {
		if err.Error() == "[ERROR] Fatal error, current item from store never exist at warehouse" {
			return errors.New("Fatal error, current item from store never exist at warehouse")
		}
		return err
	}

	return nil
}

// SetActivate implements WarehouseService.
func (service *WarehouseServiceImpl) SetActivate(tenantId int, itemId int, setInto bool) error {
	if itemId == 0 {
		return errors.New("Item ID could not be empty or fill with 0")
	}
	if tenantId == 0 {
		return errors.New("Required tenant id is empty")
	}

	err := service.Repository.SetActivate(tenantId, itemId, setInto)
	if err != nil {
		if strings.Contains(err.Error(), "PGRST116") {
			return errors.New("Fatal error, Could not edit current item. current item never exist at warehouse")
		}
		return err
	}

	return nil
}

// FindCompleteById implements WarehouseService.
func (service *WarehouseServiceImpl) FindCompleteById(itemId int, tenantId int) (*model.CategoryWithItem, error) {
	if itemId == 0 {
		return nil, errors.New("Item ID could not be empty or fill <= 0")
	}
	if tenantId == 0 {
		return nil, errors.New("Required tenant id is empty or fill <= 0")
	}

	item, err := service.Repository.FindCompleteById(itemId, tenantId)
	if err != nil {
		if err.Error() == "NO_DATA_FOUND" {
			return nil, errors.New("No data return or non exist data")
		}

		if err.Error() == "CARDINALITY_VIOLATION" {
			return nil, errors.New("Fatal error ! Current item is not valid, duplicate assigning this item category values may cause this error")
		}

		return nil, err
	}

	return item, nil
}
