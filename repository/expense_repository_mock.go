package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type ExpenseRepositoryMock struct {
	Mock *mock.Mock
}

func NewExpenseRepositoryMock(mock *mock.Mock) ExpenseRepository {
	return &ExpenseRepositoryMock{Mock: mock}
}

// Create implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) Create(expense *model.Expense) (*model.Expense, error) {
	args := repository.Mock.Called(expense)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Expense), args.Error(1)
}

// Get implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) Get(tenantId int, storeId int, categoryId int, limit int, page int, filters []*query.QueryFilter, dateFilter *query.DateFilter) ([]*model.Expense, int, int, error) {
	args := repository.Mock.Called(tenantId, storeId, categoryId, limit, page, filters, dateFilter)
	if args.Get(0) == nil {
		return nil, 0, 0, args.Error(3)
	}

	return args.Get(0).([]*model.Expense), args.Int(1), args.Int(2), args.Error(3)
}

// FindById implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) FindById(tenantId int, expenseId int) (*model.Expense, error) {
	args := repository.Mock.Called(tenantId, expenseId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Expense), args.Error(1)
}

// Edit implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) Edit(expense *model.Expense) (*model.Expense, error) {
	args := repository.Mock.Called(expense)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Expense), args.Error(1)
}

// Delete implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) Delete(tenantId int, expenseId int) error {
	return repository.Mock.Called(tenantId, expenseId).Error(0)
}

// GetCategories implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) GetCategories(tenantId int) ([]*model.ExpenseCategory, error) {
	args := repository.Mock.Called(tenantId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.ExpenseCategory), args.Error(1)
}

// CreateCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) CreateCategory(tenantId int, name string) (*model.ExpenseCategory, error) {
	args := repository.Mock.Called(tenantId, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.ExpenseCategory), args.Error(1)
}

// EditCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) EditCategory(tenantId int, categoryId int, name string) (*model.ExpenseCategory, error) {
	args := repository.Mock.Called(tenantId, categoryId, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.ExpenseCategory), args.Error(1)
}

// DeleteCategory implements ExpenseRepository.
func (repository *ExpenseRepositoryMock) DeleteCategory(tenantId int, categoryId int) error {
	return repository.Mock.Called(tenantId, categoryId).Error(0)
}
