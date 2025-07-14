package repository

import "cashier-api/model"

type User interface {
	/*
		Get all user
	*/
	// Get()

	/*
		Login user using email and password
	*/
	EmailAndPasswordLogin(email string, password string) (*model.User, error)

	/*
		Will connect to tenant table
	*/
	EmailAndPasswordRegister(user *model.User, password string) (*model.User, error)

	/*
		Deactivate user at DB. Not delete user from authentication
	*/
	// DeactivateUser(user *model.User) error
}
