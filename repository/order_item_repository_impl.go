package repository

import (
	"cashier-api/model"
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

const ORDER_ITEM_TABLE string = "order_item"

type OrderItemRepositoryImpl struct {
	Client *supabase.Client
}

func (repository *OrderItemRepositoryImpl) PlaceOrderItem(orderItem *model.OrderItem) (*model.OrderItem, error) {
	result, _, err := repository.Client.From(ORDER_ITEM_TABLE).Insert(orderItem, false, "", "representation", "").Single().Execute()
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

	var insertedOrderParam = new(model.OrderItem)
	err = json.Unmarshal(result, insertedOrderParam)
	if err != nil {
		log.Errorf("%s", err.Error())
		return nil, err
	}

	return insertedOrderParam, nil
}
