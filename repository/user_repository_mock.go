package repository

import (
	"cashier-api/model"

	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	Mock *mock.Mock
}

func NewUserRepositoryMock(mock *mock.Mock) UserRepository {
	return &UserRepositoryMock{
		Mock: mock,
	}
}

// EmailAndPasswordLogin implements UserRepository.
func (repository *UserRepositoryMock) GetByEmail(email string) (*model.UserRegisterForm, error) {
	args := repository.Mock.Called(email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.UserRegisterForm), nil
}

// EmailAndPasswordRegister implements UserRepository.
func (repository *UserRepositoryMock) CreateWithEmailAndPassword(newUser model.User, password string) (*model.User, error) {
	args := repository.Mock.Called(newUser, password)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Normal condition
	return args.Get(0).(*model.User), nil
}
