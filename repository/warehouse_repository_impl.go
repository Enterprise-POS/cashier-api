package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

const WarehouseTable string = "warehouse"

type WarehouseRepositoryImpl struct {
	Client *supabase.Client
}

func NewWarehouseRepositoryImpl(client *supabase.Client) WarehouseRepository {
	return &WarehouseRepositoryImpl{
		Client: client,
	}
}

// GetActiveItem implements WarehouseRepository.
func (warehouse *WarehouseRepositoryImpl) GetActiveItem(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {
	start := page * limit
	end := start + limit - 1

	var itemsList []*model.Item
	query := warehouse.Client.
		From("warehouse").
		Select("*", "exact", false).
		Eq("is_active", "TRUE"). // Only get is_active=TRUE warehouse item
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "").
		Limit(limit, "")

	if nameQuery != "" {
		query = query.Like("item_name", "%"+nameQuery+"%")
	}

	count, err := query.ExecuteTo(&itemsList)
	if err != nil {
		return nil, 0, err
	}

	return itemsList, int(count), nil
}

func (warehouse *WarehouseRepositoryImpl) Get(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {
	start := page * limit
	end := start + limit - 1

	var itemsList []*model.Item
	query := warehouse.Client.
		From("warehouse").
		Select("*", "exact", false).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "").
		Limit(limit, "")

	if nameQuery != "" {
		query = query.Like("item_name", "%"+nameQuery+"%")
	}

	count, err := query.ExecuteTo(&itemsList)
	if err != nil {
		return nil, 0, err
	}

	return itemsList, int(count), nil
}

func (warehouse *WarehouseRepositoryImpl) FindById(itemId int, tenantId int) (*model.Item, error) {
	var item model.Item
	_, err := warehouse.Client.
		From("warehouse").
		Select("*", "exact", false).
		Eq("item_id", strconv.Itoa(itemId)).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Single().ExecuteTo(&item)
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST116)") {
			log.Warnf("Warning ! Handled error, id not found for item with id: %d", itemId)
		} else {
			log.Error("Error fetching warehouse item:", err)
		}
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
		"p_quantity":  quantity,      // int
		"p_item_name": item.ItemName, // string

		// DEPRECATED: category table created
		"p_category": 0, // int

		"p_item_id":   item.ItemId,   // int
		"p_tenant_id": item.TenantId, // int
	})

	if strings.Contains(message, "[ERROR]") {
		return errors.New(message)
	}

	// We don't need to return the item, because we already know before
	return nil
}

func (warehouse *WarehouseRepositoryImpl) SetActivate(tenantId, itemId int, setInto bool) error {
	tobeUpdatedValue := map[string]interface{}{
		"is_active": setInto,
	}

	_, _, err := warehouse.Client.From(WarehouseTable).
		Update(tobeUpdatedValue, "", "").
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Eq("item_id", strconv.Itoa(itemId)).
		Single().
		Execute()
	if err != nil {
		return err
	}

	return nil
}

// FindCompleteById implements WarehouseRepository.
func (warehouse *WarehouseRepositoryImpl) FindCompleteById(itemId int, tenantId int) (*model.CategoryWithItem, error) {
	result := warehouse.Client.Rpc("find_complete_by_id", "", map[string]interface{}{
		"p_item_id":   itemId,   // int
		"p_tenant_id": tenantId, // int
	})

	var items []*model.CategoryWithItem
	err := json.Unmarshal([]byte(result), &items)
	if err != nil {
		if strings.Contains(result, "NO_DATA_FOUND") {
			return nil, errors.New("NO_DATA_FOUND")
		}

		if strings.Contains(result, "CARDINALITY_VIOLATION") {
			return nil, errors.New("CARDINALITY_VIOLATION")
		}

		return nil, err
	}

	return items[0], err
}
