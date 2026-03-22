package model

import "time"

type User struct {
	Id        int       `json:"id,omitempty"         gorm:"primaryKey;autoIncrement;column:id"`
	UserUuid  string    `json:"user_uuid,omitempty"  gorm:"-"`
	Name      string    `json:"name"                 gorm:"column:name"`
	Email     string    `json:"email"                gorm:"column:email"`
	Password  string    `json:"-"                    gorm:"column:password"` // hide from JSON
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime;<-:create"`

	Tenants []Tenant `json:"tenants,omitempty" gorm:"many2many:user_mtm_tenant;foreignKey:Id;joinForeignKey:UserId;References:Id;joinReferences:TenantId"`
}

func (user *User) TableName() string {
	return "user"
}

type UserRegisterForm struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
