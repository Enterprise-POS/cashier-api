package repository

import (
	"cashier-api/exception"
	"cashier-api/model"
	"encoding/json"
	"errors"
	"regexp"
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
	// end := start + limit - 1

	data := repository.Client.Rpc("get_store_stocks", "", map[string]interface{}{
		"p_tenant_id":  tenantId,
		"p_store_id":   storeId,
		"p_limit":      limit,
		"p_offset":     start,
		"p_name_query": nameQuery,
	})

	/*
		Example return, either error string contains error message or json string
		- [ERROR]
		- [
				{"category_id":1,"category_name":"Fruits","item_id":1,"item_name":"Apple","stocks":358},
				{"category_id":1,"category_name":"Fruits","item_id":267,"item_name":"Durian","stocks":10}
			]
	*/
	if strings.Contains(data, "[ERROR]") {
		// Extract the message
		var postgreSQLException *exception.PostgreSQLException
		err := json.Unmarshal([]byte(data), &postgreSQLException)
		if err != nil {
			return nil, 0, err
		}

		return nil, 0, postgreSQLException
	}

	var results []*model.StoreStockV2
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		log.Errorf("ERROR ! While unmarshaling data at CategoryRepositoryImpl.GetItemsByCategory. tenantId: %d, storeId: %d", tenantId, storeId)
		return nil, 0, err
	}

	countResult := 0
	if len(results) > 0 {
		countResult = results[0].TotalCount // Same value for all rows
	}

	return results, countResult, nil
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

// Edit implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) Edit(item *model.StoreStock) error {
	var message string = repository.Client.Rpc("edit_store_stock_item", "", map[string]interface{}{
		"p_store_stock_id": item.Id,
		"p_price":          item.Price, // Tobe updated price
		"p_store_id":       item.StoreId,
		"p_tenant_id":      item.StoreId,
		"p_item_id":        item.ItemId,
	})

	// [ERROR] || [FATAL ERROR]
	if strings.Contains(message, "ERROR") {
		// From first character [ take all the word until ], change to empty ("")
		// The blank at the end is required
		re := regexp.MustCompile(`\[[^\]]*\] `)
		cleanMsg := re.ReplaceAllString(message, "")
		return errors.New(cleanMsg)
	}

	return nil
}
