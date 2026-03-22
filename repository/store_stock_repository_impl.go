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
func (repository *StoreStockRepositoryImpl) GetV2(
	tenantId int,
	storeId int,
	limit int,
	page int,
	nameQuery string,
	categoryId int,
) ([]*model.StoreStockV2, int, error) {
	offset := page * limit
	var results []*model.StoreStockV2

	baseQuery := func() *gorm.DB {
		q := repository.Client.
			Table("store_stock").
			Joins("INNER JOIN warehouse ON warehouse.item_id = store_stock.item_id").
			Joins("LEFT JOIN category_mtm_warehouse ON category_mtm_warehouse.item_id = warehouse.item_id").
			Joins("LEFT JOIN category ON category.id = category_mtm_warehouse.category_id").
			Where("store_stock.tenant_id = ? AND store_stock.store_id = ?", tenantId, storeId)

		if nameQuery != "" {
			q = q.Where("LOWER(warehouse.item_name) LIKE LOWER(?)", "%"+nameQuery+"%")
		}

		if categoryId != 0 {
			q = q.Where("category_mtm_warehouse.category_id = ?", categoryId)
		} else if categoryId == -1 {
			q = q.Where("category_mtm_warehouse.category_id IS NULL")
		}

		return q
	}

	var totalCount int64
	err := baseQuery().
		Select("store_stock.id"). // select minimal for count
		Count(&totalCount).Error
	if err != nil {
		log.Errorf(
			"ERROR! StoreStockRepositoryImpl.GetV2 count query — tenantId: %d, storeId: %d — %s",
			tenantId, storeId, err.Error(),
		)
		return nil, 0, err
	}

	// Early return if no results — skip data query entirely
	if totalCount == 0 {
		return []*model.StoreStockV2{}, 0, nil
	}

	err = baseQuery().
		Select(`
			store_stock.id,
			store_stock.price,
			store_stock.stocks,
			store_stock.created_at,
			warehouse.item_id,
			warehouse.item_name,
			warehouse.stock_type,
			warehouse.base_price,
			warehouse.is_active,
			COALESCE(category.id, 0)             AS category_id,
			COALESCE(category.category_name, '') AS category_name
		`).
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		log.Errorf(
			"ERROR! StoreStockRepositoryImpl.GetV2 tenantId: %d, storeId: %d, categoryId: %d — %s",
			tenantId, storeId, categoryId, err.Error(),
		)
		return nil, 0, err
	}

	return results, int(totalCount), nil
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
