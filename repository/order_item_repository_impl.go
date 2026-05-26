package repository

import (
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const OrderItemTable string = "order_item"

type OrderItemRepositoryImpl struct {
	Client *gorm.DB
}

func NewOrderItemRepositoryImpl(client *gorm.DB) OrderItemRepository {
	return &OrderItemRepositoryImpl{
		Client: client,
	}
}

func (repository *OrderItemRepositoryImpl) PlaceOrderItem(orderItem *model.OrderItem) (*model.OrderItem, error) {
	err := repository.Client.Create(orderItem).Error

	if err != nil {
		if strings.Contains(err.Error(), "(23514)") {
			// Example violation error: (23514) new row for relation \"order_item\" violates check constraint \"order_item_quantity_check\"
			log.Warnf("Warning ! handled error, violation invalid attempt to insert invalid value detected while place order item with tenant_id: %d, store_id: %d", orderItem.TenantId, orderItem.StoreId)
		} else if strings.Contains(err.Error(), "(23503)") {
			// Example violation error: (23514) new row for relation "order_item" violates check constraint "order_item_discount_amount_check"
			log.Warnf("Warning ! handled error, violation unavailable foreign key detected while place order item with tenant_id: %d, store_id: %d", orderItem.TenantId, orderItem.StoreId)
		} else {
			// Fatal error
			log.Errorf("Error while place order item with tenant_id: %d, store_id: %d", orderItem.TenantId, orderItem.StoreId)
		}
		return nil, err
	}

	return orderItem, nil
}

/*
QueryFilter:

	This will be used by many repository that may be need filter
	example usage logic:
		CreatedAtAsc = false -> will order by descending
		CreatedAtAsc = true -> will order by ascending
*/

func (repository *OrderItemRepositoryImpl) Get(
	tenantId int,
	storeId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.OrderItem, int, error) {
	start := page * limit
	//end := start + limit - 1

	var results []*model.OrderItem
	var totalCount int64

	db := repository.Client.Model(&model.OrderItem{}).
		Where("tenant_id = ?", tenantId)

	if storeId > 0 {
		db = db.Where("store_id = ?", storeId)
	}

	// Apply date filter
	if dateFilter != nil {
		// This is on demand filter may change in the future, maybe use column such as update_at, etc...
		dateFilter.Column = "created_at"

		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			// Range: 1 Dec 2025 - 31 Dec 2025
			// fmt.Println(common.EpochToRFC3339(*dateFilter.StartDate), common.EpochToRFC3339(*dateFilter.EndDate))
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("created_at >= ? AND created_at < ?", startDate, endDate)
		} else if dateFilter.StartDate != nil {
			// Only start date (from 1 Dec 2025 onwards)
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			db = db.Where("created_at >= ?", startDate)
		} else if dateFilter.EndDate != nil {
			// Only end date (up to 31 Dec 2025)
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("created_at <= ?", endDate)
		}
	}

	// Count total available for this request
	db.Count(&totalCount)

	// Apply filter
	for _, filter := range filters {
		// ORDER BY
		if filter.Column == "" {
			log.Warnf("WARN ! handled error, some filter is an empty string. from tenantId: %d", tenantId)
			return nil, 0, fmt.Errorf("WARN ! handled error, some filter is an empty string. from tenantId: %d", tenantId)
		}

		// DESC / ASCENDING
		direction := "DESC"
		if filter.Ascending {
			direction = "ASC"
		}
		db = db.Order(fmt.Sprintf("%s %s", filter.Column, direction))
	}

	// Apply pagination
	result := db.Limit(limit).Offset(start).Find(&results)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return results, int(totalCount), nil
}

// Transactions implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) Transactions(params *CreateTransactionParams) (*TransactionDataReturn, error) {
	itemsJSON, err := json.Marshal(params.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize items: %w", err)
	}

	var transactionDataReturn *TransactionDataReturn
	// Because it's return row, use SELECT *
	result := repository.Client.Raw("SELECT * FROM transactions($1, $2, $3, $4, $5, $6::JSONB, $7, $8, $9)",
		params.PurchasedPrice,
		params.TotalQuantity,
		params.TotalAmount,
		params.DiscountAmount,
		params.SubTotal,

		string(itemsJSON), // cast to JSONB in the query

		params.UserId,
		params.TenantId,
		params.StoreId,
	).Scan(&transactionDataReturn)
	fmt.Println(result)

	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			log.Warnf("PostgreSQL error during transaction: code=%s, message=%s", pgErr.Code, pgErr.Message)
			return nil, errors.New(pgErr.Message) // return clean message to caller (service layer)
		}

		log.Errorf("Unexpected error during transaction: %v", result.Error)
		return nil, result.Error
	}

	return transactionDataReturn, nil
}

// FindById implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) FindById(orderItemId int, tenantId int) (*model.OrderItemWithStore, []*model.PurchasedItem, error) {
	type row struct {
		// purchased_item
		PurchasedItemId             int    `gorm:"column:purchased_item_id"`
		ItemId                      int    `gorm:"column:item_id"`
		StorePriceSnapshot          int    `gorm:"column:store_price_snapshot"`
		BasePriceSnapshot           int    `gorm:"column:base_price_snapshot"`
		Quantity                    int    `gorm:"column:quantity"`
		PurchasedItemDiscountAmount int    `gorm:"column:purchased_item_discount_amount"`
		PurchasedItemTotalAmount    int    `gorm:"column:purchased_item_total_amount"`
		ItemNameSnapshot            string `gorm:"column:item_name_snapshot"`

		// order_item
		OrderItemId             int       `gorm:"column:order_item_id"`
		PurchasedPrice          int       `gorm:"column:purchased_price"`
		Subtotal                int       `gorm:"column:subtotal"`
		TotalQuantity           int       `gorm:"column:total_quantity"`
		OrderItemTotalAmount    int       `gorm:"column:order_item_total_amount"`
		OrderItemDiscountAmount int       `gorm:"column:order_item_discount_amount"`
		CreatedAt               time.Time `gorm:"column:created_at"`
		StoreId                 int       `gorm:"column:store_id"`

		// store
		StoreName string `gorm:"column:store_name"`
	}

	var rows []row // local struct

	err := repository.Client.
		Model(&model.OrderItem{}). // Will exclude deleted_at
		Select(`
			purchased_item_list.id                  AS purchased_item_id,
			purchased_item_list.item_id,
			purchased_item_list.store_price_snapshot,
			purchased_item_list.base_price_snapshot,
			purchased_item_list.quantity,
			purchased_item_list.discount_amount     AS purchased_item_discount_amount,
			purchased_item_list.total_amount        AS purchased_item_total_amount,
			purchased_item_list.item_name_snapshot,
			order_item.id                           AS order_item_id,
			order_item.purchased_price,
			order_item.subtotal,
			order_item.total_quantity,
			order_item.total_amount                 AS order_item_total_amount,
			order_item.discount_amount              AS order_item_discount_amount,
			order_item.created_at,
			order_item.store_id,
			store.name                              AS store_name
		`).
		Joins("INNER JOIN purchased_item_list ON purchased_item_list.order_item_id = order_item.id").
		Joins("LEFT JOIN store ON store.id = order_item.store_id").
		Where("order_item.tenant_id = ? AND order_item.id = ?", tenantId, orderItemId).
		Scan(&rows).Error

	if err != nil {
		return nil, nil, fmt.Errorf("FindById query failed: %w", err)
	}
	if len(rows) == 0 {
		log.Warnf("At OrderItemRepositoryImpl.FindById with order item %d not found for tenant %d", orderItemId, tenantId)
		return nil, nil, fmt.Errorf("No data found. for order item id %d", orderItemId)
	}

	// Extract OrderItem from first row (since it's the same for all rows)
	first := rows[0]
	orderItem := &model.OrderItemWithStore{
		Id:             first.OrderItemId,
		PurchasedPrice: first.PurchasedPrice,
		Subtotal:       first.Subtotal,
		TotalQuantity:  first.TotalQuantity,
		TotalAmount:    first.OrderItemTotalAmount,
		DiscountAmount: first.OrderItemDiscountAmount,
		CreatedAt:      first.CreatedAt,
		StoreId:        first.StoreId,
		TenantId:       tenantId,
		StoreName:      first.StoreName,
	}

	// Extract all PurchasedItems
	purchasedItemList := make([]*model.PurchasedItem, len(rows))
	for i, r := range rows {
		purchasedItemList[i] = &model.PurchasedItem{
			Id:                 r.PurchasedItemId,
			ItemId:             r.ItemId,
			StorePriceSnapshot: r.StorePriceSnapshot,
			BasePriceSnapshot:  r.BasePriceSnapshot,
			Quantity:           r.Quantity,
			DiscountAmount:     r.PurchasedItemDiscountAmount,
			TotalAmount:        r.PurchasedItemTotalAmount,
			ItemNameSnapshot:   r.ItemNameSnapshot,
			OrderItemId:        orderItemId,
		}
	}

	return orderItem, purchasedItemList, nil
}

// GetProfitReport implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) GetProfitReport(tenantId int, storeId int, dateFilter *query.DateFilter) ([]*ProfitReportRow, error) {
	db := repository.Client.Table("purchased_item_list pil").
		Select(`
			pil.item_id,
			MAX(pil.item_name_snapshot) AS item_name,
			SUM(pil.quantity) AS total_quantity,
			SUM(pil.total_amount) AS total_revenue,
			SUM(pil.base_price_snapshot * pil.quantity) AS total_cogs,
			SUM(pil.discount_amount * pil.quantity) AS total_discount,
			SUM(pil.total_amount) - SUM(pil.base_price_snapshot * pil.quantity) AS total_profit
		`).
		Joins("JOIN order_item oi ON oi.id = pil.order_item_id").
		Where("oi.tenant_id = ?", tenantId).
		Group("pil.item_id").
		Order("total_profit DESC")

	if storeId > 0 {
		db = db.Where("oi.store_id = ?", storeId)
	}

	if dateFilter != nil {
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("oi.created_at >= ? AND oi.created_at < ?", startDate, endDate)
		} else if dateFilter.StartDate != nil {
			startDate := common.EpochToRFC3339(*dateFilter.StartDate)
			db = db.Where("oi.created_at >= ?", startDate)
		} else if dateFilter.EndDate != nil {
			endDate := common.EpochToRFC3339(*dateFilter.EndDate)
			db = db.Where("oi.created_at < ?", endDate)
		}
	}

	var rows []*ProfitReportRow
	if err := db.Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

// GetTenantAndStoreName implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) GetTenantAndStoreName(tenantId int, storeId int) (string, string, error) {
	var tenantName string
	if err := repository.Client.Model(&model.Tenant{}).
		Select("name").
		Where("id = ?", tenantId).
		Scan(&tenantName).Error; err != nil {
		return "", "", err
	}

	if storeId <= 0 {
		return tenantName, "All Stores", nil
	}

	var storeName string
	if err := repository.Client.Model(&model.Store{}).
		Select("name").
		Where("id = ? AND tenant_id = ?", storeId, tenantId).
		Scan(&storeName).Error; err != nil {
		return "", "", err
	}

	return tenantName, storeName, nil
}

// GetReport implements [OrderItemRepository].
func (repository *OrderItemRepositoryImpl) GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*SalesReport, error) {
	// Base condition builder to avoid repeating date filter logic
	applyFilters := func(db *gorm.DB, tablePrefix string) *gorm.DB {
		db = db.Where(tablePrefix+"tenant_id = ?", tenantId)

		if storeId > 0 {
			db = db.Where(tablePrefix+"store_id = ?", storeId)
		}

		if dateFilter != nil {
			if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
				db = db.Where(tablePrefix+"created_at >= to_timestamp(?) AND "+tablePrefix+"created_at < to_timestamp(?)",
					*dateFilter.StartDate, *dateFilter.EndDate)
			} else if dateFilter.StartDate != nil {
				db = db.Where(tablePrefix+"created_at >= to_timestamp(?)", *dateFilter.StartDate)
			} else if dateFilter.EndDate != nil {
				db = db.Where(tablePrefix+"created_at <= to_timestamp(?)", *dateFilter.EndDate)
			}
		}

		return db
	}

	// order_summary — Model() applies deleted_at IS NULL automatically
	type orderSummary struct {
		SumPurchasedPrice int
		SumSubtotal       int
		SumTotalQuantity  int
		SumDiscountAmount int
		SumTotalAmount    int
		SumTransactions   int
	}
	var oSummary orderSummary

	orderQuery := applyFilters(repository.Client.Model(&model.OrderItem{}), "")
	err := orderQuery.Select(`
        COALESCE(SUM(purchased_price), 0)  AS sum_purchased_price,
        COALESCE(SUM(subtotal), 0)         AS sum_subtotal,
        COALESCE(SUM(total_quantity), 0)   AS sum_total_quantity,
        COALESCE(SUM(discount_amount), 0)  AS sum_discount_amount,
        COALESCE(SUM(total_amount), 0)     AS sum_total_amount,
        COALESCE(COUNT(id), 0)             AS sum_transactions
    `).Scan(&oSummary).Error
	if err != nil {
		return nil, fmt.Errorf("GetSalesReport order_summary failed: %w", err)
	}

	// profit_summary — raw join since purchased_item_list is not a model
	// Still filters order_item with deleted_at via the JOIN condition
	type profitSummary struct {
		SumProfit int
	}
	var pSummary profitSummary

	profitQuery := applyFilters(
		repository.Client.Table("purchased_item_list pil").
			Joins("INNER JOIN order_item oi ON oi.id = pil.order_item_id AND oi.deleted_at IS NULL"),
		"oi.",
	)
	err = profitQuery.Select(`
        COALESCE(SUM(pil.total_amount) - SUM(pil.base_price_snapshot * pil.quantity), 0) AS sum_profit
    `).Scan(&pSummary).Error
	if err != nil {
		return nil, fmt.Errorf("GetSalesReport profit_summary failed: %w", err)
	}

	return &SalesReport{
		SumPurchasedPrice: oSummary.SumPurchasedPrice,
		SumSubtotal:       oSummary.SumSubtotal,
		SumTotalQuantity:  oSummary.SumTotalQuantity,
		SumDiscountAmount: oSummary.SumDiscountAmount,
		SumTotalAmount:    oSummary.SumTotalAmount,
		SumProfit:         pSummary.SumProfit,
		SumTransactions:   oSummary.SumTransactions,
	}, nil
}

// DeleteInvoice implements [OrderItemRepository].
func (repository *OrderItemRepositoryImpl) DeleteInvoice(orderItemId int, tenantId int) error {
	// In the future will be implement security such as:
	// - Only owner could delete the invoice
	// - Permission request
	result := repository.Client.
		Where("id = ? AND tenant_id = ?", orderItemId, tenantId).
		Delete(&model.OrderItem{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order item %d not found", orderItemId)
	}

	return nil
}
