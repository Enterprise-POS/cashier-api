package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	Repository repository.UserRepository
}

func NewUserServiceImpl(repository repository.UserRepository) UserService {
	return &UserServiceImpl{
		Repository: repository,
	}
}

func (service *UserServiceImpl) SignUpWithEmailAndPassword(email string, password string, name string) (*model.User, error) {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return nil, errors.New("Could not create user account. check the input email")
	} else if len(password) < 8 {
		return nil, errors.New("Could not create user account. check the input password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Errorf("[SignUp:1] Error while create user account, from email %s, name %s", email, name)
		log.Errorf("[SignUp:1] Error info as %s", err.Error())
		return nil, errors.New("Something gone wrong here ! Could not create user account")
	}

	newUserCandidate := model.User{
		Email: email,
		Name:  name,
		// Id: , will be create automatic by supabase
		// UserUuid: , optional, but in this method no need
		// CreatedAt: , will be create automatic by supabase
	}

	createdNewUser, err := service.Repository.CreateWithEmailAndPassword(newUserCandidate, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return createdNewUser, nil
}

// SignInWithEmailAndPassword implements UserService.
func (service *UserServiceImpl) SignInWithEmailAndPassword(email string, password string) (*model.User, string, error) {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return nil, "", errors.New("Could not create user account. check the input email")
	} else if len(password) < 8 {
		return nil, "", errors.New("Could not create user account. check the input password")
	}

	candidateUser, err := service.Repository.GetByEmail(email)
	if err != nil {
		if err.Error() == "(PGRST116) JSON object requested, multiple (or no) rows returned" {
			return nil, "", errors.New("No user with this credentials")
		}
		return nil, "", err
	}

	// Compare sent in pass with saved user pass hash
	err = bcrypt.CompareHashAndPassword([]byte(candidateUser.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("No user with this credentials")
	}

	// Generate a jwt token
	tokenExpiredTime := time.Hour * 24 * 30
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        candidateUser.Id,
		"email":      candidateUser.Email,
		"name":       candidateUser.Name,
		"uuid":       candidateUser.UserUuid,
		"created_at": candidateUser.CreatedAt,
		"exp":        time.Now().UTC().Add(tokenExpiredTime).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_S")))
	if err != nil {
		return nil, "", errors.New("Failed to create token")
	}

	loggedInUser := &model.User{
		Id:        candidateUser.Id,
		Name:      candidateUser.Name,
		Email:     candidateUser.Email,
		UserUuid:  candidateUser.UserUuid,
		CreatedAt: candidateUser.CreatedAt,
	}

	return loggedInUser, tokenString, nil
}
