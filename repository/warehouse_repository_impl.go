package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

type WarehouseRepositoryImpl struct {
	Client *supabase.Client
}

func (warehouse *WarehouseRepositoryImpl) Get(tenantId int, limit int, page int) ([]*model.Item, int, error) {
	start := page * limit
	end := start + limit - 1

	data, count, err := warehouse.Client.
		From("warehouse").

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

	var itemsList []*model.Item
	err = json.Unmarshal(data, &itemsList)
	if err != nil {
		return nil, 0, err
	}

	return itemsList, int(count), nil
}

func (warehouse *WarehouseRepositoryImpl) FindById(itemId int, tenantId int) (*model.Item, error) {
	data, _, err := warehouse.Client.
		From("warehouse").
		Select("*", "exact", false).
		Eq("item_id", strconv.Itoa(itemId)).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Single().Execute()
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST116)") {
			log.Warnf("Warning ! Handled error, id not found for item with id: %d", itemId)
		} else {
			log.Error("Error fetching warehouse item:", err)
		}
		return nil, err
	}

	var item model.Item
	err = json.Unmarshal(data, &item)
	if err != nil {
		fmt.Println("Failed to unmarshal Supabase response:", err)
		return nil, err
	}

	return &item, err
}

func (warehouse *WarehouseRepositoryImpl) CreateItem(items []*model.Item) ([]*model.Item, error) {
	result, _, err := warehouse.Client.From("warehouse").Insert(items, false, "", "representation", "").Execute()
	if err != nil {
		return nil, err
	}

	// fmt.Println("message ->", reflect.TypeOf(string(result)).Name())

	var itemsList []*model.Item
	err = json.Unmarshal(result, &itemsList)
	if err != nil {
		return nil, err
	}

	return itemsList, nil
}

func (warehouse *WarehouseRepositoryImpl) Edit(quantity int, item *model.Item) error {
	message := warehouse.Client.Rpc("edit_warehouse_item", "", map[string]interface{}{
		"p_quantity":  quantity,
		"p_item_name": item.ItemName,
		"p_category":  nil, // TODO: implement this
		"p_item_id":   item.ItemId,
		"p_tenant_id": item.TenantId,
	})

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	// We don't need to return the item, because we already know before
	return nil
}
