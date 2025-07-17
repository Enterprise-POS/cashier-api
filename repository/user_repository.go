package repository

import "cashier-api/model"

type UserRepository interface {
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
	EmailAndPasswordRegister(newUser model.User, password string) (*model.User, error)

	/*
		Deactivate user at DB. Not delete user from authentication
	*/
	// DeactivateUser(user *model.User) error
}
