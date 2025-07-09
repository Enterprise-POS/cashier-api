package repository

import "cashier-api/model"

type CategoryRepository interface {
	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategoryId(tenantId, categoryId, limit, page int) ([]*model.CategoryWithItem, error)

	/*
		Return an all items within category,
		items maybe double return, but different category id is required
	*/
	GetCategoryWithItems(tenantId, page, limit int, doCount bool) ([]*model.CategoryWithItem, error)

	/*
		Get the category name only
	*/
	Get(tenantId, page, limit int) ([]*model.Category, int, error)
}
