package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
)

const OrderItemTable string = "order_item"

type OrderItemRepositoryImpl struct {
	Client *supabase.Client
}

func NewOrderItemRepositoryImpl(client *supabase.Client) *OrderItemRepositoryImpl {
	return &OrderItemRepositoryImpl{
		Client: client,
	}
}

func (repository *OrderItemRepositoryImpl) PlaceOrderItem(orderItem *model.OrderItem) (*model.OrderItem, error) {
	var createdOrderItem *model.OrderItem
	_, err := repository.Client.From(OrderItemTable).
		Insert(orderItem, false, "", "representation", "").

		// By default insert can put multiple OrderItem
		// But in this method only one,
		// .Single will return *model.OrderItem data not []*model.OrderItem
		Single().

		// Supabase will handle the json.Unmarshal
		ExecuteTo(&createdOrderItem)
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

	return createdOrderItem, nil
}

/*
QueryFilter:

	This will be used by many repository that may be need filter
	example usage logic:
		CreatedAtAsc = false -> will order by descending
		CreatedAtAsc = true -> will order by ascending
*/

func (repository *OrderItemRepositoryImpl) Get(tenantId int, limit int, page int, filters []*query.QueryFilter) ([]*model.OrderItem, int, error) {
	start := page * limit
	end := start + limit - 1

	var results []*model.OrderItem
	filterBuilder := repository.Client.From(OrderItemTable).
		Select("*", "exact", false).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "")

		// Apply filter
	for _, filter := range filters {
		if filter.Column == "" {
			log.Warnf("WARN ! handled error, some filter is an empty string. from tenantId: %d", tenantId)
			return nil, 0, fmt.Errorf("WARN ! handled error, some filter is an empty string. from tenantId: %d", tenantId)
		} else {
			filterBuilder = filterBuilder.Order(filter.Column, &postgrest.OrderOpts{Ascending: filter.Ascending})
		}
	}

	// Execute / request to DB
	count, err := filterBuilder.ExecuteTo(&results)
	if err != nil {
		return nil, 0, err
	}

	return results, int(count), nil
}
