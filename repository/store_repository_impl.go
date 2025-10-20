package repository

import (
	"cashier-api/model"
	"fmt"
	"strconv"

	"github.com/supabase-community/supabase-go"
)

type StoreRepositoryImpl struct {
	Client *supabase.Client
}

func NewStoreRepositoryImpl(client *supabase.Client) StoreRepository {
	return &StoreRepositoryImpl{Client: client}
}

const StoreTable = "store"

// GetAll implements StoreRepository.
func (repository *StoreRepositoryImpl) GetAll(tenantId, page, limit int, includeActiveStore bool) ([]*model.Store, int, error) {
	start := page * limit
	end := start + limit - 1
	query := repository.Client.From(StoreTable).
		Select("*", "exact", false).
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Range(start, end, "").
		Limit(limit, "")

	// Will not include non-active store / Only active store will be return
	if !includeActiveStore {
		query = query.Eq("is_active", "TRUE")
	}

	var stores []*model.Store
	count, err := query.ExecuteTo(&stores)
	if err != nil {
		return nil, 0, err
	}

	return stores, int(count), nil
}

// Create implements StoreRepository.
func (repository *StoreRepositoryImpl) Create(tenantId int, name string) (*model.Store, error) {
	var createdStore *model.Store
	_, err := repository.Client.From(StoreTable).
		Insert(&model.Store{TenantId: tenantId, Name: name}, false, "", "representation", "").
		Single().
		ExecuteTo(&createdStore)
	if err != nil {
		return nil, err
	}

	return createdStore, nil
}

// SetActivate implements StoreRepository.
func (repository *StoreRepositoryImpl) SetActivate(tenantId, storeId int, setInto bool) error {
	tobeUpdatedValue := map[string]interface{}{
		"is_active": setInto,
	}

	message, _, err := repository.Client.From(StoreTable).
		Update(tobeUpdatedValue, "", "").
		Eq("tenant_id", strconv.Itoa(tenantId)).
		Eq("id", strconv.Itoa(storeId)).
		Execute()
	if err != nil {
		return err
	}

	// If the request did not do anything then nothing happen, which mean invalid/error
	if len(message) == 0 || string(message) == "[]" {
		return fmt.Errorf("[ERROR] No store found with tenant_id=%d and id=%d", tenantId, storeId)
	}

	return nil
}
