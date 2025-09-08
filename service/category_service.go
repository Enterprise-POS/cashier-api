package service

import "cashier-api/model"

type CategoryService interface {
	/*
		Get the category name only
	*/
	Get(tenantId, page, limit int) ([]*model.Category, int, error)

	/*
		Create new category
	*/
	Create(tenantId int, categories []*model.Category) ([]*model.Category, error)
}
