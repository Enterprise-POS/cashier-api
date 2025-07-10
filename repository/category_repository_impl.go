package repository

import (
	"cashier-api/model"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

type CategoryRepositoryImpl struct {
	Client        *supabase.Client
	CategoryTable string
}

func NewCategoryRepositoryImpl(client *supabase.Client) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{
		Client:        client,
		CategoryTable: "category",
	}
}

const categoryMtmWarehouseTable string = "category_mtm_warehouse"

func (repository *CategoryRepositoryImpl) GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, error) {
	start := page * limit
	// end := start + limit - 1

	/*
		Original query:

		-- Get items base on category (id)
			SELECT
				category.id AS category_id, category.category_name,
				warehouse.item_id, warehouse.item_name, warehouse.stocks
			FROM warehouse
			INNER JOIN category_mtm_warehouse ON category_mtm_warehouse.item_id=warehouse.item_id
			INNER JOIN category ON category.id=category_mtm_warehouse.category_id
			WHERE warehouse.tenant_id=p_tenant_id AND category.id=p_category_id;
	*/

	data := repository.Client.Rpc("get_items_base_on_category", "", map[string]interface{}{
		"p_tenant_id":   tenantId,
		"p_category_id": categoryId,
		"p_limit":       limit,
		"p_offset":      start,
	})

	/*
		Example return

		[
			{"category_id":1,"category_name":"Fruits","item_id":1,"item_name":"Apple","stocks":358},
			{"category_id":1,"category_name":"Fruits","item_id":267,"item_name":"Durian","stocks":10}
		]
	*/
	var results []*model.CategoryWithItem
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		log.Errorf("ERROR ! While unmarshaling data at CategoryRepositoryImpl.GetItemsByCategory. tenantId: %d, categoryId: %d", tenantId, categoryId)
		return nil, err
	}

	return results, nil
}

func (repository *CategoryRepositoryImpl) GetCategoryWithItems(tenantId, page, limit int, doCount bool) ([]*model.CategoryWithItem, int, error) {
	start := page * limit
	// end := start + limit - 1

	/*
		Example join using bare bone supabase method.

		results, _, err := repository.Client.From("category").
			Select("id, category_name, category_mtm_warehouse(warehouse(item_id))", "", false).
			Limit(limit, "category_mtm_warehouse(warehouse(item_id))").
			Range(start, end, "category_mtm_warehouse(warehouse(item_id))").
			Eq("tenant_id", strconv.Itoa(tenantId)).
			Execute()

		Instead will be using Rpc with the same query as above
	*/
	data := repository.Client.Rpc("get_category_with_items", "", map[string]interface{}{
		"p_tenant_id": tenantId,
		"p_limit":     limit,
		"p_offset":    start,
	})
	var results []*model.CategoryWithItem
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		return nil, 0, err
	}

	countResult := 0
	if doCount {
		_, count, err := repository.Client.From("warehouse").Select("item_id", "exact", false).Execute()
		if err != nil {
			return nil, 0, err
		}
		countResult = int(count)
	}

	return results, countResult, nil
}

func (repository *CategoryRepositoryImpl) Get(tenantId, page, limit int) ([]*model.Category, int, error) {
	start := page * limit
	end := start + limit - 1

	var results []*model.Category
	count, err := repository.Client.From("category").
		Select("*", "exact", false).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "").
		ExecuteTo(&results)
	if err != nil {
		return nil, 0, err
	}

	return results, int(count), nil
}

func (repository *CategoryRepositoryImpl) Create(tenantId int, categories []*model.Category) ([]*model.Category, error) {
	if repository.CategoryTable == "" {
		log.Errorf("Fatal Error ! CategoryRepositoryImpl.Create called with empty table. probably didn't use New Fn for create CategoryRepositoryImpl. TenantId: %d", tenantId)
		return nil, fmt.Errorf("CategoryRepositoryImpl.Create called with empty table. probably didn't use New Fn for create CategoryRepositoryImpl. TenantId: %d", tenantId)
	}

	var results []*model.Category
	_, err := repository.Client.From(repository.CategoryTable).
		Insert(categories, false, "", "", "").
		Eq("tenant_id", strconv.Itoa(tenantId)).
		ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repository *CategoryRepositoryImpl) Register(tobeRegisters []*model.CategoryMtmWarehouse) error {
	var results []*model.CategoryMtmWarehouse

	// WARN: inconsistent constant naming
	_, err := repository.Client.From(categoryMtmWarehouseTable).Insert(tobeRegisters, false, "", "", "").ExecuteTo(&results)
	if err != nil {
		return err
	}

	return nil
}

func (repository *CategoryRepositoryImpl) Unregister(toUnregister *model.CategoryMtmWarehouse) error {
	_, count, err := repository.Client.From(categoryMtmWarehouseTable).
		Delete("", "exact").
		Eq("category_id", strconv.Itoa(toUnregister.CategoryId)).
		Eq("item_id", strconv.Itoa(toUnregister.ItemId)).
		Execute()
	if err != nil {
		return err
	}
	if count > 1 {
		log.Errorf("FATAL ERROR multiple categories deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId)
		return errors.New("FATAL ERROR multiple categories deleted")
	}

	if count == 0 {
		log.Warnf("Warning ! Handled error, no data deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId)
		return errors.New("[WARN] No data deleted")
	}

	return nil
}

func (repository *CategoryRepositoryImpl) Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error) {
	/*
		For now, only updating Category.CategoryName is allowed
		- category_name (ok)
		-	id (x)
		- created_at (x)
		- tenant_id (x)
	*/
	tobeUpdatedValue := map[string]interface{}{
		"category_name": tobeChangeCategoryName,
	}

	var updatedCategory *model.Category
	_, err := repository.Client.From(repository.CategoryTable).
		Update(tobeUpdatedValue, "", ""). // Do not use 'exact' for returning parameter
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Eq("id", strconv.Itoa(categoryId)).
		Single().
		ExecuteTo(&updatedCategory)
	if err != nil {
		return nil, err
	}

	return updatedCategory, nil
}

func (repository *CategoryRepositoryImpl) Delete(category *model.Category) error {
	_, count, err := repository.Client.From(repository.CategoryTable).
		Delete("", "exact").
		Eq("tenant_id", strconv.Itoa(category.TenantId)).
		Eq("id", strconv.Itoa(category.Id)).
		Execute()
	if err != nil {
		return err
	}
	if count > 1 {
		log.Errorf("FATAL ERROR multiple categories deleted from categoryId: %d, tenantId: %d", category.Id, category.TenantId)
		return errors.New("FATAL ERROR multiple categories deleted")
	}

	if count == 0 {
		log.Warnf("Warning ! Handled error, no data deleted from categoryId: %d, tenantId: %d", category.Id, category.TenantId)
		return errors.New("[WARN] No data deleted")
	}

	return nil
}
