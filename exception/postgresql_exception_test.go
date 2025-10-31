package exception

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLException(t *testing.T) {
	err := &PostgreSQLException{Code: "P001", Details: "", Hint: "", Message: "Error test message"}
	assert.Error(t, err)
	assert.Equal(t, "Error test message", err.Error())
}
