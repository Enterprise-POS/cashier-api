package repository

import (
	"cashier-api/model"
	"errors"
	"strings"

	"gorm.io/gorm"
)

const WarehouseTable string = "warehouse"

type WarehouseRepositoryImpl struct {
	Client *gorm.DB
}

func NewWarehouseRepositoryImpl(client *gorm.DB) WarehouseRepository {
	return &WarehouseRepositoryImpl{
		Client: client,
	}
}

// GetActiveItem implements WarehouseRepository.
func (warehouse *WarehouseRepositoryImpl) GetActiveItem(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {

	var items []*model.Item
	var total int64

	offset := page * limit

	query := warehouse.Client.Model(&model.Item{}).
		Where("tenant_id = ?", tenantId).
		Where("is_active = ?", true)

	if nameQuery != "" {
		query = query.Where("item_name LIKE ?", nameQuery+"%")
	}

	// Get total count first
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated result
	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

func (warehouse *WarehouseRepositoryImpl) Get(tenantId int, limit int, page int, nameQuery string) ([]*model.Item, int, error) {
	var items = make([]*model.Item, 0)
	var total int64

	offset := page * limit

	query := warehouse.Client.Model(&model.Item{}).
		Where("tenant_id = ?", tenantId)

	if nameQuery != "" {
		query = query.Where("item_name LIKE ?", nameQuery+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

func (warehouse *WarehouseRepositoryImpl) FindById(itemId int, tenantId int) (*model.Item, error) {
	var item model.Item

	err := warehouse.Client.
		Where("item_id = ?", itemId).
		Where("tenant_id = ?", tenantId).
		First(&item).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (warehouse *WarehouseRepositoryImpl) CreateItem(items []*model.Item) ([]*model.Item, error) {
	if err := warehouse.Client.Create(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (warehouse *WarehouseRepositoryImpl) Edit(quantity int, item *model.Item) error {
	var result string
	err := warehouse.Client.Raw("SELECT edit_warehouse_item(?, ?, ?, ?, ?, ?)",
		quantity,
		item.ItemName,
		item.StockType,
		item.BasePrice,
		item.ItemId,
		item.TenantId,
	).Scan(&result).Error

	if err != nil {
		return err
	}

	if strings.Contains(result, "[ERROR]") {
		return errors.New(result)
	}

	return nil
}

func (warehouse *WarehouseRepositoryImpl) SetActivate(tenantId, itemId int, setInto bool) error {
	result := warehouse.Client.Model(&model.Item{}).
		Where("tenant_id = ?", tenantId).
		Where("item_id = ?", itemId).
		Update("is_active", setInto)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("ITEM_NOT_FOUND")
	}

	return nil
}

// FindCompleteById implements WarehouseRepository.
func (warehouse *WarehouseRepositoryImpl) FindCompleteById(itemId int, tenantId int) (*model.CategoryWithItem, error) {
	var result model.CategoryWithItem

	err := warehouse.Client.Raw(`
		SELECT * FROM find_complete_by_id(?, ?)
	`, tenantId, itemId).
		Scan(&result).Error

	if err != nil {
		// Preserve original PostgreSQL error codes
		if strings.Contains(err.Error(), "NO_DATA_FOUND") {
			return nil, errors.New("NO_DATA_FOUND")
		}

		if strings.Contains(err.Error(), "CARDINALITY_VIOLATION") {
			return nil, errors.New("CARDINALITY_VIOLATION")
		}

		return nil, err
	}

	return &result, nil
}
