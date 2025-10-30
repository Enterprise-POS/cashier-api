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

func (repository *StoreStockRepositoryImpl) Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error) {
	// Even user see first page written in 1, we must subtract by 1 otherwise range error
	start := page * limit
	end := start + limit - 1

	data, count, err := repository.Client.From(StoreStockTable).
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

// GetV2 implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) GetV2(tenantId int, storeId int, limit int, page int, nameQuery string) ([]*model.StoreStockV2, int, error) {
	start := page * limit
	end := start + limit - 1

	/*
		Example join using bare bone supabase method.

		results, _, err := repository.Client.From("category").
			Select("id, category_name, category_mtm_warehouse(warehouse(item_id))", "", false).
			Limit(limit, "category_mtm_warehouse(warehouse(item_id))").
			Range(start, end, "category_mtm_warehouse(warehouse(item_id))").
			Eq("tenant_id", strconv.Itoa(tenantId)).
			Execute()

		Instead will be using Rpc with the same query as above
	*/
	query := repository.Client.
		From(StoreStockTable).
		Select("id, stocks, price, item_id, warehouse(item_name)", "exact", false).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Limit(limit, "").
		Range(start, end, "")

	if nameQuery != "" {
		query = query.Like("item_name", "%"+nameQuery+"%")
	}

	var result []*model.StoreStockV2
	count, err := query.ExecuteTo(&result)
	if err != nil {
		return nil, 0, err
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
func (repository *StoreStockRepositoryImpl) TransferStockToWarehouse(quantity int, itemId int, storeId int, tenantId int) error {
	var message string = repository.Client.Rpc("transfer_stock_to_warehouse", "", map[string]interface{}{
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
func (repository *StoreStockRepositoryImpl) TransferStockToStoreStock(quantity int, itemId int, storeId int, tenantId int) error {
	var message string = repository.Client.Rpc("transfer_stocks_to_store_stock", "", map[string]interface{}{
		"p_quantity":  quantity,
		"p_item_id":   itemId,
		"p_store_id":  storeId,
		"p_tenant_id": tenantId})

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	return nil
}
