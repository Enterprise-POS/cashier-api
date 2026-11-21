package repository

import (
	common "cashier-api/helper"
	"cashier-api/helper/query"
	"cashier-api/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ExpenseRepositoryImpl struct {
	Client *gorm.DB
}

func NewExpenseRepositoryImpl(client *gorm.DB) ExpenseRepository {
	return &ExpenseRepositoryImpl{Client: client}
}

/*
The FK only checks the store / category exists, not that it is ours.
Without this a member could attach an expense to another tenant's store.
*/
func (repository *ExpenseRepositoryImpl) checkReferences(expense *model.Expense) error {
	if expense.StoreId != nil {
		var count int64
		err := repository.Client.Model(&model.Store{}).
			Where("id = ? AND tenant_id = ?", *expense.StoreId, expense.TenantId).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("Store %d not found", *expense.StoreId)
		}
	}

	if expense.ExpenseCategoryId != nil {
		var count int64
		err := repository.Client.Model(&model.ExpenseCategory{}).
			Where("id = ? AND tenant_id = ?", *expense.ExpenseCategoryId, expense.TenantId).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("Expense category %d not found", *expense.ExpenseCategoryId)
		}
	}

	return nil
}

// Create implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) Create(expense *model.Expense) (*model.Expense, error) {
	if err := repository.checkReferences(expense); err != nil {
		return nil, err
	}

	var user model.User
	if err := repository.Client.Select("name").Take(&user, expense.ResponsibleUserId).Error; err != nil {
		return nil, fmt.Errorf("Responsible user %d not found: %w", expense.ResponsibleUserId, err)
	}
	expense.ResponsibleUserNameSnapshot = user.Name

	if err := repository.Client.Create(expense).Error; err != nil {
		return nil, err
	}

	// Re-read to include the joined category name
	return repository.FindById(expense.TenantId, expense.Id)
}

// withCategoryName select expense row + its category name
func (repository *ExpenseRepositoryImpl) withCategoryName() *gorm.DB {
	return repository.Client.Model(&model.Expense{}).
		Select("expense.*, expense_category.name AS expense_category_name").
		Joins("LEFT JOIN expense_category ON expense_category.id = expense.expense_category_id")
}

// Get implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) Get(
	tenantId int,
	storeId int,
	categoryId int,
	limit int,
	page int,
	filters []*query.QueryFilter,
	dateFilter *query.DateFilter,
) ([]*model.Expense, int, int, error) {
	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Where("expense.tenant_id = ?", tenantId)

		switch {
		case storeId == -1:
			db = db.Where("expense.store_id IS NULL")
		case storeId > 0:
			db = db.Where("expense.store_id = ?", storeId)
		}

		switch {
		case categoryId == -1:
			db = db.Where("expense.expense_category_id IS NULL")
		case categoryId > 0:
			db = db.Where("expense.expense_category_id = ?", categoryId)
		}

		if dateFilter != nil {
			if dateFilter.StartDate != nil && dateFilter.EndDate != nil {
				db = db.Where("expense.expense_date >= ? AND expense.expense_date < ?",
					common.EpochToRFC3339(*dateFilter.StartDate), common.EpochToRFC3339(*dateFilter.EndDate))
			} else if dateFilter.StartDate != nil {
				db = db.Where("expense.expense_date >= ?", common.EpochToRFC3339(*dateFilter.StartDate))
			} else if dateFilter.EndDate != nil {
				db = db.Where("expense.expense_date <= ?", common.EpochToRFC3339(*dateFilter.EndDate))
			}
		}

		return db
	}

	// Count + total amount of every matched row, not only the current page
	var summary struct {
		TotalCount int
		SumAmount  int
	}
	err := applyFilters(repository.Client.Model(&model.Expense{})).
		Select("COUNT(*) AS total_count, COALESCE(SUM(expense.amount), 0) AS sum_amount").
		Scan(&summary).Error
	if err != nil {
		return nil, 0, 0, err
	}

	db := applyFilters(repository.withCategoryName())
	// Columns are whitelisted at the service, prefix because created_at also exists at expense_category
	for _, filter := range filters {
		direction := "DESC"
		if filter.Ascending {
			direction = "ASC"
		}
		db = db.Order(fmt.Sprintf("expense.%s %s", filter.Column, direction))
	}
	// Stable paging when sorting by a non unique column
	db = db.Order("expense.id DESC")

	var results []*model.Expense
	if err := db.Limit(limit).Offset(page * limit).Find(&results).Error; err != nil {
		return nil, 0, 0, err
	}

	return results, summary.TotalCount, summary.SumAmount, nil
}

// FindById implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) FindById(tenantId int, expenseId int) (*model.Expense, error) {
	var expense model.Expense
	err := repository.withCategoryName().
		Where("expense.id = ? AND expense.tenant_id = ?", expenseId, tenantId).
		Take(&expense).Error
	if err != nil {
		return nil, fmt.Errorf("Expense %d not found: %w", expenseId, err)
	}

	return &expense, nil
}

// Edit implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) Edit(expense *model.Expense) (*model.Expense, error) {
	if err := repository.checkReferences(expense); err != nil {
		return nil, err
	}

	now := time.Now()
	expense.UpdatedAt = &now

	// Select is required so nil store_id / expense_category_id are written as NULL
	result := repository.Client.Model(&model.Expense{}).
		Where("id = ? AND tenant_id = ?", expense.Id, expense.TenantId).
		Select("store_id", "expense_category_id", "amount", "description", "expense_date", "payment_type", "updated_at").
		Updates(expense)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("Expense %d not found", expense.Id)
	}

	return repository.FindById(expense.TenantId, expense.Id)
}

// Delete implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) Delete(tenantId int, expenseId int) error {
	result := repository.Client.
		Where("id = ? AND tenant_id = ?", expenseId, tenantId).
		Delete(&model.Expense{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Expense %d not found", expenseId)
	}

	return nil
}

// GetCategories implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) GetCategories(tenantId int) ([]*model.ExpenseCategory, error) {
	var categories []*model.ExpenseCategory
	err := repository.Client.
		Where("tenant_id = ?", tenantId).
		Order("name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// CreateCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) CreateCategory(tenantId int, name string) (*model.ExpenseCategory, error) {
	category := &model.ExpenseCategory{TenantId: tenantId, Name: name}
	if err := repository.Client.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// EditCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) EditCategory(tenantId int, categoryId int, name string) (*model.ExpenseCategory, error) {
	result := repository.Client.Model(&model.ExpenseCategory{}).
		Where("id = ? AND tenant_id = ?", categoryId, tenantId).
		Update("name", name)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("Expense category %d not found", categoryId)
	}

	var category model.ExpenseCategory
	if err := repository.Client.Take(&category, categoryId).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

// DeleteCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryImpl) DeleteCategory(tenantId int, categoryId int) error {
	result := repository.Client.
		Where("id = ? AND tenant_id = ?", categoryId, tenantId).
		Delete(&model.ExpenseCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Expense category %d not found", categoryId)
	}

	return nil
}
