package model

import (
	"time"
)

type Store struct {
	Id          int        `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	Name        string     `json:"name" gorm:"column:name"`
	IsActive    bool       `json:"is_active" gorm:"column:is_active"`
	TenantId    int        `json:"tenant_id" gorm:"column:tenant_id"`
	Address     string     `json:"address,omitempty" gorm:"column:address"`
	PhoneNumber string     `json:"phone_number,omitempty" gorm:"column:phone_number"`
	CreatedAt   *time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

func (store *Store) TableName() string {
	return "store"
}
