package service

import "cashier-api/model"

type CategoryService interface {
	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategoryId(tenantId, categoryId, limit, page int) ([]*model.CategoryWithItem, error)

	/*
		Return an all items within category,
		items maybe double return, but different category id is required
	*/
	GetCategoryWithItems(tenantId, page, limit int) ([]*model.CategoryWithItem, int, error)

	/*
		Get the category name only
	*/
	Get(tenantId, page, limit int) ([]*model.Category, int, error)

	/*
		Create new category
	*/
	Create(tenantId int, categories []*model.Category) ([]*model.Category, error)

	/*
		Register warehouse.item into category,
		this will inserting data into category_mtm_warehouse
	*/
	Register(tobeRegisters []*model.CategoryMtmWarehouse) error

	/*
		Unregister, deleting category_mtm_warehouse
		- Only 1 operation allowed for now
	*/
	Unregister(toUnregister *model.CategoryMtmWarehouse) error

	/*
		Update existing category (Not updating category_mtm_warehouse table)
		- only category name allowed to edit
		- only update 1 category
	*/
	Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error)

	/*
		Deleting category; NOT category_mtm_warehouse
	*/
	Delete(category *model.Category) error
}
