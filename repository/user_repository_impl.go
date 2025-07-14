package repository

import (
	"cashier-api/model"

	"github.com/supabase-community/supabase-go"
)

type UserRepositoryImpl struct {
	Client *supabase.Client
}

const UserTable string = "user"

func NewUserRepositoryImpl(client *supabase.Client) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		Client: client,
	}
}

/*
UserRepositoryImpl

	Don't give pointer for 'user pointer'
	we want copy
*/
func (repository *UserRepositoryImpl) EmailAndPasswordRegister(newUser model.User, password string) (*model.User, error) {
	// // Create user via supabase authentication
	// signUpResponse, err := repository.Client.Auth.Signup(types.SignupRequest{
	// 	Email:    newUser.Email,
	// 	Password: password,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// Mutate copy of new user data
	// We still left id because id is SQL auto_increment
	// newUser.UserUuid = signUpResponse.ID.String()
	// newUser.CreatedAt = &signUpResponse.CreatedAt

	userForm := &model.UserRegisterForm{
		Name:     newUser.Name,
		Email:    newUser.Email,
		Password: password, // only insert encrypted password here
	}

	var newCreatedUser *model.User
	_, err := repository.Client.From(UserTable).
		Insert(userForm, false, "", "representation", "").
		Single().
		ExecuteTo(&newCreatedUser)
	if err != nil {
		// If while creating registering user into table then delete user from authentication
		// err = repository.Client.Auth.AdminDeleteUser(types.AdminDeleteUserRequest{
		// 	UserID: signUpResponse.User.ID})
		// if err != nil {
		// 	log.Error("Fatal Error ! Fatal error while registering new user and could not delete the user from authentication !")
		// 	log.Errorf("Admin should handle this error immediately. User should be registered recently name: %s, email: %s", newUser.Name, newUser.Email)
		// 	return nil, fmt.Errorf("call admin immediately for this error. User should be registered recently name: %s, email: %s", newUser.Name, newUser.Email)
		// }

		return nil, err
	}

	// success
	return newCreatedUser, nil
}
