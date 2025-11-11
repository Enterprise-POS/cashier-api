package service

import "cashier-api/model"

type StoreStockService interface {
	/*
		Will get all available stock at requested storeId
	*/
	Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error)

	GetV2(tenantId int, storeId int, limit int, page int, nameQuery string) ([]*model.StoreStockV2, int, error)

	Edit(tobeEditStoreStock *model.StoreStock) error

	/*
		There is no create method for store_stock, so from warehouse transfer stock into store_stock
		warehouse quantity is always mandatory, could not transfer stock to store_stock if quantity insufficient
		will do the same for store_stock, could not transfer stock to warehouse if quantity insufficient
	*/
	TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error  // store_stock -> warehouse
	TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error // warehouse -> store_stock
}
