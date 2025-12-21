package query

/*
QueryFilter help we query the column for DESC / ASC value
*/
type ColumnName = string

type QueryFilter struct {
	Column    ColumnName `json:"column"`
	Ascending bool       `json:"ascending"`
}

// DateFilter represents a date range filter
type DateFilter struct {
	Column    string `json:"column"`     // e.g., "created_at", "updated_at", "order_date"
	StartDate *int64 `json:"start_date"` // nil means no start date filter
	EndDate   *int64 `json:"end_date"`   // nil means no end date filter
}

// Generic
const CreatedAtColumn ColumnName = "created_at"

// OrderItem, order_item
const TotalAmountColumn ColumnName = "total_amount"

// PurchasedItemList, purchased_item_list
const PurchasedItemListTable string = "purchased_item_list"
