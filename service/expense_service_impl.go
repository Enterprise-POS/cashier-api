package service

import (
	"cashier-api/helper/query"
	"cashier-api/helper/validation"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type ExpenseServiceImpl struct {
	Repository repository.ExpenseRepository
}

func NewExpenseServiceImpl(repository repository.ExpenseRepository) ExpenseService {
	return &ExpenseServiceImpl{Repository: repository}
}

// Only these columns can be sorted, they go into ORDER BY
var expenseSortableColumns = map[string]bool{
	"expense_date": true,
	"amount":       true,
	"created_at":   true,
}

const maxExpenseDescription = 500
const maxExpenseCategoryName = 100

var minExpenseDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
var maxExpenseDate = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

// validateExpense checks + normalizes the user editable fields (shared by Create and Edit)
func validateExpense(expense *model.Expense) error {
	if expense.TenantId < 1 {
		return errors.New("Invalid tenant id")
	}

	if expense.Amount < 1 {
		return fmt.Errorf("Amount must be greater than 0. Given amount %d", expense.Amount)
	}

	expense.Description = strings.TrimSpace(expense.Description)
	if utf8.RuneCountInString(expense.Description) > maxExpenseDescription {
		return fmt.Errorf("Description could not more than %d characters", maxExpenseDescription)
	}

	if expense.StoreId != nil && *expense.StoreId < 1 {
		return errors.New("Invalid store id")
	}

	if expense.ExpenseCategoryId != nil && *expense.ExpenseCategoryId < 1 {
		return errors.New("Invalid expense category id")
	}

	if expense.PaymentType == "" {
		expense.PaymentType = model.PaymentTypeCash
	}
	switch expense.PaymentType {
	case model.PaymentTypeCash, model.PaymentTypeQRIS, model.PaymentTypeCard, model.PaymentTypeEWallet, model.PaymentTypeOther:
	default:
		return fmt.Errorf("Invalid payment type: %s", expense.PaymentType)
	}

	if expense.ExpenseDate.IsZero() {
		expense.ExpenseDate = time.Now()
	}
	// Same upper bound as order search (2100). A typo / bad client must not create a year 1970 or 2200 expense
	if expense.ExpenseDate.Before(minExpenseDate) || !expense.ExpenseDate.Before(maxExpenseDate) {
		return fmt.Errorf("Invalid expense date: %s", expense.ExpenseDate.Format("2006-01-02"))
	}

	return nil
}

// Create implements ExpenseService.
func (service *ExpenseServiceImpl) Create(expense *model.Expense) (*model.Expense, error) {
	if expense.ResponsibleUserId < 1 {
		return nil, errors.New("Invalid user id")
	}

	if err := validateExpense(expense); err != nil {
		return nil, err
	}

	return service.Repository.Create(expense)
}

// Get implements ExpenseService.
func (service *ExpenseServiceImpl) Get(
	tenantId int,
	storeId int,
	categoryId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.Expense, int, int, error) {
	if tenantId < 1 {
		return nil, 0, 0, errors.New("Invalid tenant id")
	}

	if limit < 1 || limit > 100 {
		return nil, 0, 0, fmt.Errorf("limit must be between 1 and 100. Given limit %d", limit)
	}

	if page < 1 {
		return nil, 0, 0, fmt.Errorf("page could not less than 1 (page >= 1). Given page %d", page)
	}

	if storeId < -1 {
		return nil, 0, 0, fmt.Errorf("Invalid store id %d", storeId)
	}

	if categoryId < -1 {
		return nil, 0, 0, fmt.Errorf("Invalid expense category id %d", categoryId)
	}

	for _, filter := range filters {
		if filter == nil || !expenseSortableColumns[filter.Column] {
			return nil, 0, 0, errors.New("Invalid sort column, allowed: expense_date, amount, created_at")
		}
	}
	if len(filters) == 0 {
		filters = []*query.QueryFilter{{Column: "expense_date", Ascending: false}}
	}

	if dateFilter != nil {
		if (dateFilter.StartDate != nil && *dateFilter.StartDate < 0) || (dateFilter.EndDate != nil && *dateFilter.EndDate < 0) {
			return nil, 0, 0, errors.New("Invalid date filter timestamp")
		}
		if dateFilter.StartDate != nil && dateFilter.EndDate != nil && *dateFilter.StartDate > *dateFilter.EndDate {
			return nil, 0, 0, fmt.Errorf("Start date (%d) cannot be after end date (%d)", *dateFilter.StartDate, *dateFilter.EndDate)
		}
	}

	// By default SQL start from 0 index
	return service.Repository.Get(tenantId, storeId, categoryId, limit, page-1, filters, dateFilter)
}

// FindById implements ExpenseService.
func (service *ExpenseServiceImpl) FindById(tenantId int, expenseId int) (*model.Expense, error) {
	if tenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	if expenseId < 1 {
		return nil, errors.New("Invalid expense id")
	}

	return service.Repository.FindById(tenantId, expenseId)
}

// Edit implements ExpenseService.
func (service *ExpenseServiceImpl) Edit(expense *model.Expense) (*model.Expense, error) {
	if expense.Id < 1 {
		return nil, errors.New("Invalid expense id")
	}

	if err := validateExpense(expense); err != nil {
		return nil, err
	}

	return service.Repository.Edit(expense)
}

// Delete implements ExpenseService.
func (service *ExpenseServiceImpl) Delete(tenantId int, expenseId int) error {
	if tenantId < 1 {
		return errors.New("Invalid tenant id")
	}

	if expenseId < 1 {
		return errors.New("Invalid expense id")
	}

	return service.Repository.Delete(tenantId, expenseId)
}

// validateCategoryName returns the trimmed name
func validateCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > maxExpenseCategoryName || !validation.ItemNameRegex.MatchString(name) {
		return "", fmt.Errorf("Current expense category name is not allowed: %s", name)
	}

	return name, nil
}

func categoryDuplicateError(err error) error {
	// pgx error text: "... (SQLSTATE 23505)"
	if strings.Contains(err.Error(), "SQLSTATE 23505") {
		return errors.New("Current expense category name already used / duplicate name")
	}

	return err
}

// GetCategories implements ExpenseService.
func (service *ExpenseServiceImpl) GetCategories(tenantId int) ([]*model.ExpenseCategory, error) {
	if tenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	return service.Repository.GetCategories(tenantId)
}

// CreateCategory implements ExpenseService.
func (service *ExpenseServiceImpl) CreateCategory(tenantId int, name string) (*model.ExpenseCategory, error) {
	if tenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	name, err := validateCategoryName(name)
	if err != nil {
		return nil, err
	}

	category, err := service.Repository.CreateCategory(tenantId, name)
	if err != nil {
		return nil, categoryDuplicateError(err)
	}

	return category, nil
}

// EditCategory implements ExpenseService.
func (service *ExpenseServiceImpl) EditCategory(tenantId int, categoryId int, name string) (*model.ExpenseCategory, error) {
	if tenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	if categoryId < 1 {
		return nil, errors.New("Invalid expense category id")
	}

	name, err := validateCategoryName(name)
	if err != nil {
		return nil, err
	}

	category, err := service.Repository.EditCategory(tenantId, categoryId, name)
	if err != nil {
		return nil, categoryDuplicateError(err)
	}

	return category, nil
}

// DeleteCategory implements ExpenseService.
func (service *ExpenseServiceImpl) DeleteCategory(tenantId int, categoryId int) error {
	if tenantId < 1 {
		return errors.New("Invalid tenant id")
	}

	if categoryId < 1 {
		return errors.New("Invalid expense category id")
	}

	return service.Repository.DeleteCategory(tenantId, categoryId)
}
