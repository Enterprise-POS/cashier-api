package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
)

type StoreStockServiceImpl struct {
	Repository        repository.StoreStockRepository
	ItemNameRegexRule *regexp.Regexp
}

func NewStoreStockServiceImpl(repository repository.StoreStockRepository) StoreStockService {
	return &StoreStockServiceImpl{
		Repository: repository,

		// The regex rule is the same with warehouse_service
		ItemNameRegexRule: regexp.MustCompile(
			`^[\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z][\p{Han}\p{Hiragana}\p{Katakana}a-zA-Z0-9' ]*$`),
	}
}

// Get implements StoreStockService.
func (service *StoreStockServiceImpl) Get(
	tenantId int,
	storeId int,
	limit int,
	page int,
) ([]*model.StoreStock, int, error) {
	if limit < 1 {
		return nil, 0, fmt.Errorf("Limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}
	if storeId < 1 {
		return nil, 0, errors.New("Store id could not be empty or fill with 0")
	}
	if tenantId < 1 {
		return nil, 0, errors.New("Tenant id could not be empty or fill with 0")
	}

	storeStocks, count, err := service.Repository.Get(tenantId, storeId, limit, page-1)
	if err != nil {
		return nil, 0, err
	}

	return storeStocks, count, nil
}

// GetV2 implements StoreStockService.
func (service *StoreStockServiceImpl) GetV2(
	tenantId int,
	storeId int,
	limit int,
	page int,
	nameQuery string,
) ([]*model.StoreStockV2, int, error) {
	if limit < 1 {
		return nil, 0, fmt.Errorf("Limit could not less then 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less then 1 (page >= 1). Given page %d", page)
	}
	if storeId < 1 {
		return nil, 0, errors.New("Store id could not be empty or fill with 0")
	}
	if tenantId < 1 {
		return nil, 0, errors.New("Tenant id could not be empty or fill with 0")
	}

	if nameQuery != "" {
		if !service.ItemNameRegexRule.MatchString(nameQuery) {
			return nil, 0, fmt.Errorf("Invalid searching by name: %s", nameQuery)
		}
	}

	storeStocksV2, count, err := service.Repository.GetV2(tenantId, storeId, limit, page-1, nameQuery)
	if err != nil {
		return nil, 0, err
	}

	return storeStocksV2, count, nil
}

// TransferStockToStoreStock implements StoreStockService.
func (service *StoreStockServiceImpl) TransferStockToStoreStock(
	quantity int,
	itemId int,
	storeId int,
	tenantId int,
) error {
	if quantity < 1 {
		return errors.New("Quantity could not be empty or fill with 0")
	}
	if itemId < 1 {
		return errors.New("Item id could not be empty or fill with 0")
	}
	if storeId < 1 {
		return errors.New("Store id could not be empty or fill with 0")
	}
	if tenantId < 1 {
		return errors.New("Tenant id could not be empty or fill with 0")
	}

	err := service.Repository.TransferStockToStoreStock(quantity, itemId, storeId, tenantId)
	if err != nil {
		return err
	}

	return nil
}

// TransferStockToWarehouse implements StoreStockService.
func (service *StoreStockServiceImpl) TransferStockToWarehouse(
	quantity int,
	itemId int,
	storeId int,
	tenantId int,
) error {
	if quantity < 1 {
		return errors.New("Quantity could not be empty or fill with 0")
	}
	if itemId < 1 {
		return errors.New("Item id could not be empty or fill with 0")
	}
	if storeId < 1 {
		return errors.New("Store id could not be empty or fill with 0")
	}
	if tenantId < 1 {
		return errors.New("Tenant id could not be empty or fill with 0")
	}

	err := service.Repository.TransferStockToWarehouse(quantity, itemId, storeId, tenantId)
	if err != nil {
		return err
	}

	return nil
}

// Edit implements StoreStockService.
func (service *StoreStockServiceImpl) Edit(tobeEditStoreStock *model.StoreStock) error {
	if tobeEditStoreStock.Price < 0 || tobeEditStoreStock.Price > 100_000_000 {
		return fmt.Errorf(
			"Invalid value for price, please check the update request price. %d is not allowed",
			tobeEditStoreStock.Price,
		)
	}
	if tobeEditStoreStock.ItemId < 1 {
		return errors.New("Item id could not be empty or fill with 0")
	}
	if tobeEditStoreStock.StoreId < 1 {
		return errors.New("Store id could not be empty or fill with 0")
	}
	if tobeEditStoreStock.TenantId < 1 {
		return errors.New("Tenant id could not be empty or fill with 0")
	}

	err := service.Repository.Edit(tobeEditStoreStock)
	if err != nil {
		return err
	}

	return nil
}
