package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/supabase-community/supabase-go"
)

type WarehouseRepositoryImpl struct {
	Client *supabase.Client
}

func (warehouse *WarehouseRepositoryImpl) FindById(itemId int, tenantId int) *model.Item {
	data, _, err := warehouse.Client.
		From("warehouse").
		Select("*", "exact", false).
		Eq("item_id", strconv.Itoa(itemId)).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Single().Execute()
	if err != nil {
		fmt.Println("Error fetching warehouse item:", err)
		return nil
	}

	// Marshal response to your model
	var item model.Item
	err = json.Unmarshal(data, &item)
	if err != nil {
		fmt.Println("Failed to unmarshal Supabase response:", err)
		return nil
	}

	return &item
}

func (warehouse *WarehouseRepositoryImpl) CreateItem(item *model.Item) (*model.Item, error) {
	result, _, err := warehouse.Client.From("warehouse").Insert(item, false, "", "representation", "").Single().Execute()
	if err != nil {
		return nil, err
	}

	var data = new(model.Item)
	err = json.Unmarshal(result, data)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("[ERROR] Fatal ! Could not Unmarshal item at CreateItem")
	}

	return data, nil
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
		return errors.New(strings.Replace(message, "[ERROR] ", "", 1))
	}

	// We don't need to return the item, because we already know before
	return nil
}
