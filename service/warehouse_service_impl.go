package service

import (
	"cashier-api/model"
	"cashier-api/repository"
)

type WarehouseServiceImpl struct {
	Repository repository.WarehouseRepository
}

func NewWarehouseServiceImpl(repository repository.WarehouseRepository) WarehouseService {
	return &WarehouseServiceImpl{Repository: repository}
}

func (service *WarehouseServiceImpl) GetWarehouseItems(tenantId, limit, page int) ([]*model.Item, int, error) {
	return service.Repository.Get(tenantId, limit, page-1)
}
