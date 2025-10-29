package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

type StoreStockRepositoryImpl struct {
	Client *supabase.Client
}

const StoreStockTable string = "store_stock"

func NewStoreStockRepositoryImpl(client *supabase.Client) StoreStockRepository {
	return &StoreStockRepositoryImpl{Client: client}
}

func (storeStock *StoreStockRepositoryImpl) Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error) {
	// Even user see first page written in 1, we must subtract by 1 otherwise range error
	start := page * limit
	end := start + limit - 1

	data, count, err := storeStock.Client.From(StoreStockTable).
		// exact: Will return the total items are there
		Select("*", "exact", false).

		// Only return the requested tenant
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "").
		Limit(limit, "").
		Execute()

	if err != nil {
		return nil, 0, err
	}

	var result []*model.StoreStock
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Error(err.Error())
		return nil, 0, errors.New("fatal error while converting data")
	}

	return result, int(count), nil
}

/*
TransferStockToWarehouse:

	This RPC also decrease/increment the stocks at warehouse also store_stock

	Un-exist item_id at warehouse will cause FATAL ERROR

	rather than use model.StoreStock, quantity is required,
	We want to prevent race condition at the DB.

	(warehouse -> store_stock)

	TODO: resolve security alert from supabase, 'search_path'
*/
func (storeStock *StoreStockRepositoryImpl) TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error {
	var message string = storeStock.Client.Rpc("transfer_stock_to_warehouse", "", map[string]interface{}{
		"p_quantity":  quantity,
		"p_item_id":   itemId,
		"p_store_id":  storeId,
		"p_tenant_id": tenantId})

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	return nil
}

/*
TransferStockToStoreStock:

	This RPC also decrease/increment the stocks at warehouse also store_stock

	By default if current item stored but 'never exist' at the 'store_stock',
	it will create price with default 'price = 0'

	rather than use model.Item, quantity is required,
	We want to prevent race condition at the DB.

	(store_stock -> warehouse)

	TODO: resolve security alert from supabase, 'search_path'
*/
func (storeStock *StoreStockRepositoryImpl) TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error {
	var message string = storeStock.Client.Rpc("transfer_stocks_to_store_stock", "", map[string]interface{}{
		"p_quantity":  quantity,
		"p_item_id":   itemId,
		"p_store_id":  storeId,
		"p_tenant_id": tenantId})

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	return nil
}
