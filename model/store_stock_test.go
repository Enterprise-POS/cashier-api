package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStoreStock(t *testing.T) {
	now := time.Now()
	storeStock := StoreStock{
		Id:        1,
		Stocks:    99,
		Price:     10000,
		ItemId:    1,
		CreatedAt: &now,
		TenantId:  1,
		StoreId:   1,
	}

	assert.Equal(t, 1, storeStock.Id)
	assert.Equal(t, 99, storeStock.Stocks)
	assert.Equal(t, 10000, storeStock.Price)
	assert.Equal(t, 1, storeStock.ItemId)
	assert.Equal(t, 1, storeStock.TenantId)
	assert.Equal(t, 1, storeStock.StoreId)
	assert.NotNil(t, storeStock.CreatedAt)
}
