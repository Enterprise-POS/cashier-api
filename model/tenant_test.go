package model

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTenant(t *testing.T) {
	now := time.Now()
	tenant := &Tenant{
		Id:          1,
		Name:        "Test tenant",
		OwnerUserId: 1,
		CreatedAt:   &now,
		// IsActive:    false,
	}

	assert.Equal(t, 1, tenant.Id)
	assert.Equal(t, "Test tenant", tenant.Name)
	assert.Equal(t, 1, tenant.OwnerUserId)
	assert.Equal(t, time.Now().UTC().Day(), tenant.CreatedAt.UTC().Day())
	assert.Equal(t, false, tenant.IsActive)

	fmt.Println(os.Getenv("MODE"))
}

func TestUserMtmTenant(t *testing.T) {
	now := time.Now()
	tenant := &UserMtmTenant{
		Id:        1,
		UserId:    1,
		TenantId:  1,
		CreatedAt: &now,
	}

	assert.Equal(t, 1, tenant.Id)
	assert.Equal(t, 1, tenant.UserId)
	assert.Equal(t, 1, tenant.TenantId)
	assert.Equal(t, time.Now().UTC().Day(), tenant.CreatedAt.UTC().Day())
}
