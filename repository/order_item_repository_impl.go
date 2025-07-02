package repository

import (
	"cashier-api/model"
	"encoding/json"

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
		log.Errorf("Error while place order item with tenant_id: %d, store_id: %d", orderItem.TenantId, orderItem.StoreId)
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
