package repository

import (
	"cashier-api/model"
	"strconv"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
)

type CategoryRepositoryImpl struct {
	Client *supabase.Client
}

func (repository *CategoryRepositoryImpl) GetItemsByCategory(id int, tenantId int, limit int, page int, doCount bool) ([]*model.CategoryWithItem, int, error) {
	start := page * limit
	end := start + limit - 1

	/*
		Original query:

		-- Get items base on category (id)
		SELECT * FROM warehouse
		INNER JOIN category_mtm_warehouse ON category_mtm_warehouse.item_id=warehouse.item_id
		INNER JOIN category ON category.id=category_mtm_warehouse.category_id
		WHERE warehouse.tenant_id=1 AND category.id=2;
	*/

	var results []*model.CategoryWithItem

	var query *postgrest.QueryBuilder = repository.Client.From("category_mtm_warehouse")
	var filterBuilder *postgrest.FilterBuilder
	if doCount {
		filterBuilder = query.Select("warehouse(item_id, item_name, tenant_id), category(id, category_name)", "exact", false)
	} else {
		filterBuilder = query.Select("warehouse(item_id, item_name, tenant_id), category(id, category_name)", "", false)
	}

	count, err := filterBuilder.Eq("warehouse.tenant_id", strconv.Itoa(tenantId)).
		Eq("category.id", strconv.Itoa(id)).
		Range(start, end, "").
		ExecuteTo(&results)
	if err != nil {
		return nil, 0, err
	}

	return results, int(count), nil
}

// count, err := repository.Client.From("category").
// 	Select("*", "exact", false).
// 	Eq("id", strconv.Itoa(id)).
// 	Range(start, end, "").
// 	Limit(limit, "").
// 	ExecuteTo(&results)
