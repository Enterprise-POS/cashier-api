package service

import "cashier-api/model"

type UserService interface {
	SignUpWithEmailAndPassword(email string, password string, name string) (*model.User, error)
	SignInWithEmailAndPassword(email string, password string) (*model.User, string, error)
}
