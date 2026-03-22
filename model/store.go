package model

import "time"

type Store struct {
	Id        int       `json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `json:"name" gorm:"column:name"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;<-:create"`
	IsActive  bool      `json:"is_active" gorm:"column:is_active"`
	TenantId  int       `json:"tenant_id" gorm:"column:tenant_id"`
}

func (store *Store) TableName() string {
	return "store"
}
