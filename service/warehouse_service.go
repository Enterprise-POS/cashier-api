package service

import "cashier-api/model"

/*
package service

	typically handle the logic by user inputs
*/
type WarehouseService interface {
	GetWarehouseItems(tenantId, limit, page int) ([]*model.Item, int, error)
}
