package repository

import (
	"cashier-api/helper/query"
	"cashier-api/model"
	"errors"
	"strconv"

	"github.com/supabase-community/supabase-go"
)

type PurchasedItemRepositoryImpl struct {
	Client *supabase.Client
}

func NewPurchasedItemRepositoryImpl(client *supabase.Client) PurchasedItemRepository {
	return &PurchasedItemRepositoryImpl{Client: client}
}

/*
CreateList:

	If we want insert 10 row and 1 row data violate / for example un-exist order_item_id
	then the supabase will fail all the 10 row,
	un exist`item_id will result error !
*/
func (repository *PurchasedItemRepositoryImpl) CreateList(data []*model.PurchasedItem, withReturn bool) ([]*model.PurchasedItem, error) {
	if withReturn {
		var purchasedItemList []*model.PurchasedItem
		_, err := repository.Client.From(query.PurchasedItemTable).
			Insert(data, false, "", "representation", "").
			ExecuteTo(&purchasedItemList)
		if err != nil {
			return nil, err
		}

		return purchasedItemList, nil
	} else {
		_, _, err := repository.Client.From(query.PurchasedItemTable).
			Insert(data, false, "", "representation", "").
			Execute() // Use .Execute() because we don't want the result
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func (repository *PurchasedItemRepositoryImpl) GetByOrderItemId(orderItemId int) ([]*model.PurchasedItem, error) {
	// order_item_id guarantee return unique list only for that order_item
	// Do not limit the query, we want all the list.
	var result []*model.PurchasedItem
	_, err := repository.Client.From(query.PurchasedItemTable).
		Select("*", "", false).
		Eq("order_item_id", strconv.Itoa(orderItemId)).
		ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("fatal error list of purchased item not available")
	}

	return result, nil
}
