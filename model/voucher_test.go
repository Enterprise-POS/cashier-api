package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoucherItemCombinationsUsePostgresArrayValues(t *testing.T) {
	voucher := &Voucher{
		ItemIdCombinations: VoucherItemIDs{9299, 9296},
	}
	array, err := voucher.ItemIdCombinations.Value()
	require.NoError(t, err)
	assert.Equal(t, "{9299,9296}", array)

	var loaded VoucherItemIDs
	err = loaded.Scan("{9299,9296}")
	require.NoError(t, err)
	assert.Equal(t, VoucherItemIDs{9299, 9296}, loaded)
}

func TestVoucherDecrementUsesLowercaseJSONKey(t *testing.T) {
	data, err := json.Marshal(Voucher{Decrement: 100})
	require.NoError(t, err)

	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &fields))
	assert.Equal(t, float64(100), fields["decrement"])
	assert.NotContains(t, fields, "Decrement")
}
