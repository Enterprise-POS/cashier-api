package model

import "time"

type Store struct {
	Id        int        `json:"id,omitempty"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	IsActive  bool       `json:"is_active"`
	TenantId  int        `json:"tenant_id"`
}
