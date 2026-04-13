package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"errors"
	"fmt"

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
	queryFilters []*query.QueryFilter,
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
			// Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}) // Order("created_at DESC")

		if nameQuery != "" {
			q = q.Where("LOWER(warehouse.item_name) LIKE LOWER(?)", "%"+nameQuery+"%")
		}

		if categoryId != 0 {
			q = q.Where("category_mtm_warehouse.category_id = ?", categoryId)
		} else if categoryId == -1 {
			q = q.Where("category_mtm_warehouse.category_id IS NULL")
		}

		hasCustomOrder := false
		for _, f := range queryFilters {
			if f.Column == query.CreatedAtColumn {
				hasCustomOrder = true
				if f.Ascending {
					q = q.Order("store_stock.created_at ASC")
				} else {
					q = q.Order("store_stock.created_at DESC")
				}
				break
			}
		}

		if !hasCustomOrder {
			q = q.Order("store_stock.created_at DESC")
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
			store_stock.updated_at,
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
	return repository.Client.Transaction(func(tx *gorm.DB) error {
		// Single query: fetch warehouse item + its matching store stock in one preload
		var warehouseItem model.Item
		err := tx.
			Preload("StoreStocks", "store_id = ? AND tenant_id = ?", storeId, tenantId).
			Where("item_id = ? AND tenant_id = ?", itemId, tenantId).
			Set("gorm:query_option", "FOR UPDATE").
			Take(&warehouseItem).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("[ERROR] Fatal error, current item from store never exist at warehouse")
		}
		if err != nil {
			return err
		}

		// Validate store stock exists (preload returns empty slice if not found)
		if len(warehouseItem.StoreStocks) == 0 {
			return errors.New("[ERROR] Not exist item at the store or invalid item")
		}

		storeStock := warehouseItem.StoreStocks[0]

		// Validate stock sufficiency before any mutations
		realizedStoreStock := storeStock.Stocks - quantity
		if realizedStoreStock < 0 {
			return errors.New("[ERROR] Not enough stock")
		}

		// Update warehouse stock
		err = tx.
			Model(&model.Item{}).
			Where("item_id = ? AND tenant_id = ?", itemId, tenantId).
			Update("stocks", warehouseItem.Stocks+quantity).Error
		if err != nil {
			return err
		}

		// Update store stock
		return tx.
			Model(&model.StoreStock{}).
			Where("id = ?", storeStock.Id).
			Update("stocks", realizedStoreStock).Error
	})
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
	return repository.Client.Transaction(func(tx *gorm.DB) error {
		// Fetch warehouse item + matching store stock in one preload, with row lock
		var warehouseItem model.Item
		err := tx.
			Preload("StoreStocks", "store_id = ? AND tenant_id = ?", storeId, tenantId).
			Where("item_id = ? AND tenant_id = ?", itemId, tenantId).
			Set("gorm:query_option", "FOR UPDATE").
			Take(&warehouseItem).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("[ERROR] Not exist item at the warehouse or invalid item")
		}
		if err != nil {
			return err
		}

		// Validate stock sufficiency
		realizedWarehouseStock := warehouseItem.Stocks - quantity
		if realizedWarehouseStock < 0 {
			return errors.New("[ERROR] Not enough stock")
		}

		// Upsert store stock
		if len(warehouseItem.StoreStocks) > 0 {
			// Update existing store stock
			storeStock := warehouseItem.StoreStocks[0]
			err = tx.
				Model(&model.StoreStock{}).
				Where("id = ?", storeStock.Id).
				Update("stocks", gorm.Expr("stocks + ?", quantity)).Error
			if err != nil {
				return err
			}
		} else {
			// Create new store stock row
			err = tx.Create(&model.StoreStock{
				ItemId:   itemId,
				Stocks:   quantity,
				StoreId:  storeId,
				TenantId: tenantId,
			}).Error
			if err != nil {
				return err
			}
		}

		// Deduct warehouse stock
		return tx.
			Model(&model.Item{}).
			Where("item_id = ? AND tenant_id = ?", itemId, tenantId).
			Update("stocks", realizedWarehouseStock).Error
	})
}

// Edit implements StoreStockRepository.
func (repository *StoreStockRepositoryImpl) Edit(item *model.StoreStock) error {
	// Validate item exists
	var count int64
	err := repository.Client.Model(&model.StoreStock{}).
		Where("item_id = ? AND tenant_id = ? AND store_id = ? AND id = ?", item.ItemId, item.TenantId, item.StoreId, item.Id).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("Item does not exist at this store or invalid item ID")
	}

	// Perform update
	result := repository.Client.Model(&model.StoreStock{Id: item.Id}).
		Where("tenant_id = ?", item.TenantId).
		Updates(map[string]any{
			"price": item.Price, // GORM auto-updates UpdatedAt
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("No stock found for tenant_id %d and store_id %d and store_stock.id %d",
			item.TenantId, item.StoreId, item.Id)
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
