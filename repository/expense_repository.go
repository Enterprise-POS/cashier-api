package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
)

/*
Expense: manual business spending, plus the tenant's own expense categories.
Every method is scoped by tenantId, never trust an id without its tenant.
*/
type ExpenseRepository interface {
	/*
		Create new expense. ResponsibleUserNameSnapshot is filled from ResponsibleUserId.
		StoreId / ExpenseCategoryId (when not nil) must belong to the same tenant
	*/
	Create(expense *model.Expense) (*model.Expense, error)

	/*
		Search expenses, 2nd return is the count and 3rd the SUM(amount) of all matched data (not only this page)
			storeId:    0 = all, -1 = only whole business (store_id IS NULL), > 0 = that store
			categoryId: 0 = all, -1 = uncategorized, > 0 = that category
			page:       start from 0
		Date filter is applied to expense_date
	*/
	Get(tenantId int, storeId int, categoryId int, limit int, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.Expense, int, int, error)

	/*
		Find 1 expense
	*/
	FindById(tenantId int, expenseId int) (*model.Expense, error)

	/*
		Edit store_id, expense_category_id, amount, description, expense_date, payment_type.
		Same tenant rule as Create
	*/
	Edit(expense *model.Expense) (*model.Expense, error)

	/*
		Soft delete
	*/
	Delete(tenantId int, expenseId int) error

	/*
		All categories of current tenant, sorted by name
	*/
	GetCategories(tenantId int) ([]*model.ExpenseCategory, error)

	CreateCategory(tenantId int, name string) (*model.ExpenseCategory, error)

	EditCategory(tenantId int, categoryId int, name string) (*model.ExpenseCategory, error)

	/*
		Hard delete, the expenses become uncategorized (ON DELETE SET NULL)
	*/
	DeleteCategory(tenantId int, categoryId int) error
}
