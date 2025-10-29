package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
)

type StoreStockServiceImpl struct {
	Repository repository.StoreStockRepository
}

func NewStoreStockServiceImpl(repository repository.StoreStockRepository) StoreStockService {
	return &StoreStockServiceImpl{Repository: repository}
}

// Get implements StoreStockService.
func (service *StoreStockServiceImpl) Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error) {
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

// TransferStockToStoreStock implements StoreStockService.
func (service *StoreStockServiceImpl) TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error {
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
func (service *StoreStockServiceImpl) TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error {
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
