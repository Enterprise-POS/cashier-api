package model

import (
	"time"
)

type Tenant struct {
	Id          int       `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	Name        string    `json:"name" gorm:"column:name"`
	OwnerUserId int       `json:"owner_user_id" gorm:"column:owner_user_id"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active"` // by default at database is TRUE
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`

	Users []User `json:"users,omitempty" gorm:"many2many:user_mtm_tenant;foreignKey:Id;joinForeignKey:TenantId;References:Id;joinReferences:UserId"`
}

func (tenant *Tenant) TableName() string {
	return "tenant"
}

type UserMtmTenant struct {
	Id        int        `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	UserId    int        `json:"user_id" gorm:"column:user_id"`
	TenantId  int        `json:"tenant_id" gorm:"column:tenant_id"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
}

func (userMtmTenant *UserMtmTenant) TableName() string {
	return "user_mtm_tenant"
}
