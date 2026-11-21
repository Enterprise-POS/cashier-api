package repository

import (
	"cashier-api/helper/client"
	"cashier-api/helper/query"
	"cashier-api/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Every subtest runs in its own transaction and rolls back, nothing is left at DB
func TestExpenseRepository(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	type fixture struct {
		tx       *gorm.DB
		repo     ExpenseRepository
		tenantId int
		storeId  int
		user     *model.User
	}
	setup := func(t *testing.T) *fixture {
		tx := gormClient.Begin()
		t.Cleanup(func() { tx.Rollback() })

		tenantId, storeId, user := seedOrderItemTestDependencies(t, tx, "expense_test@example.com", "Expense")
		return &fixture{tx: tx, repo: NewExpenseRepositoryImpl(tx), tenantId: tenantId, storeId: storeId, user: user}
	}
	newExpense := func(f *fixture, amount int, storeId *int, categoryId *int, date time.Time) *model.Expense {
		return &model.Expense{
			TenantId:          f.tenantId,
			StoreId:           storeId,
			ExpenseCategoryId: categoryId,
			Amount:            amount,
			Description:       "Test expense",
			ExpenseDate:       date,
			PaymentType:       model.PaymentTypeCash,
			ResponsibleUserId: f.user.Id,
		}
	}

	t.Run("Create", func(t *testing.T) {
		t.Run("NormalCreate", func(t *testing.T) {
			f := setup(t)
			category, err := f.repo.CreateCategory(f.tenantId, "Electricity")
			require.NoError(t, err)

			created, err := f.repo.Create(newExpense(f, 150_000, &f.storeId, &category.Id, time.Now()))
			require.NoError(t, err)
			assert.NotZero(t, created.Id)
			assert.Equal(t, 150_000, created.Amount)
			assert.Equal(t, f.user.Name, created.ResponsibleUserNameSnapshot)
			require.NotNil(t, created.ExpenseCategoryName)
			assert.Equal(t, "Electricity", *created.ExpenseCategoryName)
			require.NotNil(t, created.StoreId)
			assert.Equal(t, f.storeId, *created.StoreId)
		})

		t.Run("WholeBusinessUncategorized", func(t *testing.T) {
			f := setup(t)
			created, err := f.repo.Create(newExpense(f, 50_000, nil, nil, time.Now()))
			require.NoError(t, err)
			assert.Nil(t, created.StoreId)
			assert.Nil(t, created.ExpenseCategoryId)
			assert.Nil(t, created.ExpenseCategoryName)
		})

		t.Run("OtherTenantStoreAndCategoryRejected", func(t *testing.T) {
			f := setup(t)
			otherTenantId, otherStoreId, _ := seedOrderItemTestDependencies(t, f.tx, "expense_test_other@example.com", "Expense Other")
			otherCategory, err := f.repo.CreateCategory(otherTenantId, "Other Tenant Category")
			require.NoError(t, err)

			created, err := f.repo.Create(newExpense(f, 1_000, &otherStoreId, nil, time.Now()))
			assert.Error(t, err)
			assert.Nil(t, created)

			created, err = f.repo.Create(newExpense(f, 1_000, nil, &otherCategory.Id, time.Now()))
			assert.Error(t, err)
			assert.Nil(t, created)
		})
	})

	t.Run("Get", func(t *testing.T) {
		f := setup(t)
		category, err := f.repo.CreateCategory(f.tenantId, "Rent")
		require.NoError(t, err)

		base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		_, err = f.repo.Create(newExpense(f, 100, &f.storeId, &category.Id, base)) // store, categorized
		require.NoError(t, err)
		_, err = f.repo.Create(newExpense(f, 300, &f.storeId, nil, base.AddDate(0, 0, 1))) // store, uncategorized
		require.NoError(t, err)
		_, err = f.repo.Create(newExpense(f, 200, nil, &category.Id, base.AddDate(0, 0, 2))) // whole business
		require.NoError(t, err)
		_, err = f.repo.Create(newExpense(f, 999, nil, nil, base.AddDate(0, 1, 0))) // next month
		require.NoError(t, err)

		t.Run("AllWithCount", func(t *testing.T) {
			results, count, sum, err := f.repo.Get(f.tenantId, 0, 0, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 4, count)
			assert.Equal(t, 100+300+200+999, sum)
			assert.Len(t, results, 4)
		})

		t.Run("StoreFilters", func(t *testing.T) {
			_, count, _, err := f.repo.Get(f.tenantId, f.storeId, 0, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 2, count)

			_, count, _, err = f.repo.Get(f.tenantId, -1, 0, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 2, count)
		})

		t.Run("CategoryFilters", func(t *testing.T) {
			_, count, _, err := f.repo.Get(f.tenantId, 0, category.Id, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 2, count)

			_, count, _, err = f.repo.Get(f.tenantId, 0, -1, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 2, count)
		})

		t.Run("DateRangeAndSort", func(t *testing.T) {
			start := base.Unix()
			end := base.AddDate(0, 1, 0).Unix() // exclusive, next month row is excluded
			results, count, _, err := f.repo.Get(f.tenantId, 0, 0, 10, 0,
				[]*query.QueryFilter{{Column: "amount", Ascending: true}},
				&query.DateFilter{StartDate: &start, EndDate: &end})
			require.NoError(t, err)
			assert.Equal(t, 3, count)
			require.Len(t, results, 3)
			assert.Equal(t, []int{100, 200, 300}, []int{results[0].Amount, results[1].Amount, results[2].Amount})
		})

		t.Run("Pagination", func(t *testing.T) {
			results, count, sum, err := f.repo.Get(f.tenantId, 0, 0, 3, 1, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 4, count)
			assert.Equal(t, 100+300+200+999, sum) // whole result, not only this page
			assert.Len(t, results, 1)
		})

		t.Run("OtherTenantSeesNothing", func(t *testing.T) {
			otherTenantId, _, _ := seedOrderItemTestDependencies(t, f.tx, "expense_test_other@example.com", "Expense Other")
			_, count, _, err := f.repo.Get(otherTenantId, 0, 0, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 0, count)
		})
	})

	t.Run("Edit", func(t *testing.T) {
		t.Run("NormalEdit", func(t *testing.T) {
			f := setup(t)
			created, err := f.repo.Create(newExpense(f, 1_000, &f.storeId, nil, time.Now()))
			require.NoError(t, err)

			// Move to whole business, new amount
			created.StoreId = nil
			created.Amount = 2_500
			created.Description = "Edited"
			edited, err := f.repo.Edit(created)
			require.NoError(t, err)
			assert.Nil(t, edited.StoreId)
			assert.Equal(t, 2_500, edited.Amount)
			assert.Equal(t, "Edited", edited.Description)
			assert.Equal(t, f.user.Name, edited.ResponsibleUserNameSnapshot)
		})

		t.Run("OtherTenantNotFound", func(t *testing.T) {
			f := setup(t)
			created, err := f.repo.Create(newExpense(f, 1_000, nil, nil, time.Now()))
			require.NoError(t, err)

			otherTenantId, _, _ := seedOrderItemTestDependencies(t, f.tx, "expense_test_other@example.com", "Expense Other")
			created.TenantId = otherTenantId
			edited, err := f.repo.Edit(created)
			assert.Error(t, err)
			assert.Nil(t, edited)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("SoftDelete", func(t *testing.T) {
			f := setup(t)
			created, err := f.repo.Create(newExpense(f, 1_000, nil, nil, time.Now()))
			require.NoError(t, err)

			require.NoError(t, f.repo.Delete(f.tenantId, created.Id))

			var deleted model.Expense
			require.NoError(t, f.tx.Unscoped().First(&deleted, created.Id).Error)
			assert.True(t, deleted.DeletedAt.Valid)

			found, err := f.repo.FindById(f.tenantId, created.Id)
			assert.Error(t, err)
			assert.Nil(t, found)

			_, count, sum, err := f.repo.Get(f.tenantId, 0, 0, 10, 0, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, 0, count)
			assert.Equal(t, 0, sum)

			// Deleting twice is not found
			assert.Error(t, f.repo.Delete(f.tenantId, created.Id))
		})

		t.Run("OtherTenantNotFound", func(t *testing.T) {
			f := setup(t)
			created, err := f.repo.Create(newExpense(f, 1_000, nil, nil, time.Now()))
			require.NoError(t, err)

			otherTenantId, _, _ := seedOrderItemTestDependencies(t, f.tx, "expense_test_other@example.com", "Expense Other")
			assert.Error(t, f.repo.Delete(otherTenantId, created.Id))
		})
	})

	t.Run("Category", func(t *testing.T) {
		t.Run("CreateGetEdit", func(t *testing.T) {
			f := setup(t)
			_, err := f.repo.CreateCategory(f.tenantId, "Salary")
			require.NoError(t, err)
			created, err := f.repo.CreateCategory(f.tenantId, "Internet")
			require.NoError(t, err)

			categories, err := f.repo.GetCategories(f.tenantId)
			require.NoError(t, err)
			require.Len(t, categories, 2)
			assert.Equal(t, "Internet", categories[0].Name) // sorted by name

			edited, err := f.repo.EditCategory(f.tenantId, created.Id, "Internet & Phone")
			require.NoError(t, err)
			assert.Equal(t, "Internet & Phone", edited.Name)

			otherTenantId, _, _ := seedOrderItemTestDependencies(t, f.tx, "expense_test_other@example.com", "Expense Other")
			_, err = f.repo.EditCategory(otherTenantId, created.Id, "Hijack")
			assert.Error(t, err)
			assert.Error(t, f.repo.DeleteCategory(otherTenantId, created.Id))
		})

		t.Run("DeleteMakesExpenseUncategorized", func(t *testing.T) {
			f := setup(t)
			category, err := f.repo.CreateCategory(f.tenantId, "Temporary")
			require.NoError(t, err)
			created, err := f.repo.Create(newExpense(f, 1_000, nil, &category.Id, time.Now()))
			require.NoError(t, err)

			require.NoError(t, f.repo.DeleteCategory(f.tenantId, category.Id))

			found, err := f.repo.FindById(f.tenantId, created.Id)
			require.NoError(t, err)
			assert.Nil(t, found.ExpenseCategoryId)
			assert.Nil(t, found.ExpenseCategoryName)
		})

		t.Run("Duplicate", func(t *testing.T) {
			f := setup(t)
			_, err := f.repo.CreateCategory(f.tenantId, "Duplicate")
			require.NoError(t, err)

			// Last statement of this transaction, a failed insert aborts the transaction
			_, err = f.repo.CreateCategory(f.tenantId, "Duplicate")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "23505")
		})
	})

	t.Run("SalesReportExpenses", func(t *testing.T) {
		f := setup(t)
		base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		_, err := f.repo.Create(newExpense(f, 100, &f.storeId, nil, base))
		require.NoError(t, err)
		_, err = f.repo.Create(newExpense(f, 200, nil, nil, base)) // whole business
		require.NoError(t, err)
		_, err = f.repo.Create(newExpense(f, 400, nil, nil, base.AddDate(0, 1, 0))) // out of range
		require.NoError(t, err)
		deleted, err := f.repo.Create(newExpense(f, 800, nil, nil, base))
		require.NoError(t, err)
		require.NoError(t, f.repo.Delete(f.tenantId, deleted.Id))

		start := base.Unix()
		end := base.AddDate(0, 1, 0).Unix()
		dateFilter := &query.DateFilter{StartDate: &start, EndDate: &end}
		orderItemRepo := NewOrderItemRepositoryImpl(f.tx)

		allStores, err := orderItemRepo.GetSalesReport(f.tenantId, 0, dateFilter)
		require.NoError(t, err)
		assert.Equal(t, 300, allStores.SumExpenses) // store + whole business, deleted excluded
		assert.Equal(t, allStores.SumProfit-300, allStores.NetProfit)

		oneStore, err := orderItemRepo.GetSalesReport(f.tenantId, f.storeId, dateFilter)
		require.NoError(t, err)
		assert.Equal(t, 100, oneStore.SumExpenses) // whole business expense only at "All stores"
	})
}
