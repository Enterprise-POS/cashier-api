package service

import "cashier-api/model"

type UserService interface {
	/*
		Register
	*/
	SignUpWithEmailAndPassword(email string, password string, name string) (*model.User, error)

	/*
		Log in
	*/
	SignInWithEmailAndPassword(email string, password string) (*model.User, string, error)
}
