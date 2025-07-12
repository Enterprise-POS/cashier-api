package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryFilter(t *testing.T) {
	query1 := QueryFilter{
		Column:    CreatedAtColumn,
		Ascending: false,
	}
	assert.Equal(t, CreatedAtColumn, query1.Column)
	assert.Equal(t, false, query1.Ascending)

	query2 := QueryFilter{
		Column:    TotalAmountColumn,
		Ascending: false,
	}
	assert.Equal(t, TotalAmountColumn, query2.Column)
	assert.Equal(t, false, query2.Ascending)
}
