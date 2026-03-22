package repository

import (
	"cashier-api/model"
	"fmt"

	"gorm.io/gorm"
)

type StoreRepositoryImpl struct {
	Client *gorm.DB
}

func NewStoreRepositoryImpl(client *gorm.DB) StoreRepository {
	return &StoreRepositoryImpl{Client: client}
}

const StoreTable = "store"

// GetAll implements StoreRepository.
func (repository *StoreRepositoryImpl) GetAll(tenantId, page, limit int, includeNonActive bool) ([]*model.Store, int, error) {
	start := page * limit

	var stores []*model.Store
	var totalCount int64

	query := repository.Client.Model(&model.Store{}).
		Where("tenant_id = ?", tenantId)

	if !includeNonActive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(start).Limit(limit).Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, int(totalCount), nil
}

// Create implements StoreRepository.
func (repository *StoreRepositoryImpl) Create(tenantId int, name string) (*model.Store, error) {
	store := &model.Store{
		TenantId: tenantId,
		Name:     name,
		IsActive: true,
	}

	if err := repository.Client.Create(store).Error; err != nil {
		return nil, err
	}

	return store, nil
}

// SetActivate implements StoreRepository.
func (repository *StoreRepositoryImpl) SetActivate(tenantId, storeId int, setInto bool) error {
	result := repository.Client.Model(&model.Store{}).
		Where("tenant_id", tenantId).Where("id", storeId).
		Update("is_active", setInto)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("[ERROR] No store found with tenant_id=%d and id=%d", tenantId, storeId)
	}

	return nil
}

// Edit implements StoreRepository.
func (repository *StoreRepositoryImpl) Edit(tobeEditStore *model.Store) (*model.Store, error) {
	result := repository.Client.Model(&model.Store{}).
		Where("tenant_id", tobeEditStore.TenantId).Where("id", tobeEditStore.Id).
		Update("name", tobeEditStore.Name)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		// Bug fix: original used tobeEditStore.Id twice, should be TenantId and Id
		return nil, fmt.Errorf("[ERROR] No store found with tenant_id=%d and id=%d", tobeEditStore.TenantId, tobeEditStore.Id)
	}

	// Fetch the updated record to return it
	var updatedStore model.Store
	if err := repository.Client.
		Where("tenant_id = ? AND id = ?", tobeEditStore.TenantId, tobeEditStore.Id).
		First(&updatedStore).Error; err != nil {
		return nil, err
	}

	return &updatedStore, nil
}
