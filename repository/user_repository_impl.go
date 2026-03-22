package repository

import (
	"cashier-api/model"

	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	Client *gorm.DB
}

const UserTable string = "user"

func NewUserRepositoryImpl(client *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		Client: client,
	}
}

/*
UserRepositoryImpl

	Don't give pointer for 'user pointer'
	we want copy
*/
func (repository *UserRepositoryImpl) CreateWithEmailAndPassword(newUser model.User, password string) (*model.User, error) {
	newUser.Password = password // assign the (pre-hashed) password

	result := repository.Client.Create(&newUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}

func (repository *UserRepositoryImpl) GetByEmail(email string) (*model.User, error) {
	var user model.User

	result := repository.Client.Take(&user, "email = ?", email)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
