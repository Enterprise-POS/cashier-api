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
func (repository *OrderItemRepositoryImpl) Transactions(params *CreateTransactionParams) (int, error) {
	itemsJSON, err := json.Marshal(params.Items)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize items: %w", err)
	}

	var createdOrderItemId int
	result := repository.Client.Raw("SELECT transactions($1, $2, $3, $4, $5, $6::JSONB, $7, $8, $9)",
		params.PurchasedPrice,
		params.TotalQuantity,
		params.TotalAmount,
		params.DiscountAmount,
		params.SubTotal,

		string(itemsJSON), // cast to JSONB in the query

		params.UserId,
		params.TenantId,
		params.StoreId,
	).Scan(&createdOrderItemId)

	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			log.Warnf("PostgreSQL error during transaction: code=%s, message=%s", pgErr.Code, pgErr.Message)
			return 0, fmt.Errorf(pgErr.Message) // return clean message to caller (service layer)
		}

		log.Errorf("Unexpected error during transaction: %v", result.Error)
		return 0, result.Error
	}

	return createdOrderItemId, nil
}

// FindById implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) FindById(orderItemId int, tenantId int) (*model.OrderItem, []*model.PurchasedItem, error) {
	var results []struct {
		// PurchasedItem fields
		Id                 int    `gorm:"column:id"`
		ItemId             int    `gorm:"column:item_id"`
		StorePriceSnapshot int    `gorm:"column:store_price_snapshot"`
		BasePriceSnapshot  int    `gorm:"column:base_price_snapshot"`
		Quantity           int    `gorm:"column:quantity"`
		DiscountAmount     int    `gorm:"column:discount_amount"`
		TotalAmount        int    `gorm:"column:total_amount"`
		ItemNameSnapshot   string `gorm:"column:item_name_snapshot"`

		// OrderItem fields
		OrderItemId                  int       `gorm:"column:order_item_id"`
		OrderItemStorePurchasedPrice int       `gorm:"column:order_item_purchased_price"`
		OrderItemSubtotal            int       `gorm:"column:order_item_subtotal"`
		OrderItemTotalQuantity       int       `gorm:"column:order_item_total_quantity"`
		OrderItemTotalAmount         int       `gorm:"column:order_item_total_amount"`
		OrderItemCreatedAt           time.Time `gorm:"column:order_item_created_at"`
		OrderItemStoreId             int       `gorm:"column:order_item_store_id"`
	}

	result := repository.Client.Raw("SELECT * FROM get_order_item_details_by_id(?, ?)", orderItemId, tenantId).Scan(&results)
	if result.Error != nil {
		return nil, nil, result.Error
	}
	if len(results) == 0 {
		return nil, nil, errors.New("no data found")
	}

	// Extract OrderItem from first row (since it's the same for all rows)
	orderItem := &model.OrderItem{
		Id:             results[0].OrderItemId,
		PurchasedPrice: results[0].OrderItemStorePurchasedPrice,
		Subtotal:       results[0].OrderItemSubtotal,
		TotalQuantity:  results[0].OrderItemTotalQuantity,
		TotalAmount:    results[0].OrderItemTotalAmount,
		CreatedAt:      results[0].OrderItemCreatedAt,
		TenantId:       tenantId,
		DiscountAmount: 0,
		StoreId:        results[0].OrderItemStoreId,
	}

	// Extract all PurchasedItems
	var purchasedItemList []*model.PurchasedItem
	for _, row := range results {
		purchasedItemList = append(purchasedItemList, &model.PurchasedItem{
			Id:                 row.Id,
			ItemId:             row.ItemId,
			StorePriceSnapshot: row.StorePriceSnapshot,
			BasePriceSnapshot:  row.BasePriceSnapshot,
			Quantity:           row.Quantity,
			DiscountAmount:     row.DiscountAmount,
			TotalAmount:        row.TotalAmount,
			ItemNameSnapshot:   row.ItemNameSnapshot,

			// We don't request the order_item_id because
			// we already know if the data return it's guaranteed
			// that the order_item_id is from parameter is correct
			OrderItemId: orderItemId,
		})
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

// GetReport implements OrderItemRepository.
func (repository *OrderItemRepositoryImpl) GetSalesReport(tenantId int, storeId int, dateFilter *query.DateFilter) (*SalesReport, error) {
	var salesReport []*SalesReport

	var result *gorm.DB
	if dateFilter != nil {
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
			result = repository.Client.Raw("SELECT * FROM sales_report(?, ?, ?, ?)",
				tenantId, storeId, *dateFilter.StartDate, *dateFilter.EndDate).Scan(&salesReport)
		} else if dateFilter.StartDate != nil {
			result = repository.Client.Raw("SELECT * FROM sales_report(?, ?, ?, NULL)",
				tenantId, storeId, *dateFilter.StartDate).Scan(&salesReport)
		} else if dateFilter.EndDate != nil {
			result = repository.Client.Raw("SELECT * FROM sales_report(?, ?, NULL, ?)",
				tenantId, storeId, *dateFilter.EndDate).Scan(&salesReport)
		} else {
			result = repository.Client.Raw("SELECT * FROM sales_report(?, ?)",
				tenantId, storeId).Scan(&salesReport)
		}
	} else {
		result = repository.Client.Raw("SELECT * FROM sales_report(?, ?)",
			tenantId, storeId).Scan(&salesReport)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	if len(salesReport) == 0 {
		log.Errorf("Sales report return nil. tenantId: %d, storeId: %d", tenantId, storeId)
		return nil, errors.New("unexpected error while requesting sales report. Please try again later")
	}

	return salesReport[0], nil
}
