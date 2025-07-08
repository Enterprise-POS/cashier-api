package repository

import "cashier-api/model"

type CategoryRepository interface {
	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategory(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, error)
}
