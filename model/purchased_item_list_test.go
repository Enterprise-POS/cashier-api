package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurchasedItem(t *testing.T) {
	now := time.Now()
	purchasedItemList := &PurchasedItem{
		Id:             1,
		CreatedAt:      &now,
		ItemId:         1,
		OrderItemId:    1,
		Quantity:       99,
		PurchasedPrice: 99 * 10000,
		DiscountAmount: 0,
		TotalAmount:    99 * 10000,
	}

	assert.Equal(t, 1, purchasedItemList.Id)
	assert.Equal(t, now.Day(), purchasedItemList.CreatedAt.Day())
	assert.Equal(t, 1, purchasedItemList.ItemId)
	assert.Equal(t, 1, purchasedItemList.OrderItemId)
	assert.Equal(t, 99, purchasedItemList.Quantity)
	assert.Equal(t, 99*10000, purchasedItemList.PurchasedPrice)
	assert.Equal(t, 0, purchasedItemList.DiscountAmount)
	assert.Equal(t, 99*10000, purchasedItemList.TotalAmount)
}
