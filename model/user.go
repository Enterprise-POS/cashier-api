package model

import "time"

type User struct {
	Id        int        `json:"id,omitempty"`
	UserUuid  string     `json:"user_uuid,omitempty"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type UserRegisterForm struct {
	Id        int        `json:"id,omitempty"`
	UserUuid  string     `json:"user_uuid,omitempty"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Password  string     `json:"password"`
}
