package repository

import "cashier-api/model"

type StoreStockRepository interface {
	Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error)

	GetV2(tenantId int, storeId int, limit int, page int, nameQuery string) ([]*model.StoreStockV2, int, error)

	TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error
	TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error

	// FindById(itemId int, tenantId int) *model.StoreStock
	// CreateItem(item []*model.Item) error

	/*
		Even it's said to edit, the method not allowed for editing stock data,
		Other metadata such as 'price' is allowed
	*/
	Edit(item *model.StoreStock) error

	/*
		Yet there is no delete method for stock that quantity less than 0
	*/
	// Delete()
}
