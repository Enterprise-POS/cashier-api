package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestVoucherUsageGORMRelationships(t *testing.T) {
	orderItemSchema, err := schema.Parse(&OrderItem{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	voucherUsages, ok := orderItemSchema.Relationships.Relations["VoucherUsages"]
	require.True(t, ok, "OrderItem.VoucherUsages relationship must be parsed")
	assert.Equal(t, schema.HasMany, voucherUsages.Type)
	assert.True(t, voucherUsages.Field.Readable)
	assert.False(t, voucherUsages.Field.Creatable)
	assert.False(t, voucherUsages.Field.Updatable)
	require.Len(t, voucherUsages.References, 1)
	assert.Equal(t, "OrderItemId", voucherUsages.References[0].ForeignKey.Name)
	assert.Equal(t, "Id", voucherUsages.References[0].PrimaryKey.Name)

	voucherUsageSchema, err := schema.Parse(&VoucherUsage{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	for name, foreignKey := range map[string]string{
		"Voucher":   "VoucherId",
		"OrderItem": "OrderItemId",
	} {
		t.Run(name, func(t *testing.T) {
			relationship, ok := voucherUsageSchema.Relationships.Relations[name]
			require.True(t, ok, "VoucherUsage.%s relationship must be parsed", name)
			require.Len(t, relationship.References, 1)
			assert.Equal(t, foreignKey, relationship.References[0].ForeignKey.Name)
			assert.Equal(t, "Id", relationship.References[0].PrimaryKey.Name)
		})
	}
}
