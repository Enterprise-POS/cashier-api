package repository

import (
	"cashier-api/model"
	"errors"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StoreStockRepositoryImpl struct {
	Client *gorm.DB
}

const StoreStockTable string = "store_stock"

func NewStoreStockRepositoryImpl(client *gorm.DB) StoreStockRepository {
	return &StoreStockRepositoryImpl{Client: client}
}

func (repository *StoreStockRepositoryImpl) Get(tenantId int, storeId int, limit int, page int) ([]*model.StoreStock, int, error) {
	start := page * limit

	var results []*model.StoreStock
	var totalCount int64

	query := repository.Client.Model(&model.StoreStock{}).
		Where("tenant_id = ?", tenantId)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(start).Limit(limit).Find(&results).Error; err != nil {
		return nil, 0, errors.New("fatal error while converting data")
	}

	return results, int(totalCount), nil
}

// GetV2 implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) GetV2(tenantId int, storeId int, limit int, page int, nameQuery string) ([]*model.StoreStockV2, int, error) {
	start := page * limit

	var results []*model.StoreStockV2
	err := repository.Client.Raw(
		"SELECT * FROM get_store_stocks(?, ?, ?, ?, ?)",
		tenantId, storeId, limit, start, nameQuery,
	).Scan(&results).Error
	if err != nil {
		log.Errorf("ERROR ! While querying data at StoreStockRepositoryImpl.GetV2. tenantId: %d, storeId: %d", tenantId, storeId)
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
	var message string
	err := repository.Client.Raw(
		"SELECT transfer_stock_to_warehouse(?, ?, ?, ?)",
		quantity, itemId, storeId, tenantId,
	).Scan(&message).Error
	if err != nil {
		return err
	}

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
	var message string
	err := repository.Client.Raw(
		"SELECT transfer_stocks_to_store_stock(?, ?, ?, ?)",
		quantity, itemId, storeId, tenantId,
	).Scan(&message).Error
	if err != nil {
		return err
	}

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	return nil
}

// Edit implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) Edit(item *model.StoreStock) error {
	var message string
	err := repository.Client.Raw(
		"SELECT edit_store_stock_item(?, ?, ?, ?, ?)",
		item.Id, item.Price, item.StoreId, item.TenantId, item.ItemId,
	).Scan(&message).Error
	if err != nil {
		return err
	}

	// [ERROR] || [FATAL ERROR]
	if strings.Contains(message, "ERROR") {
		// Strip the bracket prefix e.g. "[ERROR] " or "[FATAL ERROR] "
		re := regexp.MustCompile(`\[[^\]]*\] `)
		cleanMsg := re.ReplaceAllString(message, "")
		return errors.New(cleanMsg)
	}

	return nil
}

// LoadCashierData implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) LoadCashierData(tenantId int, storeId int) ([]*model.CashierData, error) {
	var cashierData []*model.CashierData
	err := repository.Client.Raw(
		"SELECT * FROM load_cashier_data(?, ?)",
		tenantId, storeId,
	).Scan(&cashierData).Error
	if err != nil {
		return nil, err
	}

	if len(cashierData) == 0 {
		return nil, errors.New("unexpected null response from database")
	}

	return cashierData, nil
}
