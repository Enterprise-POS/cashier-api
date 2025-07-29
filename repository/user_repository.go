package repository

import "cashier-api/model"

type UserRepository interface {
	/*
		Get all user
	*/
	GetByEmail(email string) (*model.UserRegisterForm, error)

	/*
		Will connect to tenant table
	*/
	CreateWithEmailAndPassword(newUser model.User, password string) (*model.User, error)

	/*
		Deactivate user at DB. Not delete user from authentication
	*/
	// DeactivateUser(user *model.User) error

	/*
		User will sign in using traditional email and password
	*/
}
