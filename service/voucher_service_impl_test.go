package service

import (
	"cashier-api/exception"
	"cashier-api/helper/query"
	"cashier-api/model"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type voucherRepositoryStub struct {
	voucher          *model.Voucher
	vouchers         []*model.Voucher
	voucherCount     int
	getVouchersErr   error
	getTenantId      int
	getLimit         int
	getPage          int
	getSearchQuery   string
	getSorts         []*query.QueryFilter
	findVoucherErr   error
	updatedVoucher   *model.Voucher
	voucherUpdates   *model.VoucherUpdate
	updateVoucherErr error
	createErr        error
	deleteErr        error
	findErr          error
	createdVoucher   *model.Voucher
	deletedVoucher   int
	deletedTenant    int
	lookupCode       string
	lookupTenant     int
	incrementCalls   int
	incrementErr     error
}

func (repository *voucherRepositoryStub) GetVouchers(tenantId, limit, page int, searchQuery string, queryFilters []*query.QueryFilter) ([]*model.Voucher, int, error) {
	repository.getTenantId = tenantId
	repository.getLimit = limit
	repository.getPage = page
	repository.getSearchQuery = searchQuery
	repository.getSorts = queryFilters
	return repository.vouchers, repository.voucherCount, repository.getVouchersErr
}

func (repository *voucherRepositoryStub) FindVoucherById(voucherId, tenantId int) (*model.Voucher, error) {
	return repository.voucher, repository.findVoucherErr
}

func (repository *voucherRepositoryStub) UpdateVoucher(voucherId, tenantId int, updates *model.VoucherUpdate) (*model.Voucher, error) {
	repository.voucherUpdates = updates
	return repository.updatedVoucher, repository.updateVoucherErr
}

func (repository *voucherRepositoryStub) CreateVoucher(voucher *model.Voucher) (*model.Voucher, error) {
	repository.createdVoucher = voucher
	if repository.createErr != nil {
		return nil, repository.createErr
	}
	voucher.Id = 123
	voucher.CreatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return voucher, nil
}

func (repository *voucherRepositoryStub) DeleteVoucher(voucherId, tenantId int) error {
	repository.deletedVoucher = voucherId
	repository.deletedTenant = tenantId
	return repository.deleteErr
}

func (repository *voucherRepositoryStub) FindVoucherByCode(code string, tenantId int) (*model.Voucher, error) {
	repository.lookupCode = code
	repository.lookupTenant = tenantId
	return repository.voucher, repository.findErr
}

func (repository *voucherRepositoryStub) IncrementVoucherUsedCount(voucherId, tenantId int) error {
	repository.incrementCalls++
	return repository.incrementErr
}

func TestVoucherServiceImpl(t *testing.T) {
	t.Run("EditVoucher", func(t *testing.T) {
		repositoryStub := &voucherRepositoryStub{
			voucher:        &model.Voucher{Id: 12, TenantId: 7, Code: "OLD", Name: "Old"},
			updatedVoucher: &model.Voucher{Id: 12, TenantId: 7, Code: "NEW", Name: "New"},
		}
		service := NewVoucherServiceImpl(repositoryStub)
		name := "New"
		code := " NEW "
		updates := &model.VoucherUpdate{Name: &name, Code: &code}

		updated, err := service.EditVoucher(12, 7, updates)

		require.NoError(t, err)
		assert.Same(t, repositoryStub.updatedVoucher, updated)
		assert.Equal(t, "NEW", *updates.Code)
		assert.Same(t, updates, repositoryStub.voucherUpdates)

		repositoryStub.updateVoucherErr = &pgconn.PgError{Code: "23505"}
		_, err = service.EditVoucher(12, 7, updates)
		var voucherError *exception.VoucherServiceError
		require.ErrorAs(t, err, &voucherError)
		assert.Equal(t, exception.VoucherServiceErrorConflict, voucherError.Kind)

		repositoryStub.updateVoucherErr = nil
		repositoryStub.voucher.StartAt = nil
		repositoryStub.voucher.EndAt = nil
		repositoryStub.voucherUpdates = nil
		startAt := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		endAt := startAt
		_, err = service.EditVoucher(12, 7, &model.VoucherUpdate{StartAt: &startAt, StartAtSet: true, EndAt: &endAt, EndAtSet: true})
		assert.EqualError(t, err, "Voucher end_at must be after start_at")
		assert.Nil(t, repositoryStub.voucherUpdates)

		_, err = service.EditVoucher(12, 7, &model.VoucherUpdate{})
		assert.EqualError(t, err, "At least one editable voucher field is required")
	})

	t.Run("GetVouchers", func(t *testing.T) {
		vouchers := []*model.Voucher{{Id: 10, Name: "Save", Code: "SAVE10"}}
		sort := []*query.QueryFilter{{Column: query.VoucherNameColumn, Ascending: true}}
		repository := &voucherRepositoryStub{vouchers: vouchers, voucherCount: 1}
		service := NewVoucherServiceImpl(repository)

		result, count, err := service.GetVouchers(7, 5, 2, " save ", sort)

		require.NoError(t, err)
		assert.Equal(t, vouchers, result)
		assert.Equal(t, 1, count)
		assert.Equal(t, 7, repository.getTenantId)
		assert.Equal(t, 5, repository.getLimit)
		assert.Equal(t, 1, repository.getPage)
		assert.Equal(t, "save", repository.getSearchQuery)
		assert.Same(t, sort[0], repository.getSorts[0])

		codeSort := []*query.QueryFilter{{Column: query.VoucherCodeColumn, Ascending: true}}
		_, _, err = service.GetVouchers(7, 5, 1, "", codeSort)
		require.NoError(t, err)
		assert.Same(t, codeSort[0], repository.getSorts[0])

		_, _, err = service.GetVouchers(7, 5, 1, "", []*query.QueryFilter{{Column: query.ItemName}})
		assert.EqualError(t, err, "Invalid sort column: item_name")
	})

	t.Run("CreateVoucher", func(t *testing.T) {
		t.Run("ReturnsCreatedVoucher", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := newPercentageVoucher("SAVE10")

			created, err := service.CreateVoucher(voucher)
			require.NoError(t, err)
			assert.Equal(t, "SAVE10", voucher.Code)
			assert.Same(t, voucher, repository.createdVoucher)
			require.NotNil(t, created)
			assert.Equal(t, 123, created.Id)
			assert.False(t, created.CreatedAt.IsZero())
		})

		t.Run("MapsPostgresUniqueViolationToSafeConflict", func(t *testing.T) {
			databaseError := &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint voucher_code_key",
			}
			repository := &voucherRepositoryStub{createErr: databaseError}
			service := NewVoucherServiceImpl(repository)

			created, err := service.CreateVoucher(newPercentageVoucher("SAVE10"))
			assert.Nil(t, created)
			var voucherError *exception.VoucherServiceError
			require.ErrorAs(t, err, &voucherError)
			assert.Equal(t, exception.VoucherServiceErrorConflict, voucherError.Kind)
			assert.Equal(t, "A voucher with conflicting information already exists.", voucherError.Error())
			assert.NotContains(t, voucherError.Error(), databaseError.Message)
		})

		t.Run("AllowsEmptyName", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := newPercentageVoucher("EMPTY")
			voucher.Name = ""

			created, err := service.CreateVoucher(voucher)
			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Empty(t, created.Name)
		})

		t.Run("RejectsNamesAndCodesWithUnsupportedCharacters", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := newPercentageVoucher("SAVE<10")
			voucher.Name = "Summer<Deal"

			created, err := service.CreateVoucher(voucher)
			assert.EqualError(t, err, "Invalid voucher name")
			assert.Nil(t, created)
			assert.Nil(t, repository.createdVoucher)

			voucher.Name = "Summer Deal"
			created, err = service.CreateVoucher(voucher)
			assert.EqualError(t, err, "Invalid voucher code")
			assert.Nil(t, created)
			assert.Nil(t, repository.createdVoucher)
		})

		t.Run("RejectsNameOver255Characters", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := newPercentageVoucher("LONG")
			voucher.Name = strings.Repeat("界", 256)

			created, err := service.CreateVoucher(voucher)
			assert.Error(t, err)
			assert.Nil(t, created)
			assert.Nil(t, repository.createdVoucher)
		})

		t.Run("RejectsEndAtNotAfterStartAt", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := newPercentageVoucher("DATES")
			startAt := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
			endAt := startAt
			voucher.StartAt = &startAt
			voucher.EndAt = &endAt

			created, err := service.CreateVoucher(voucher)
			assert.Error(t, err)
			assert.Nil(t, created)
			assert.Nil(t, repository.createdVoucher)
		})

		t.Run("RejectsITEMQUANTITIESWithMultipleItemIDs", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)
			voucher := &model.Voucher{
				TenantId:           7,
				Code:               "BAD",
				VoucherType:        model.ITEM_QUANTITIES,
				Decrement:          100,
				ItemQuantities:     2,
				ItemIdCombinations: []int{1, 2},
			}

			created, err := service.CreateVoucher(voucher)
			assert.Error(t, err)
			assert.Nil(t, created)
			assert.Nil(t, repository.createdVoucher)
		})
	})

	t.Run("DeleteVoucher", func(t *testing.T) {
		t.Run("ScopesDeleteToTenant", func(t *testing.T) {
			repository := &voucherRepositoryStub{}
			service := NewVoucherServiceImpl(repository)

			err := service.DeleteVoucher(3, 7)
			require.NoError(t, err)
			assert.Equal(t, 3, repository.deletedVoucher)
			assert.Equal(t, 7, repository.deletedTenant)
		})
	})

	t.Run("PreviewVoucher", func(t *testing.T) {
		repository := &voucherRepositoryStub{voucher: newPercentageVoucher("PREVIEW")}
		repository.voucher.Id = 12
		service := NewVoucherServiceImpl(repository)

		before, discount, after, voucherId, eligible, err := service.PreviewVoucher(7, "PREVIEW", []*model.Item{
			{ItemId: 1, TenantId: 7, BasePrice: 500},
		})

		require.NoError(t, err)
		assert.Equal(t, 500, before)
		assert.Equal(t, 50, discount)
		assert.Equal(t, 450, after)
		assert.Equal(t, 12, voucherId)
		assert.True(t, eligible)
		assert.Zero(t, repository.incrementCalls)
	})

	t.Run("CommitVoucherUsage", func(t *testing.T) {
		repository := &voucherRepositoryStub{}
		service := NewVoucherServiceImpl(repository)

		err := service.CommitVoucherUsage(12, 7)

		require.NoError(t, err)
		assert.Equal(t, 1, repository.incrementCalls)
	})

	t.Run("UseVoucher", func(t *testing.T) {
		t.Run("DiscountCalculation", func(t *testing.T) {
			tests := []struct {
				name         string
				voucher      *model.Voucher
				items        []*model.Item
				wantBefore   int
				wantDiscount int
				wantAfter    int
				wantErr      bool
				wantUsed     int
			}{
				{
					name: "PercentageCeilWithMaximumDecrement",
					voucher: &model.Voucher{
						VoucherType:      model.FROM_TOTAL_LIMITED_PERCENTAGE,
						Percentage:       10,
						RoundingPolicy:   "CEIL",
						MaximumDecrement: 50,
					},
					items:        []*model.Item{{ItemId: 1, TenantId: 7, BasePrice: 501}},
					wantBefore:   501,
					wantDiscount: 50,
					wantAfter:    451,
					wantUsed:     1,
				},
				{
					name: "StaticDecrement",
					voucher: &model.Voucher{
						VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement:   100,
					},
					items:        []*model.Item{{ItemId: 1, TenantId: 7, BasePrice: 500}},
					wantBefore:   500,
					wantDiscount: 100,
					wantAfter:    400,
					wantUsed:     1,
				},
				{
					name: "ExactItemCombination",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_COMBINATIONS,
						Decrement:          100,
						ItemIdCombinations: []int{1, 2},
					},
					items: []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 2, TenantId: 7, BasePrice: 300},
					},
					wantBefore:   600,
					wantDiscount: 100,
					wantAfter:    500,
					wantUsed:     1,
				},
				{
					name: "RepeatedItemIDsRequireMatchingQuantities",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_COMBINATIONS,
						Decrement:          100,
						ItemIdCombinations: []int{1, 1, 1},
					},
					items: []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 1, TenantId: 7, BasePrice: 300},
					},
					wantBefore:   900,
					wantDiscount: 100,
					wantAfter:    800,
					wantUsed:     1,
				},
				{
					name: "ExtraItemInvalidatesExactCombination",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_COMBINATIONS,
						Decrement:          100,
						ItemIdCombinations: []int{1, 2},
					},
					items: []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 2, TenantId: 7, BasePrice: 300},
						{ItemId: 3, TenantId: 7, BasePrice: 200},
					},
					wantBefore:   800,
					wantDiscount: 0,
					wantAfter:    800,
				},
				{
					name: "AdditionalUnitsInvalidateExactCombination",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_COMBINATIONS,
						Decrement:          100,
						ItemIdCombinations: []int{1, 1, 1},
					},
					items: []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 1, TenantId: 7, BasePrice: 300},
						{ItemId: 1, TenantId: 7, BasePrice: 300},
					},
					wantBefore:   1200,
					wantDiscount: 0,
					wantAfter:    1200,
				},
				{
					name: "ItemQuantityThresholdAppliesOnceToWholeCart",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_QUANTITIES,
						Decrement:          100,
						ItemIdCombinations: []int{1},
						ItemQuantities:     2,
					},
					items: []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 200},
						{ItemId: 1, TenantId: 7, BasePrice: 200},
						{ItemId: 2, TenantId: 7, BasePrice: 500},
					},
					wantBefore:   900,
					wantDiscount: 100,
					wantAfter:    800,
					wantUsed:     1,
				},
				{
					name: "ItemQuantityBelowThreshold",
					voucher: &model.Voucher{
						VoucherType:        model.ITEM_QUANTITIES,
						Decrement:          100,
						ItemIdCombinations: []int{1},
						ItemQuantities:     2,
					},
					items:        []*model.Item{{ItemId: 1, TenantId: 7, BasePrice: 500}},
					wantBefore:   500,
					wantDiscount: 0,
					wantAfter:    500,
				},
				{
					name: "DecrementExceedsSubtotal",
					voucher: &model.Voucher{
						VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement:   501,
					},
					items:   []*model.Item{{ItemId: 1, TenantId: 7, BasePrice: 500}},
					wantErr: true,
				},
			}

			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					test.voucher.Id = 1
					repository := &voucherRepositoryStub{voucher: test.voucher}
					service := NewVoucherServiceImpl(repository)

					before, discount, after, err := service.UseVoucher(7, " SAVE10 ", test.items)
					if test.wantErr {
						require.Error(t, err)
						assert.Zero(t, repository.incrementCalls)
						return
					}
					require.NoError(t, err)
					assert.Equal(t, test.wantBefore, before)
					assert.Equal(t, test.wantDiscount, discount)
					assert.Equal(t, test.wantAfter, after)
					assert.Equal(t, test.wantUsed, repository.incrementCalls)
					assert.Equal(t, "SAVE10", repository.lookupCode)
					assert.Equal(t, 7, repository.lookupTenant)
				})
			}
		})

		t.Run("PropagatesRepositoryError", func(t *testing.T) {
			wantErr := errors.New("database unavailable")
			repository := &voucherRepositoryStub{findErr: wantErr}
			service := NewVoucherServiceImpl(repository)

			_, _, _, err := service.UseVoucher(7, "SAVE10", []*model.Item{
				{ItemId: 1, TenantId: 7, BasePrice: 500},
			})
			var voucherError *exception.VoucherServiceError
			require.ErrorAs(t, err, &voucherError)
			assert.Equal(t, exception.VoucherServiceErrorInternal, voucherError.Kind)
			assert.NotContains(t, err.Error(), wantErr.Error())
			assert.Zero(t, repository.incrementCalls)
		})

		t.Run("UsageWindowAndLimit", func(t *testing.T) {
			now := time.Now()
			tests := []struct {
				name    string
				voucher *model.Voucher
				wantErr bool
			}{
				{
					name: "NotStarted",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10, StartAt: timePointer(now.Add(time.Hour)),
					},
					wantErr: true,
				},
				{
					name: "Expired",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10, EndAt: timePointer(now.Add(-time.Hour)),
					},
					wantErr: true,
				},
				{
					name: "UsageLimitReached",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10, UsageLimit: 1, UsedCount: 1,
					},
					wantErr: true,
				},
				{
					name: "UsageLimitZeroIsUnlimited",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10,
					},
				},
				{
					name: "OnlyStartBound",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10, StartAt: timePointer(now.Add(-time.Hour)),
					},
				},
				{
					name: "OnlyEndBound",
					voucher: &model.Voucher{
						Id: 1, VoucherType: model.FROM_TOTAL_LIMITED_STATIC_DECREMENT,
						Decrement: 10, EndAt: timePointer(now.Add(time.Hour)),
					},
				},
			}

			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					repository := &voucherRepositoryStub{voucher: test.voucher}
					service := NewVoucherServiceImpl(repository)

					_, _, _, err := service.UseVoucher(7, "SAVE10", []*model.Item{
						{ItemId: 1, TenantId: 7, BasePrice: 100},
					})
					wantIncrements := 1
					if test.wantErr {
						wantIncrements = 0
					}
					if test.wantErr {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
					}
					assert.Equal(t, wantIncrements, repository.incrementCalls)
				})
			}
		})
	})
}

func newPercentageVoucher(code string) *model.Voucher {
	return &model.Voucher{
		Code:           code,
		Name:           "Test Voucher",
		TenantId:       7,
		VoucherType:    model.FROM_TOTAL_LIMITED_PERCENTAGE,
		Percentage:     10,
		RoundingPolicy: "FLOOR",
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
