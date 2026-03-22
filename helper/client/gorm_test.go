package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenConnection(t *testing.T) {
	var db = CreateGormClient()
	assert.NotNil(t, db)
}
