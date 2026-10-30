package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"testing"

	"gorm.io/gorm"
)

func TestVoucherRepositoryImpl(t *testing.T) {
	var gormClient *gorm.DB = client.CreateGormClient()

	voucherRepository := NewVoucherRepositoryImpl(gormClient)

	t.Run("CreateVoucher", func(t *testing.T) {
		t.Run("NormalCreateVoucher", func(t *testing.T) {
			// FROM_TOTAL_LIMITED_PERCENTAGE
			voucher := &model.Voucher{
				Code:             "TEST2026",
				Name:             "Test Voucher",
				TenantId:         333,
				VoucherType:      model.FROM_TOTAL_LIMITED_PERCENTAGE,
				UsageLimit:       0,
				Percentage:       10,
				MaximumDecrement: 10_000,
				RoundingPolicy:   "ROUND",
			}
			voucherRepository.CreateVoucher(voucher)
		})
	})
}
