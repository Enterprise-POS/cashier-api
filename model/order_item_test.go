package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderItem(t *testing.T) {
	now := time.Now()
	orderItem := OrderItem{
		Id:             1,
		PurchasedPrice: 10000,
		CreatedAt:      &now,
		TotalQuantity:  1,
		TotalAmount:    10000,
		DiscountAmount: 0,
		Subtotal:       10000,
		TenantId:       1,
		StoreId:        1,
	}

	assert.Equal(t, 1, orderItem.Id)
	assert.Equal(t, 1, orderItem.TotalQuantity)
	assert.Equal(t, 10000, orderItem.PurchasedPrice)
	assert.Equal(t, 10000, orderItem.TotalAmount)
	assert.Equal(t, 0, orderItem.DiscountAmount)
	assert.Equal(t, 10000, orderItem.Subtotal)
	assert.Equal(t, 1, orderItem.TenantId)
	assert.Equal(t, 1, orderItem.StoreId)
	assert.NotNil(t, orderItem.CreatedAt)
}
