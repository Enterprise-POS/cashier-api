package service

import "cashier-api/model"

type StoreService interface {
	/*
		Get All store, and filter available either active / non active only
	*/
	GetAll(tenantId, page, limit int, includeNonActive bool) ([]*model.Store, int, error)

	/*
		Create new store
	*/
	Create(tenantId int, name string) (*model.Store, error)

	/*
		Either set to active / non-active
	*/
	SetActivate(tenantId, storeId int, setInto bool) error

	/*
		Edit store name
	*/
}
