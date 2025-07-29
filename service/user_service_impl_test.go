package service

import (
	"cashier-api/model"
	"cashier-api/repository"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserServiceImpl(t *testing.T) {
	if os.Getenv("JWT_S") == "" {
		t.Skip("Required ENV not available: JWT_S")
	}

	t.Run("SignUpWithEmailAndPassword", func(t *testing.T) {
		var userRepo = repository.NewUserRepositoryMock(&mock.Mock{}).(*repository.UserRepositoryMock)
		var userService = NewUserServiceImpl(userRepo)

		t.Run("NormalSignUp", func(t *testing.T) {
			now := time.Now()
			dummyIdentity := "NormalSignUp_" + uuid.NewString()
			dummyPassword := "12345678"
			expectedDummyUser := model.User{
				Name:      "TestUserServiceImpl " + dummyIdentity,
				Email:     dummyIdentity + "@gmail.com",
				Id:        1,
				UserUuid:  "",
				CreatedAt: &now,
			}

			// customMatcher say: "I don’t care about the full model.User, but if the .Email and .Name fields match expectedDummyUser, then treat it as a match."
			customMatcher := mock.MatchedBy(func(u model.User) bool {
				return u.Email == expectedDummyUser.Email && u.Name == expectedDummyUser.Name
			})

			userRepo.Mock.On("CreateWithEmailAndPassword", customMatcher, mock.AnythingOfType("string")).Return(&expectedDummyUser, nil)

			createdDummyUser, err := userService.SignUpWithEmailAndPassword(expectedDummyUser.Email, dummyPassword, expectedDummyUser.Name)
			assert.Nil(t, err)
			assert.NotNil(t, createdDummyUser)
			assert.Equal(t, expectedDummyUser.Id, createdDummyUser.Id)
			assert.Equal(t, expectedDummyUser.Name, createdDummyUser.Name)
			assert.Equal(t, expectedDummyUser.Email, createdDummyUser.Email)
			assert.Equal(t, expectedDummyUser.UserUuid, createdDummyUser.UserUuid)
			assert.NotNil(t, expectedDummyUser.CreatedAt, createdDummyUser.CreatedAt)
		})

		t.Run("InvalidEmail", func(t *testing.T) {
			dummyIdentity := "InvalidEmail" + uuid.NewString()
			dummyPassword := "12345678"
			expectedDummyUser := model.User{
				Name:  "TestUserServiceImpl " + dummyIdentity,
				Email: dummyIdentity + "gmail.com", // no @ while registering
			}
			errorMessage := "Could not create user account. check the input email"

			// We don't even need mock to test email and password validation
			createdDummyUser, err := userService.SignUpWithEmailAndPassword(expectedDummyUser.Email, dummyPassword, expectedDummyUser.Name)
			assert.Nil(t, createdDummyUser)
			assert.NotNil(t, err)
			assert.Equal(t, errorMessage, err.Error())
		})

		t.Run("InvalidPassword", func(t *testing.T) {
			dummyIdentity := "InvalidPassword" + uuid.NewString()
			dummyPassword := "123456" // less than 8 characters are not allowed
			expectedDummyUser := model.User{
				Name:  "TestUserServiceImpl " + dummyIdentity,
				Email: dummyIdentity + "@gmail.com",
			}
			errorMessage := "Could not create user account. check the input password"

			createdDummyUser, err := userService.SignUpWithEmailAndPassword(expectedDummyUser.Email, dummyPassword, expectedDummyUser.Name)
			assert.Nil(t, createdDummyUser)
			assert.NotNil(t, err)
			assert.Equal(t, errorMessage, err.Error())
		})
	})

	t.Run("SignInWithEmailAndPassword", func(t *testing.T) {
		var userRepo = repository.NewUserRepositoryMock(&mock.Mock{}).(*repository.UserRepositoryMock)
		var userService = NewUserServiceImpl(userRepo)

		t.Run("NormalSignIn", func(t *testing.T) {
			now := time.Now()
			dummyIdentity := "NormalSignUp_" + uuid.NewString()
			dummyPassword := "12345678"
			expectedDummyUser := model.User{
				Name:      "TestUserServiceImpl " + dummyIdentity,
				Email:     dummyIdentity + "@gmail.com",
				Id:        1,
				UserUuid:  "",
				CreatedAt: &now,
			}

			// customMatcher say: "I don’t care about the full model.User, but if the .Email and .Name fields match expectedDummyUser, then treat it as a match."
			customMatcher := mock.MatchedBy(func(u model.User) bool {
				return u.Email == expectedDummyUser.Email && u.Name == expectedDummyUser.Name
			})

			userRepo.Mock.On("CreateWithEmailAndPassword", customMatcher, mock.AnythingOfType("string")).Return(&expectedDummyUser, nil)
			createdDummyUser, err := userService.SignUpWithEmailAndPassword(expectedDummyUser.Email, dummyPassword, expectedDummyUser.Name)
			require.Nil(t, err)
			require.NotNil(t, createdDummyUser)

			hashedPassword := "$2a$10$BhhTb567SYl3CEZw.s9MlOsZCswa3/UdzTcGQcaU6zrbRMIbDiFiK"
			expectedDummyUserRegisterForm := &model.UserRegisterForm{
				Id:        createdDummyUser.Id,
				Name:      createdDummyUser.Name,
				Email:     createdDummyUser.Email,
				UserUuid:  createdDummyUser.UserUuid,
				CreatedAt: createdDummyUser.CreatedAt,
				Password:  hashedPassword,
			}

			userRepo.Mock.On("GetByEmail", createdDummyUser.Email).Return(expectedDummyUserRegisterForm, nil)
			user, tokenString, err := userService.SignInWithEmailAndPassword(createdDummyUser.Email, dummyPassword)
			assert.Nil(t, err)
			assert.NotNil(t, user)
			assert.NotEqual(t, "", tokenString)
			assert.Equal(t, createdDummyUser.Id, user.Id)
			assert.Equal(t, createdDummyUser.Name, user.Name)
			assert.Equal(t, createdDummyUser.Email, user.Email)

			// fmt.Println(user)
			// fmt.Println(token)

			// Test token payload
			claims := jwt.MapClaims{}

			// Parse and validate the token
			token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
				// Optional: verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("JWT_S")), nil
			})

			assert.Nil(t, err)
			assert.NotNil(t, token, "If this fail then failed to parse JWT")
			assert.True(t, token.Valid)

			for key, val := range claims {
				// fmt.Printf("%s: %v\n", key, val)
				switch key {
				case "sub":
					sub, ok := val.(float64)
					assert.True(t, ok)
					assert.Equal(t, user.Id, int(sub))
				case "uuid":
					assert.Equal(t, user.UserUuid, val)
				case "email":
					assert.Equal(t, user.Email, val)
				case "name":
					assert.Equal(t, user.Name, val)
				case "created_at":
					createdAtStr, ok := val.(string)
					assert.True(t, ok)

					createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
					assert.Nil(t, err)
					assert.NotNil(t, createdAt)
				case "exp":
					exp, ok := val.(float64)
					assert.True(t, ok)

					time := time.Unix(int64(exp), 0)
					assert.NotNil(t, time)
				default:
					fmt.Printf("Un tested jwt value key: %s", key)
				}
			}
		})
	})
}
