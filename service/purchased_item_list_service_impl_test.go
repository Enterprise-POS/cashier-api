package service

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPurchasedItemListService(t *testing.T) {
	// newService wires a fresh mock repository into the service under test
	// so each subtest gets isolated expectations.
	newService := func() (PurchasedItemService, *mock.Mock) {
		m := new(mock.Mock)
		repo := repository.NewPurchasedItemRepositoryMock(m)
		return NewPurchasedItemServiceImpl(repo), m
	}

	t.Run("PurchasedItemListLogs", func(t *testing.T) {
		t.Run("Validation", func(t *testing.T) {
			t.Run("TenantIdZero", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				rows, total, err := purchasedItemListService.PurchasedItemListLogs(0, 0, nil, 10, 1, nil, nil)
				assert.Nil(t, rows)
				assert.Zero(t, total)
				assert.EqualError(t, err, "Tenant id is Required !")
			})

			t.Run("TenantIdNegative", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(-1, 0, nil, 10, 1, nil, nil)
				assert.EqualError(t, err, "Tenant id is Required !")
			})

			t.Run("NegativeStoreId", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, -1, nil, 10, 1, nil, nil)
				assert.EqualError(t, err, "Given store id value is not allowed. storeId: -1")
			})

			t.Run("TooManyItemIds", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				ids := make([]int, 101)
				for i := range ids {
					ids[i] = i + 1
				}

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, ids, 10, 1, nil, nil)
				assert.EqualError(t, err, "Too many item_id values (max 100)")
			})

			t.Run("ExactlyMaxItemIdsIsAllowed", func(t *testing.T) {
				purchasedItemListService, m := newService()

				ids := make([]int, 100)
				for i := range ids {
					ids[i] = i + 1
				}

				m.On(
					"PurchasedItemListLogs",
					1, 0, ids, 10, 0,
					(*query.DateFilter)(nil),
					[]query.QueryFilter(nil),
				).Return([]*model.PurchasedItem{}, 0, nil)

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, ids, 10, 1, nil, nil)
				assert.NoError(t, err)
				m.AssertExpectations(t)
			})

			t.Run("InvalidItemIdValue", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, []int{5, 0}, 10, 1, nil, nil)
				assert.EqualError(t, err, "invalid item_id: 0")
			})

			t.Run("NegativeLimit", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, -1, 1, nil, nil)
				assert.EqualError(t, err, "Limit could not less then 1 (limit >= 1). Given limit -1")
			})

			t.Run("LimitTooLarge", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 1001, 1, nil, nil)
				assert.EqualError(t, err, "Limit could not greater then 1000. Given limit 1001")
			})

			t.Run("LimitAtUpperBoundIsAllowed", func(t *testing.T) {
				purchasedItemListService, m := newService()

				m.On(
					"PurchasedItemListLogs",
					1, 0, []int(nil), 1000, 0,
					(*query.DateFilter)(nil),
					[]query.QueryFilter(nil),
				).Return([]*model.PurchasedItem{}, 0, nil)

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 1000, 1, nil, nil)
				assert.NoError(t, err)
				m.AssertExpectations(t)
			})

			// NOTE: the service checks "if page < 1 { return error }" BEFORE the
			// "else if page == 0" default branch, so page == 0 always errors and
			// never falls through to the default. Unlike `limit`, there is no
			// way to get an implicit default page via 0 — callers must pass >= 1.
			t.Run("PageZeroIsRejectedNotDefaulted", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 0, nil, nil)
				assert.EqualError(t, err, "page could not less then 1 (page >= 1). Given page 0")
			})

			t.Run("NegativePage", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, -1, nil, nil)
				assert.EqualError(t, err, "page could not less then 1 (page >= 1). Given page -1")
			})
		})

		t.Run("DateFilterValidation", func(t *testing.T) {
			t.Run("StartAfterEnd", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				start, end := int64(200), int64(100)
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, &query.DateFilter{
					StartDate: &start,
					EndDate:   &end,
				}, nil)
				assert.EqualError(t, err, "Start date (200) cannot be after end date (100)")
			})

			t.Run("NegativeStartDate", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				start := int64(-1)
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, &query.DateFilter{
					StartDate: &start,
				}, nil)
				assert.EqualError(t, err, "Invalid start date timestamp: -1")
			})

			t.Run("NegativeEndDate", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				end := int64(-1)
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, &query.DateFilter{
					EndDate: &end,
				}, nil)
				assert.EqualError(t, err, "Invalid emd date timestamp: -1")
			})

			t.Run("StartDateTooFarInFuture", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				start := int64(4102444801) // one second past the 2100-01-01 cap
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, &query.DateFilter{
					StartDate: &start,
				}, nil)
				assert.EqualError(t, err, "Start date is too far in the future: 4102444801")
			})

			t.Run("EndDateTooFarInFuture", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				end := int64(4102444801)
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, &query.DateFilter{
					EndDate: &end,
				}, nil)
				assert.EqualError(t, err, "End date is too far in the future: 4102444801")
			})

			t.Run("MaxTimestampIsAllowed", func(t *testing.T) {
				purchasedItemListService, m := newService()

				maxTimestamp := int64(4102444800)
				dateFilter := &query.DateFilter{StartDate: &maxTimestamp, EndDate: &maxTimestamp}

				m.On(
					"PurchasedItemListLogs",
					1, 0, []int(nil), 10, 0,
					dateFilter,
					[]query.QueryFilter(nil),
				).Return([]*model.PurchasedItem{}, 0, nil)

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, dateFilter, nil)
				assert.NoError(t, err)
				m.AssertExpectations(t)
			})
		})

		t.Run("FilterColumnValidation", func(t *testing.T) {
			t.Run("InvalidColumn", func(t *testing.T) {
				purchasedItemListService, _ := newService()

				filters := []query.QueryFilter{{Column: "not_a_real_column", Ascending: true}}
				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, nil, filters)
				assert.EqualError(t, err, "Invalid filter column: not_a_real_column")
			})

			t.Run("ValidColumnsPassThrough", func(t *testing.T) {
				purchasedItemListService, m := newService()

				filters := []query.QueryFilter{
					{Column: query.TotalAmountColumn, Ascending: false},
				}

				m.On(
					"PurchasedItemListLogs",
					1, 0, []int(nil), 10, 0,
					(*query.DateFilter)(nil),
					filters,
				).Return([]*model.PurchasedItem{}, 0, nil)

				_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, nil, filters)
				assert.NoError(t, err)
				m.AssertExpectations(t)
			})
		})

		t.Run("DelegatesToRepository", func(t *testing.T) {
			purchasedItemListService, m := newService()

			expected := []*model.PurchasedItem{{Id: 1}, {Id: 2}}

			// page=1 from the caller should reach the repository as page-1=0.
			m.On(
				"PurchasedItemListLogs",
				1, 5, []int{10, 20}, 25, 0,
				(*query.DateFilter)(nil),
				[]query.QueryFilter(nil),
			).Return(expected, 2, nil)

			rows, total, err := purchasedItemListService.PurchasedItemListLogs(1, 5, []int{10, 20}, 25, 1, nil, nil)
			assert.NoError(t, err)
			assert.Equal(t, expected, rows)
			assert.Equal(t, 2, total)
			m.AssertExpectations(t)
		})

		t.Run("SecondPageIsZeroIndexedForRepository", func(t *testing.T) {
			purchasedItemListService, m := newService()

			m.On(
				"PurchasedItemListLogs",
				1, 0, []int(nil), 10, 1, // page=2 in -> page-1=1 out
				(*query.DateFilter)(nil),
				[]query.QueryFilter(nil),
			).Return([]*model.PurchasedItem{}, 0, nil)

			_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 2, nil, nil)
			assert.NoError(t, err)
			m.AssertExpectations(t)
		})

		t.Run("DefaultsLimitWhenZero", func(t *testing.T) {
			purchasedItemListService, m := newService()

			m.On(
				"PurchasedItemListLogs",
				1, 0, []int(nil), 10, 0, // limit=0 in -> defaulted to 10 out
				(*query.DateFilter)(nil),
				[]query.QueryFilter(nil),
			).Return([]*model.PurchasedItem{}, 0, nil)

			_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 0, 1, nil, nil)
			assert.NoError(t, err)
			m.AssertExpectations(t)
		})

		t.Run("EmptyItemIdsIsAllowed", func(t *testing.T) {
			// The "at least one item_id" check is commented out in the service,
			// so an empty/nil itemIds slice must pass validation and reach the repo.
			purchasedItemListService, m := newService()

			m.On(
				"PurchasedItemListLogs",
				1, 0, []int(nil), 10, 0,
				(*query.DateFilter)(nil),
				[]query.QueryFilter(nil),
			).Return([]*model.PurchasedItem{}, 0, nil)

			_, _, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, nil, nil)
			assert.NoError(t, err)
			m.AssertExpectations(t)
		})

		t.Run("PropagatesRepositoryError", func(t *testing.T) {
			purchasedItemListService, m := newService()

			repoErr := errors.New("boom")
			m.On(
				"PurchasedItemListLogs",
				1, 0, []int(nil), 10, 0,
				(*query.DateFilter)(nil),
				[]query.QueryFilter(nil),
			).Return(nil, 0, repoErr)

			rows, total, err := purchasedItemListService.PurchasedItemListLogs(1, 0, nil, 10, 1, nil, nil)
			assert.Nil(t, rows)
			assert.Zero(t, total)
			assert.Equal(t, repoErr, err)
			m.AssertExpectations(t)
		})
	})

}
