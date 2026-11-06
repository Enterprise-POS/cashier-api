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
	StartDate *int64 `json:"start_date"` // nil means no start date filter
	EndDate   *int64 `json:"end_date"`   // nil means no end date filter

	// * DEPRECATED later delete for this property, use QueryFilter for filtering date
	Column ColumnName `json:"column"` // e.g., "created_at", "updated_at", "order_date"

}

// Generic sortable columns. Voucher-specific columns are used to sort voucher lists.
const (
	CreatedAtColumn   ColumnName = "created_at"
	TotalAmountColumn ColumnName = "total_amount" // OrderItem, order_item
	Quantity          ColumnName = "quantity"     // PurchasedItemList
	ItemName          ColumnName = "item_name"    // Warehouse, StoreStock
	VoucherNameColumn ColumnName = "name"         // Voucher
	VoucherCodeColumn ColumnName = "code"         // Voucher
)

// PurchasedItem, purchased_item_list
const PurchasedItemTable string = "purchased_item_list"

func IsValidColumn(column ColumnName) bool {
	switch column {
	case CreatedAtColumn, TotalAmountColumn, Quantity, ItemName, VoucherNameColumn, VoucherCodeColumn:
		return true
	default:
		return false
	}
}
