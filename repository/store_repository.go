package repository

import "cashier-api/model"

/*
This is Store, different from Tenant.
While tenant a company, store is where
the user will sell their item
*/
type StoreRepository interface {
	/*
		Get All store, and filter available either active / non active only
	*/
	GetAll(tenantId, page, limit int, includeNonActive bool) ([]*model.Store, int, error)

	/*
		Edit store properties, example name
	*/
	Edit(tobeEditStore *model.Store) (*model.Store, error)

	/*
		Create new store
	*/
	Create(tenantId int, name string) (*model.Store, error)

	/*
		Either set to active / non-active
	*/
	SetActivate(tenantId, storeId int, setInto bool) error
}
