package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpenseServiceImpl(t *testing.T) {
	expenseRepository := repository.NewExpenseRepositoryMock(&mock.Mock{}).(*repository.ExpenseRepositoryMock)
	expenseService := NewExpenseServiceImpl(expenseRepository)

	tenantId := 1
	intPtr := func(v int) *int { return &v }
	validExpense := func() *model.Expense {
		return &model.Expense{TenantId: tenantId, ResponsibleUserId: 1, Amount: 10_000, Description: "  Ice  "}
	}

	t.Run("Create", func(t *testing.T) {
		t.Run("DefaultsApplied", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			expenseRepository.Mock.On("Create", mock.Anything).Return(&model.Expense{Id: 1}, nil)

			expense := validExpense()
			created, err := expenseService.Create(expense)
			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, model.PaymentTypeCash, expense.PaymentType)
			assert.False(t, expense.ExpenseDate.IsZero())
			assert.Equal(t, "Ice", expense.Description)
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{} // repository must never be reached

			cases := map[string]func(e *model.Expense){
				"ZeroAmount":         func(e *model.Expense) { e.Amount = 0 },
				"NegativeAmount":     func(e *model.Expense) { e.Amount = -5 },
				"NoTenant":           func(e *model.Expense) { e.TenantId = 0 },
				"NoUser":             func(e *model.Expense) { e.ResponsibleUserId = 0 },
				"ZeroStore":          func(e *model.Expense) { e.StoreId = intPtr(0) },
				"ZeroCategory":       func(e *model.Expense) { e.ExpenseCategoryId = intPtr(0) },
				"UnknownPaymentType": func(e *model.Expense) { e.PaymentType = "BITCOIN" },
				"LongDescription":    func(e *model.Expense) { e.Description = strings.Repeat("a", 501) },
				"Date1970":           func(e *model.Expense) { e.ExpenseDate = time.Unix(0, 0) },
				"Date2200":           func(e *model.Expense) { e.ExpenseDate = time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC) },
			}
			for name, breakIt := range cases {
				t.Run(name, func(t *testing.T) {
					expense := validExpense()
					breakIt(expense)
					created, err := expenseService.Create(expense)
					assert.Error(t, err)
					assert.Nil(t, created)
				})
			}
			expenseRepository.Mock.AssertNotCalled(t, "Create", mock.Anything)
		})
	})

	t.Run("Edit", func(t *testing.T) {
		expenseRepository.Mock = &mock.Mock{}
		expense := validExpense()
		edited, err := expenseService.Edit(expense) // no id
		assert.Error(t, err)
		assert.Nil(t, edited)

		expense.Id = 7
		expenseRepository.Mock.On("Edit", expense).Return(expense, nil)
		edited, err = expenseService.Edit(expense)
		assert.NoError(t, err)
		assert.Equal(t, 7, edited.Id)
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("DefaultSortAndZeroBasedPage", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			defaultSort := []*query.QueryFilter{{Column: "expense_date", Ascending: false}}
			expenseRepository.Mock.On("Get", tenantId, -1, -1, 10, 1, defaultSort, (*query.DateFilter)(nil)).
				Return([]*model.Expense{}, 0, 0, nil)

			_, _, _, err := expenseService.Get(tenantId, -1, -1, 10, 2, nil, nil)
			assert.NoError(t, err)
			expenseRepository.Mock.AssertExpectations(t)
		})

		t.Run("RejectUnknownSortColumn", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			for _, column := range []string{"", "tenant_id", "amount; DROP TABLE expense", "name"} {
				_, _, _, err := expenseService.Get(tenantId, 0, 0, 10, 1, []*query.QueryFilter{{Column: column}}, nil)
				assert.Error(t, err, column)
			}
			expenseRepository.Mock.AssertNotCalled(t, "Get")
		})

		t.Run("InvalidParameter", func(t *testing.T) {
			start, end := int64(200), int64(100)
			negative := int64(-1)
			_, _, _, err := expenseService.Get(0, 0, 0, 10, 1, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, 0, 0, 1, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, 0, 101, 1, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, 0, 10, 0, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, -2, 0, 10, 1, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, -2, 10, 1, nil, nil)
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, 0, 10, 1, nil, &query.DateFilter{StartDate: &start, EndDate: &end})
			assert.Error(t, err)
			_, _, _, err = expenseService.Get(tenantId, 0, 0, 10, 1, nil, &query.DateFilter{StartDate: &negative})
			assert.Error(t, err)
		})
	})

	t.Run("Category", func(t *testing.T) {
		t.Run("NameTrimmed", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			expenseRepository.Mock.On("CreateCategory", tenantId, "Rent").Return(&model.ExpenseCategory{Id: 1, Name: "Rent"}, nil)

			created, err := expenseService.CreateCategory(tenantId, "  Rent  ")
			assert.NoError(t, err)
			assert.Equal(t, "Rent", created.Name)
		})

		t.Run("InvalidName", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			for _, name := range []string{"", "   ", "-Starts with symbol", strings.Repeat("a", 101), "Bad\nName"} {
				created, err := expenseService.CreateCategory(tenantId, name)
				assert.Error(t, err, name)
				assert.Nil(t, created)
			}
			expenseRepository.Mock.AssertNotCalled(t, "CreateCategory")
		})

		t.Run("DuplicateNameFriendlyError", func(t *testing.T) {
			expenseRepository.Mock = &mock.Mock{}
			expenseRepository.Mock.On("EditCategory", tenantId, 3, "Rent").
				Return(nil, errors.New(`ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)`))

			edited, err := expenseService.EditCategory(tenantId, 3, "Rent")
			assert.Nil(t, edited)
			assert.EqualError(t, err, "Current expense category name already used / duplicate name")
		})

		t.Run("InvalidId", func(t *testing.T) {
			_, err := expenseService.EditCategory(tenantId, 0, "Rent")
			assert.Error(t, err)
			assert.Error(t, expenseService.DeleteCategory(tenantId, 0))
			assert.Error(t, expenseService.Delete(tenantId, 0))
		})
	})
}
