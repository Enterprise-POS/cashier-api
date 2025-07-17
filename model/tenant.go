package model

import "time"

type Tenant struct {
	Id          int        `json:"id,omitempty"`
	Name        string     `json:"name"`
	OwnerUserId int        `json:"owner_user_id"`
	IsActive    bool       `json:"is_active"` // by default at database is TRUE
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type UserMtmTenant struct {
	Id        int        `json:"id,omitempty"`
	UserId    int        `json:"user_id"`
	TenantId  int        `json:"tenant_id"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
