package repository

import "cashier-api/model"

type StoreStockRepository interface {
	Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error)

	TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error
	TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error

	// FindById(itemId int, tenantId int) *model.StoreStock
	// CreateItem(item []*model.Item) error
	// Edit(quantity int, item *model.Item) error

	/*
		Create new store, will
	*/
	Create() error
}
