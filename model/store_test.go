package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStoreTest(t *testing.T) {
	now := time.Now()
	storeTest := Store{
		Id:        1,
		Name:      "Store Model Test",
		IsActive:  true,
		TenantId:  1,
		CreatedAt: &now,
	}

	assert.IsType(t, int(0), storeTest.Id)
	assert.IsType(t, "", storeTest.Name)
	assert.IsType(t, true, storeTest.IsActive)
	assert.IsType(t, int(0), storeTest.TenantId)
	assert.IsType(t, (*time.Time)(nil), storeTest.CreatedAt)

	assert.Equal(t, 1, storeTest.Id)
	assert.Equal(t, "Store Model Test", storeTest.Name)
	assert.Equal(t, true, storeTest.IsActive)
	assert.Equal(t, 1, storeTest.TenantId)
	assert.NotNil(t, storeTest.CreatedAt)
	assert.WithinDuration(t, now, *storeTest.CreatedAt, time.Second, "CreatedAt should be around now")
}
