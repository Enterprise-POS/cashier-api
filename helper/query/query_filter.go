package query

/*
QueryFilter help we query the column for DESC / ASC value
*/
type ColumnName = string

type QueryFilter struct {
	Column    ColumnName
	Ascending bool
}

// Generic
const CreatedAtColumn ColumnName = "created_at"

// OrderItem, order_item
const TotalAmountColumn ColumnName = "total_amount"

// PurchasedItemList, purchased_item_list
const PurchasedItemListTable string = "purchased_item_list"
