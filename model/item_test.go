package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultProperties(t *testing.T) {
	var item = new(Item)
	assert.Equal(t, 0, item.ItemId)
	assert.Nil(t, item.CreatedAt)
}

func TestDefinedProperties(t *testing.T) {
	now := time.Now()
	item := &Item{
		ItemId:    1,
		ItemName:  "Hello World",
		Stocks:    999,
		TenantId:  1,
		CreatedAt: &now,
	}

	assert.Equal(t, 1, item.ItemId)
	assert.Equal(t, "Hello World", item.ItemName)
	assert.Equal(t, 999, item.Stocks)
	assert.Equal(t, 1, item.TenantId)
	assert.NotNil(t, item.CreatedAt)
}
