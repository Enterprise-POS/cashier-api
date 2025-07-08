package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategory(t *testing.T) {
	now := time.Now()
	category := &Category{
		Id:           1,
		CategoryName: "Test Category 1",
		CreatedAt:    &now,
		TenantId:     1,
	}

	assert.Equal(t, 1, category.Id)
	assert.Equal(t, "Test Category 1", category.CategoryName)
	assert.Equal(t, now.Day(), category.CreatedAt.Day())
	assert.Equal(t, 1, category.TenantId)
}

func TestCategoryWithItem(t *testing.T) {
	now := time.Now()
	categoryWithItem := &CategoryWithItem{
		Id:           1, // omitempty
		CategoryName: "Test Category With Item Name",
		CreatedAt:    &now, // omitempty
		TenantId:     1,
		ItemId:       1,
		ItemName:     "Apple",
		Stocks:       10,
		IsActive:     true,
	}

	assert.Equal(t, 1, categoryWithItem.Id)
	assert.Equal(t, "Test Category With Item Name", categoryWithItem.CategoryName)
	assert.Equal(t, &now, categoryWithItem.CreatedAt)
	assert.Equal(t, 1, categoryWithItem.TenantId)
	assert.Equal(t, 1, categoryWithItem.ItemId)
	assert.Equal(t, "Apple", categoryWithItem.ItemName)
	assert.Equal(t, 10, categoryWithItem.Stocks)
	assert.Equal(t, true, categoryWithItem.IsActive)
}
