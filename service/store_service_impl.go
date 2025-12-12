package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type StoreServiceImpl struct {
	Repository        repository.StoreRepository
	CategoryNameRegex *regexp.Regexp
}

func NewStoreServiceImpl(repository repository.StoreRepository) StoreService {
	return &StoreServiceImpl{
		Repository:        repository,
		CategoryNameRegex: regexp.MustCompile(`^[a-zA-Z0-9_ ]{1,50}$`), // We wil limit the name for store
	}
}

// Create implements StoreService.
func (service *StoreServiceImpl) Create(tenantId int, name string) (*model.Store, error) {
	if tenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	if !service.CategoryNameRegex.MatchString(name) {
		return nil, fmt.Errorf("Current store name is not allowed: %s", name)
	}

	createdStore, err := service.Repository.Create(tenantId, name)
	if err != nil {
		if strings.Contains(err.Error(), "(23505)") {
			return nil, errors.New("Current store name already used / duplicate name")
		}

		return nil, err
	}

	return createdStore, nil
}

// GetAll implements StoreService.
func (service *StoreServiceImpl) GetAll(tenantId int, page int, limit int, includeNonActive bool) ([]*model.Store, int, error) {
	// 0 means, usually null but GO does not allow null so instead null will get 0
	if tenantId < 1 {
		return nil, 0, errors.New("Invalid tenant id")
	}

	if limit < 1 {
		return nil, 0, fmt.Errorf("limit could not less than 1 (limit >= 1). Given limit %d", limit)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("page could not less than 1 (page >= 1). Given page %d", page)
	}

	// By default SQL will be start from 0 index, if page 1 then page have to subtracted by 1 (page = 0)
	stores, count, err := service.Repository.GetAll(tenantId, page-1, limit, includeNonActive)
	if err != nil {
		if strings.Contains(err.Error(), "(PGRST103)") {
			return nil, 0, errors.New("Requested range not satisfiable")
		}

		return nil, 0, err
	}

	return stores, count, nil
}

// SetActivate implements StoreService.
func (service *StoreServiceImpl) SetActivate(tenantId int, storeId int, setInto bool) error {
	if storeId < 1 {
		return errors.New("Invalid store id")
	}

	if tenantId < 1 {
		return errors.New("Invalid tenant id")
	}

	err := service.Repository.SetActivate(tenantId, storeId, setInto)
	if err != nil {
		return err
	}

	return nil
}

// Edit implements StoreService.
func (service *StoreServiceImpl) Edit(tobeEditStore *model.Store) (*model.Store, error) {
	if tobeEditStore.Id < 1 {
		return nil, errors.New("Invalid store id")
	}

	if tobeEditStore.TenantId < 1 {
		return nil, errors.New("Invalid tenant id")
	}

	if !service.CategoryNameRegex.MatchString(tobeEditStore.Name) {
		return nil, fmt.Errorf("Current store name is not allowed: %s", tobeEditStore.Name)
	}

	editedStore, err := service.Repository.Edit(tobeEditStore)
	if err != nil {
		return nil, err
	}

	return editedStore, nil
}
