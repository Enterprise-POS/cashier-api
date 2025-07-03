package repository

import (
	"cashier-api/model"
	"encoding/json"

	"github.com/supabase-community/supabase-go"
)

const PURCHASED_ITEM_LIST_TABLE string = "purchased_item_list"

type PurchasedItemListRepositoryImpl struct {
	Client *supabase.Client
}

/*
CreateList:

	If we want insert 10 row and 1 row data violate / for example un-exist order_item_id
	then the supabase will fail all the 10 row
*/
func (repository *PurchasedItemListRepositoryImpl) CreateList(data []*model.PurchasedItemList, withReturn bool) ([]*model.PurchasedItemList, error) {
	if withReturn {
		result, _, err := repository.Client.From(PURCHASED_ITEM_LIST_TABLE).Insert(data, false, "", "representation", "").Execute()
		if err != nil {
			return nil, err
		}

		var purchasedItemList []*model.PurchasedItemList
		err = json.Unmarshal(result, &purchasedItemList)
		if err != nil {
			return nil, err
		}

		return purchasedItemList, nil
	} else {
		_, _, err := repository.Client.From(PURCHASED_ITEM_LIST_TABLE).Insert(data, false, "", "representation", "").Execute()
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}
