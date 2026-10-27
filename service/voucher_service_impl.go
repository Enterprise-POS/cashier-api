package service

import (
	"cashier-api/exception"
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const maxVoucherNameLength = 255

type VoucherServiceImpl struct {
	Repository repository.VoucherRepository
}

func NewVoucherServiceImpl(repository repository.VoucherRepository) VoucherService {
	return &VoucherServiceImpl{
		Repository: repository,
	}
}

func (service *VoucherServiceImpl) CreateVoucher(voucher *model.Voucher) (*model.Voucher, error) {
	if voucher == nil {
		return nil, errors.New("Voucher is required")
	}
	if voucher.Id != 0 {
		return nil, errors.New("Voucher ID must not be provided")
	}
	if voucher.TenantId < 1 {
		return nil, errors.New("Tenant ID is required")
	}
	if utf8.RuneCountInString(voucher.Name) > maxVoucherNameLength {
		return nil, fmt.Errorf("Voucher name cannot exceed %d characters", maxVoucherNameLength)
	}
	voucher.Code = strings.TrimSpace(voucher.Code)
	if voucher.Code == "" {
		return nil, errors.New("Voucher code is required")
	}
	if !model.IsValidVoucherType(voucher.VoucherType) {
		return nil, fmt.Errorf("Invalid voucher type: %s", voucher.VoucherType)
	}
	if !voucher.CreatedAt.IsZero() {
		return nil, errors.New("Voucher created_at must not be provided")
	}
	if voucher.DeletedAt.Valid {
		return nil, errors.New("Voucher deleted_at must not be provided")
	}
	if voucher.UsedCount != 0 {
		return nil, errors.New("Voucher used_count must not be provided")
	}
	if voucher.UsageLimit < 0 {
		return nil, errors.New("Voucher usage_limit cannot be negative")
	}
	if voucher.StartAt != nil && voucher.EndAt != nil && !voucher.EndAt.After(*voucher.StartAt) {
		return nil, errors.New("Voucher end_at must be after start_at")
	}

	// This is default value even the user not specified anything
	if voucher.RoundingPolicy == "" {
		voucher.RoundingPolicy = "ROUND"
	}

	switch voucher.VoucherType {
	case model.FROM_TOTAL_LIMITED_PERCENTAGE:
		if voucher.Percentage < 1 || voucher.Percentage > 100 {
			return nil, errors.New("Voucher percentage must be between 1 and 100")
		}
		if voucher.MaximumDecrement < 0 {
			return nil, errors.New("Voucher maximum decrement cannot be negative")
		}
		if !isValidRoundingPolicy(voucher.RoundingPolicy) {
			return nil, fmt.Errorf("Invalid voucher rounding policy: %s", voucher.RoundingPolicy)
		}
	case model.FROM_TOTAL_LIMITED_STATIC_DECREMENT, model.ITEM_COMBINATIONS, model.ITEM_QUANTITIES:
		if voucher.Decrement < 1 {
			return nil, errors.New("Voucher decrement must be greater than zero")
		}
	}

	switch voucher.VoucherType {
	case model.ITEM_COMBINATIONS:
		if err := validateItemIds(voucher.ItemIdCombinations); err != nil {
			return nil, err
		}
	case model.ITEM_QUANTITIES:
		if len(voucher.ItemIdCombinations) != 1 || voucher.ItemIdCombinations[0] < 1 {
			return nil, errors.New("ITEM_QUANTITIES requires exactly one valid item ID")
		}
		if voucher.ItemQuantities < 1 {
			return nil, errors.New("Voucher item quantities must be greater than zero")
		}
	}

	createdVoucher, err := service.Repository.CreateVoucher(voucher)
	if err != nil {
		return nil, service.handleVoucherRepositoryError("create", voucher.TenantId, err)
	}

	return createdVoucher, nil
}

func (service *VoucherServiceImpl) DeleteVoucher(voucherId, tenantId int) error {
	if voucherId < 1 {
		return errors.New("Voucher ID is required")
	}
	if tenantId < 1 {
		return errors.New("Tenant ID is required")
	}

	err := service.Repository.DeleteVoucher(voucherId, tenantId)
	if errors.Is(err, repository.ErrVoucherNotFound) {
		return &exception.VoucherServiceError{
			Kind:    exception.VoucherServiceErrorNotFound,
			Message: "Voucher not found.",
		}
	}
	if err != nil {
		return service.handleVoucherRepositoryError("delete", tenantId, err)
	}

	return nil
}

func (service *VoucherServiceImpl) UseVoucher(tenantId int, code string, items []*model.Item) (int, int, int, error) {
	if tenantId < 1 {
		return 0, 0, 0, errors.New("Tenant ID is required")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, 0, 0, errors.New("Voucher code is required")
	}
	if len(items) == 0 {
		return 0, 0, 0, errors.New("At least one purchased item is required")
	}

	totalPrice := 0
	maxInt := int(^uint(0) >> 1)
	for _, item := range items {
		if item == nil {
			return 0, 0, 0, errors.New("Purchased item cannot be nil")
		}
		if item.ItemId < 1 {
			return 0, 0, 0, errors.New("Purchased item ID is required")
		}
		if item.TenantId != tenantId {
			return 0, 0, 0, errors.New("Purchased item does not belong to the requested tenant")
		}
		if item.BasePrice < 0 {
			return 0, 0, 0, fmt.Errorf("Purchased item price cannot be negative: %d", item.BasePrice)
		}
		if totalPrice > maxInt-item.BasePrice {
			return 0, 0, 0, errors.New("Purchased item total exceeds the maximum supported amount")
		}
		totalPrice += item.BasePrice
	}

	voucher, err := service.Repository.FindVoucherByCode(code, tenantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, 0, &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorNotFound,
				Message: "Voucher not found.",
			}
		}
		return 0, 0, 0, service.handleVoucherRepositoryError("find", tenantId, err)
	}
	if voucher == nil {
		return 0, 0, 0, errors.New("Voucher not found")
	}
	now := time.Now()
	if voucher.StartAt != nil && now.Before(*voucher.StartAt) {
		return 0, 0, 0, errors.New("Voucher is not active yet")
	}
	if voucher.EndAt != nil && now.After(*voucher.EndAt) {
		return 0, 0, 0, errors.New("Voucher has expired")
	}
	if voucher.UsageLimit < 0 || voucher.UsedCount < 0 {
		return 0, 0, 0, errors.New("Invalid voucher usage configuration")
	}
	if voucher.UsageLimit > 0 && voucher.UsedCount >= voucher.UsageLimit {
		return 0, 0, 0, errors.New("Voucher usage limit reached")
	}

	discountPrice, eligible, err := calculateVoucherDiscount(voucher, items, totalPrice)
	if err != nil {
		return 0, 0, 0, err
	}
	if !eligible {
		return totalPrice, 0, totalPrice, nil
	}
	if err := service.Repository.IncrementVoucherUsedCount(voucher.Id, tenantId); err != nil {
		if errors.Is(err, repository.ErrVoucherUsageUnavailable) {
			return 0, 0, 0, &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorConflict,
				Message: "Voucher usage limit has been reached or the voucher is no longer available.",
			}
		}
		return 0, 0, 0, service.handleVoucherRepositoryError("increment usage", tenantId, err)
	}

	return totalPrice, discountPrice, totalPrice - discountPrice, nil
}

func (service *VoucherServiceImpl) handleVoucherRepositoryError(operation string, tenantId int, err error) error {
	fields := log.Fields{
		"operation": operation,
		"component": "voucher_service",
		"tenant_id": tenantId,
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		fields["postgres_code"] = postgresError.Code
		fields["constraint"] = postgresError.ConstraintName
		log.WithFields(fields).WithError(err).Error("Voucher database operation failed")

		switch postgresError.Code {
		case "23505":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorConflict,
				Message: "A voucher with conflicting information already exists.",
			}
		case "23503":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorValidation,
				Message: "One or more referenced records do not exist.",
			}
		case "23P01":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorConflict,
				Message: "Voucher data conflicts with an existing record.",
			}
		case "23502", "23514", "22001", "22P02":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorValidation,
				Message: "Voucher data is invalid. Please check the submitted fields.",
			}
		case "40001", "40P01":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorConflict,
				Message: "The voucher could not be updated due to a concurrent request. Please try again.",
			}
		case "08000", "08003", "08006", "08001", "08004", "08007", "08P01",
			"53300", "57P01", "57P02", "57P03", "57014":
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorUnavailable,
				Message: "Voucher service is temporarily unavailable. Please try again later.",
			}
		default:
			if strings.HasPrefix(postgresError.Code, "08") ||
				strings.HasPrefix(postgresError.Code, "53") ||
				strings.HasPrefix(postgresError.Code, "57") {
				return &exception.VoucherServiceError{
					Kind:    exception.VoucherServiceErrorUnavailable,
					Message: "Voucher service is temporarily unavailable. Please try again later.",
				}
			}
			return &exception.VoucherServiceError{
				Kind:    exception.VoucherServiceErrorInternal,
				Message: "Unable to complete the voucher request. Please try again later.",
			}
		}
	}

	log.WithFields(fields).WithError(err).Error("Voucher database operation failed")
	return &exception.VoucherServiceError{
		Kind:    exception.VoucherServiceErrorInternal,
		Message: "Unable to complete the voucher request. Please try again later.",
	}
}

func calculateVoucherDiscount(voucher *model.Voucher, items []*model.Item, totalPrice int) (int, bool, error) {
	if voucher == nil {
		return 0, false, errors.New("Voucher not found")
	}

	switch voucher.VoucherType {
	case model.FROM_TOTAL_LIMITED_PERCENTAGE:
		if voucher.MaximumDecrement < 0 {
			return 0, false, errors.New("Voucher maximum decrement cannot be negative")
		}
		discountPrice, err := percentageDiscount(totalPrice, voucher.Percentage, voucher.RoundingPolicy)
		if err != nil {
			return 0, false, err
		}
		if voucher.MaximumDecrement > 0 && discountPrice > voucher.MaximumDecrement {
			discountPrice = voucher.MaximumDecrement
		}
		return discountPrice, true, nil
	case model.FROM_TOTAL_LIMITED_STATIC_DECREMENT:
		if voucher.Decrement < 1 {
			return 0, false, errors.New("Voucher decrement must be greater than zero")
		}
		discountPrice, err := validateDecrement(voucher.Decrement, totalPrice)
		return discountPrice, true, err
	case model.ITEM_COMBINATIONS:
		if voucher.Decrement < 1 {
			return 0, false, errors.New("Voucher decrement must be greater than zero")
		}
		if err := validateItemIds(voucher.ItemIdCombinations); err != nil {
			return 0, false, err
		}
		if !matchesExactItemCombination(voucher.ItemIdCombinations, items) {
			return 0, false, nil
		}
		discountPrice, err := validateDecrement(voucher.Decrement, totalPrice)
		return discountPrice, true, err
	case model.ITEM_QUANTITIES:
		if voucher.Decrement < 1 {
			return 0, false, errors.New("Voucher decrement must be greater than zero")
		}
		if len(voucher.ItemIdCombinations) != 1 || voucher.ItemIdCombinations[0] < 1 || voucher.ItemQuantities < 1 {
			return 0, false, errors.New("Invalid ITEM_QUANTITIES voucher configuration")
		}
		quantity := 0
		for _, item := range items {
			if item.ItemId == voucher.ItemIdCombinations[0] {
				quantity++
			}
		}
		if quantity < voucher.ItemQuantities {
			return 0, false, nil
		}
		discountPrice, err := validateDecrement(voucher.Decrement, totalPrice)
		return discountPrice, true, err
	default:
		return 0, false, fmt.Errorf("Invalid voucher type: %s", voucher.VoucherType)
	}
}

func percentageDiscount(totalPrice, percentage int, roundingPolicy string) (int, error) {
	if percentage < 1 || percentage > 100 {
		return 0, errors.New("Voucher percentage must be between 1 and 100")
	}

	whole := totalPrice / 100 * percentage
	remainder := totalPrice % 100 * percentage
	var fractional int
	switch roundingPolicy {
	case "CEIL":
		fractional = (remainder + 99) / 100
	case "ROUND":
		fractional = (remainder + 50) / 100
	case "FLOOR":
		fractional = remainder / 100
	default:
		return 0, fmt.Errorf("Invalid voucher rounding policy: %s", roundingPolicy)
	}

	return whole + fractional, nil
}

func validateDecrement(decrement, totalPrice int) (int, error) {
	if decrement < 0 {
		return 0, errors.New("Voucher decrement cannot be negative")
	}
	if decrement > totalPrice {
		return 0, errors.New("Voucher decrement exceeds the total price")
	}
	return decrement, nil
}

func matchesExactItemCombination(voucherItemIds []int, items []*model.Item) bool {
	if len(voucherItemIds) == 0 {
		return false
	}

	expected := make(map[int]int, len(voucherItemIds))
	for _, itemId := range voucherItemIds {
		if itemId < 1 {
			return false
		}
		expected[itemId]++
	}

	purchased := make(map[int]int, len(items))
	for _, item := range items {
		purchased[item.ItemId]++
	}
	if len(expected) != len(purchased) {
		return false
	}
	for itemId, quantity := range expected {
		if purchased[itemId] != quantity {
			return false
		}
	}

	return true
}

func validateItemIds(itemIds []int) error {
	if len(itemIds) == 0 {
		return errors.New("At least one voucher item ID is required")
	}
	for _, itemId := range itemIds {
		if itemId < 1 {
			return errors.New("Voucher item IDs must be greater than zero")
		}
	}
	return nil
}

func isValidRoundingPolicy(roundingPolicy string) bool {
	switch roundingPolicy {
	case "CEIL", "ROUND", "FLOOR":
		return true
	default:
		return false
	}
}
