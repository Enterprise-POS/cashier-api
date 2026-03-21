package repository

import (
	"cashier-api/model"
	"errors"

	"gorm.io/gorm"
)

type PurchasedItemRepositoryImpl struct {
	Client *gorm.DB
}

func NewPurchasedItemRepositoryImpl(client *gorm.DB) PurchasedItemRepository {
	return &PurchasedItemRepositoryImpl{Client: client}
}

/*
CreateList:

	If we want insert 10 row and 1 row data violate / for example un-exist order_item_id
	then the supabase will fail all the 10 row,
	un exist`item_id will result error !
*/
func (repository *PurchasedItemRepositoryImpl) CreateList(data []*model.PurchasedItem, withReturn bool) ([]*model.PurchasedItem, error) {
	result := repository.Client.Create(&data)
	if result.Error != nil {
		return nil, result.Error
	}

	if withReturn {
		// GORM mutates data in-place, so it already contains id, created_at, etc.
		return data, nil
	}

	return nil, nil
}

func (repository *PurchasedItemRepositoryImpl) GetByOrderItemId(orderItemId int) ([]*model.PurchasedItem, error) {
	var result []*model.PurchasedItem
	if err := repository.Client.Where("order_item_id = ?", orderItemId).Find(&result).Error; err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("fatal error list of purchased item not available")
	}

	return result, nil
}
