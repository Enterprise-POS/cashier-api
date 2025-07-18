package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"fmt"
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
