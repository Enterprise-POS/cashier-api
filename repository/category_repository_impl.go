package repository

import (
	"cashier-api/model"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CategoryRepositoryImpl struct {
	Client *gorm.DB
}

const CategoryTable string = "category"
const CategoryMtmWarehouseTable string = "category_mtm_warehouse"

func NewCategoryRepositoryImpl(client *gorm.DB) CategoryRepository {
	return &CategoryRepositoryImpl{
		Client: client,
	}
}

func (repository *CategoryRepositoryImpl) GetItemsByCategoryId(tenantId int, categoryId int, limit int, page int) ([]*model.CategoryWithItem, int, error) {
	start := page * limit

	var results []*model.CategoryWithItem
	err := repository.Client.
		Model(&model.Item{}).
		Select(`
			category.id AS category_id,
			category.category_name,
			warehouse.item_id,
			warehouse.item_name,
			warehouse.stocks,
			warehouse.base_price,
			COUNT(*) OVER() AS total_count
		`).
		Joins("INNER JOIN category_mtm_warehouse ON category_mtm_warehouse.item_id = warehouse.item_id").
		Joins("INNER JOIN category ON category.id = category_mtm_warehouse.category_id").
		Where("warehouse.tenant_id = ? AND category.id = ?", tenantId, categoryId).
		Limit(limit).
		Offset(start).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	countResult := 0
	if len(results) > 0 {
		countResult = results[0].TotalCount
	}

	return results, countResult, nil
}

func (repository *CategoryRepositoryImpl) GetCategoryWithItems(tenantId, page, limit int) ([]*model.CategoryWithItem, int, error) {
	start := page * limit

	var results = make([]*model.CategoryWithItem, 0)
	err := repository.Client.
		Model(&model.Item{}).
		Select(`
			category.id AS category_id,
			category.category_name,
			warehouse.item_id,
			warehouse.item_name,
			warehouse.stocks,
			warehouse.base_price,
			COUNT(*) OVER() AS total_count
		`).
		Joins("INNER JOIN category_mtm_warehouse ON category_mtm_warehouse.item_id = warehouse.item_id").
		Joins("INNER JOIN category ON category.id = category_mtm_warehouse.category_id").
		Where("warehouse.tenant_id = ?", tenantId).
		Limit(limit).
		Offset(start).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	countResult := 0
	if len(results) > 0 {
		countResult = results[0].TotalCount
	}

	return results, countResult, nil
}

func (repository *CategoryRepositoryImpl) Get(tenantId, page, limit int, nameQuery string) ([]*model.Category, int, error) {
	start := page * limit

	var results = make([]*model.Category, 0)
	var totalCount int64

	query := repository.Client.Model(&model.Category{}).Where("tenant_id = ?", tenantId)

	if nameQuery != "" {
		query = query.Where("category_name LIKE ?", nameQuery+"%")
	}

	// Get total count before applying pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(start).Limit(limit).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, int(totalCount), nil
}

func (repository *CategoryRepositoryImpl) Create(tenantId int, categories []*model.Category) ([]*model.Category, error) {
	if CategoryTable == "" {
		log.Errorf("Fatal Error ! CategoryRepositoryImpl.Create called with empty table. probably didn't use New Fn for create CategoryRepositoryImpl. TenantId: %d", tenantId)
		return nil, fmt.Errorf("CategoryRepositoryImpl.Create called with empty table. probably didn't use New Fn for create CategoryRepositoryImpl. TenantId: %d", tenantId)
	}

	// Ensure all categories belong to the given tenant
	for _, c := range categories {
		c.TenantId = tenantId
	}

	if err := repository.Client.Create(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (repository *CategoryRepositoryImpl) Register(tobeRegisters []*model.CategoryMtmWarehouse) error {
	if err := repository.Client.Create(&tobeRegisters).Error; err != nil {
		return err
	}
	return nil
}

func (repository *CategoryRepositoryImpl) Unregister(toUnregister *model.CategoryMtmWarehouse) error {
	result := repository.Client.
		Where("category_id = ? AND item_id = ?", toUnregister.CategoryId, toUnregister.ItemId).
		Delete(&model.CategoryMtmWarehouse{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 1 {
		log.Errorf("FATAL ERROR multiple categories deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId)
		return errors.New("FATAL ERROR multiple categories deleted")
	}
	if result.RowsAffected == 0 {
		log.Warnf("Warning ! Handled error, no data deleted from categoryId: %d, itemId: %d", toUnregister.CategoryId, toUnregister.ItemId)
		return errors.New("[WARN] No data deleted")
	}

	return nil
}

func (repository *CategoryRepositoryImpl) EditItemCategory(tenantId int, editedItemCategory *model.CategoryMtmWarehouse) error {
	if editedItemCategory.CategoryId <= 0 || editedItemCategory.ItemId <= 0 {
		return errors.New("[ERROR] Invalid request: invalid category_id or item_id")
	}

	return repository.Client.Transaction(func(tx *gorm.DB) error {
		// SELECT EXISTS (SELECT 1 FROM warehouse WHERE item_id = p_item_id AND tenant_id = p_tenant_id)
		err := tx.Model(&model.Item{}).
			Where("item_id = ? AND tenant_id = ?", editedItemCategory.ItemId, tenantId).
			First(&model.Item{}).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("[ERROR] Fatal error, current item from store never exist at warehouse")
		}
		if err != nil {
			return err
		}

		// SELECT COUNT(*) = 1 FROM category_mtm_warehouse WHERE item_id = p_item_id
		var linkCount int64
		if err := tx.Model(&model.CategoryMtmWarehouse{}).
			Where("item_id = ?", editedItemCategory.ItemId).
			Count(&linkCount).Error; err != nil {
			return err
		}
		if linkCount != 1 {
			return errors.New("[ERROR] Item has multiple categories either not registered by any category")
		}

		// UPDATE category_mtm_warehouse SET category_id = p_category_id WHERE item_id = p_item_id
		result := tx.Model(&model.CategoryMtmWarehouse{}).
			Where("item_id = ?", editedItemCategory.ItemId).
			Update("category_id", editedItemCategory.CategoryId)
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "foreign key") {
				return errors.New("[ERROR] Update failed: category (id) does not exist")
			}
			return errors.New("[ERROR] Unexpected database error")
		}

		return nil
	})
}

func (repository *CategoryRepositoryImpl) Update(tenantId int, categoryId int, tobeChangeCategoryName string) (*model.Category, error) {
	/*
		For now, only updating Category.CategoryName is allowed
		- category_name (ok)
		-	id (x)
		- created_at (x)
		- tenant_id (x)
	*/
	var updatedCategory model.Category

	err := repository.Client.Model(&updatedCategory).
		Where("tenant_id = ? AND id = ?", tenantId, categoryId).
		Update("category_name", tobeChangeCategoryName).Error
	if err != nil {
		return nil, err
	}

	// Fetch the updated record to return it
	if err := repository.Client.
		Where("tenant_id = ? AND id = ?", tenantId, categoryId).
		First(&updatedCategory).Error; err != nil {
		return nil, err
	}

	return &updatedCategory, nil
}

func (repository *CategoryRepositoryImpl) Delete(category *model.Category) error {
	/*
		NOTE
		When category deleted then category_mtm_warehouse that have the
		same deleted category id will be automatically deleted
	*/
	result := repository.Client.
		Where("tenant_id = ? AND id = ?", category.TenantId, category.Id).
		Delete(&model.Category{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 1 {
		log.Errorf("FATAL ERROR multiple categories deleted from categoryId: %d, tenantId: %d", category.Id, category.TenantId)
		return errors.New("FATAL ERROR multiple categories deleted")
	}
	if result.RowsAffected == 0 {
		log.Warnf("Warning ! Handled error, no data deleted from categoryId: %d, tenantId: %d", category.Id, category.TenantId)
		return errors.New("[WARN] No data deleted")
	}

	return nil
}
