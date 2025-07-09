package repository

import "cashier-api/model"

type CategoryRepository interface {
	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, error)

	/*
		Return an all items within category,
		items maybe double return, but different category id is required
	*/
	GetCategoryWithItems(tenantId, limit, page int) ([]*model.CategoryWithItem, error)

	/*
		Get the category name only
	*/
	Get(tenantId int) []*model.Category
}
