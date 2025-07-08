package repository

import "cashier-api/model"

type CategoryRepository interface {
	/*
		Items = Warehouse Table Item
		Written by category but 'id' needed, not 'category_name'
	*/
	GetItemsByCategory(id int, tenantId int, limit int, page int, doCount bool) ([]*model.CategoryWithItem, int, error)
}
