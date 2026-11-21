package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

type ExpenseService interface {
	/*
		Create new expense. ExpenseDate zero means now, empty PaymentType means CASH
	*/
	Create(expense *model.Expense) (*model.Expense, error)

	/*
		Search expenses, page start from 1
			storeId:    0 = all, -1 = only whole business, > 0 = that store
			categoryId: 0 = all, -1 = uncategorized, > 0 = that category
		Sortable columns: expense_date, amount, created_at (default expense_date DESC)
	*/
	Get(tenantId int, storeId int, categoryId int, limit int, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.Expense, int, int, error)

	FindById(tenantId int, expenseId int) (*model.Expense, error)

	Edit(expense *model.Expense) (*model.Expense, error)

	Delete(tenantId int, expenseId int) error

	GetCategories(tenantId int) ([]*model.ExpenseCategory, error)

	CreateCategory(tenantId int, name string) (*model.ExpenseCategory, error)

	EditCategory(tenantId int, categoryId int, name string) (*model.ExpenseCategory, error)

	DeleteCategory(tenantId int, categoryId int) error
}
