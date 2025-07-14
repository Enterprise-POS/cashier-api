package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	now := time.Now()
	user := &User{
		Id:        1,
		UserUuid:  "00000000-0000-0000-0000-000000000000",
		Name:      "Test user",
		Email:     "Test@gmail.com",
		CreatedAt: &now,
	}

	assert.Equal(t, time.Now().UTC().Day(), user.CreatedAt.UTC().Day())
	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "Test user", user.Name)
	assert.Equal(t, "Test@gmail.com", user.Email)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", user.UserUuid)
}

func TestUserRegisterForm(t *testing.T) {
	now := time.Now()
	user := &UserRegisterForm{
		Id:        1,
		UserUuid:  "00000000-0000-0000-0000-000000000000",
		Name:      "Test user",
		Email:     "Test@gmail.com",
		CreatedAt: &now,
		Password:  "12345678",
	}

	assert.Equal(t, time.Now().UTC().Day(), user.CreatedAt.UTC().Day())
	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "Test user", user.Name)
	assert.Equal(t, "Test@gmail.com", user.Email)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", user.UserUuid)
	assert.Equal(t, "12345678", user.Password)
}
