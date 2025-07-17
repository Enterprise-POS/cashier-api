package service

import (
	"cashier-api/repository"
)

type WarehouseServiceImpl struct {
	Repository repository.WarehouseRepository
}

func NewWarehouseServiceImpl(repository repository.WarehouseRepository) WarehouseService {
	return &WarehouseServiceImpl{Repository: repository}
}

func (service *WarehouseServiceImpl) GetWarehouseItems() {
}
