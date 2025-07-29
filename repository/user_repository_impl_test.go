package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryImpl(t *testing.T) {
	var supabaseClient = client.CreateSupabaseClient()

	t.Run("CreateWithEmailAndPassword", func(t *testing.T) {
		userRepositoryImpl := NewUserRepositoryImpl(supabaseClient)

		t.Run("NormalRegister", func(t *testing.T) {
			testUser := model.User{
				Name:  "Test_EmailAndPasswordRegister_NormalRegister 1",
				Email: "TestEmailAndPasswordRegisterNormalRegister@gmail.com",
			}
			password := "password123"
			newCreatedTestUser, err := userRepositoryImpl.CreateWithEmailAndPassword(testUser, password)

			assert.Nil(t, err)
			assert.NotNil(t, newCreatedTestUser)
			assert.Equal(t, testUser.Name, newCreatedTestUser.Name)
			assert.Equal(t, testUser.Email, newCreatedTestUser.Email)
			assert.NotNil(t, newCreatedTestUser.CreatedAt)

			// Clean up
			// DB
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedTestUser.Id)).Execute()
			require.Nil(t, err)

			// Supabase Auth
			// WARNING ! Deleting user from authentication is prohibited if in production.
			// the code below will not work if you don't set the MODE = 'dev'
			// uuid, err := uuid.Parse(newCreatedTestUser.UserUuid)
			// require.Nil(t, err)
			// err = supabaseClient.Auth.AdminDeleteUser(types.AdminDeleteUserRequest{UserID: uuid})
			// require.Nil(t, err)
		})

		t.Run("DuplicateEmail", func(t *testing.T) {
			testUser := model.User{
				Name:  "Test_EmailAndPasswordRegister_DuplicateId 1",
				Email: "TestEmailAndPasswordRegisterDuplicateId1@gmail.com",
			}
			password := "password123"
			newCreatedTestUser, err := userRepositoryImpl.CreateWithEmailAndPassword(testUser, password)

			require.Nil(t, err)
			require.NotNil(t, newCreatedTestUser)
			require.Equal(t, testUser.Name, newCreatedTestUser.Name)
			require.Equal(t, testUser.Email, newCreatedTestUser.Email)
			require.NotNil(t, newCreatedTestUser.CreatedAt)

			// user create using the same email
			duplicateUser, err := userRepositoryImpl.CreateWithEmailAndPassword(*newCreatedTestUser, password)
			assert.NotNil(t, err)
			assert.Nil(t, duplicateUser)
			assert.Equal(t, "(23505) duplicate key value violates unique constraint \"user_email_key\"", err.Error())

			// Only delete the first test user, because the second user should not be created
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedTestUser.Id)).Execute()
			require.Nil(t, err)
		})

		t.Run("DuplicateUUID", func(t *testing.T) {
			testUser := model.User{
				Name:     "Test_EmailAndPasswordRegister_DuplicateUUID 1",
				Email:    "TestEmailAndPasswordRegisterDuplicateUUID1@gmail.com",
				UserUuid: "00000000-0000-0000-0000-000000000000",
			}
			password := "password123"
			newCreatedTestUser, err := userRepositoryImpl.CreateWithEmailAndPassword(testUser, password)

			require.Nil(t, err)
			require.NotNil(t, newCreatedTestUser)
			require.Equal(t, testUser.Name, newCreatedTestUser.Name)
			require.Equal(t, testUser.Email, newCreatedTestUser.Email)
			require.NotNil(t, newCreatedTestUser.CreatedAt)

			// Using the same testUser data but because the Id is empty it should be ok, but the user_uuid field is
			// UNIQUE so should be return error
			// assign different email, so duplicate email could be by passed
			testUser.Email = "TestEmailAndPasswordRegisterDuplicateUUID2@gmail.com"
			duplicateUser, err := userRepositoryImpl.CreateWithEmailAndPassword(testUser, password)
			assert.NotNil(t, err)
			assert.Nil(t, duplicateUser)
			assert.Equal(t, "(23505) duplicate key value violates unique constraint \"user_user_uuid_key\"", err.Error())

			// Only delete the first test user, because the second user should not be created
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedTestUser.Id)).Execute()
			require.Nil(t, err)
		})
	})

	t.Run("GetByEmail", func(t *testing.T) {
		userRepo := NewUserRepositoryImpl(supabaseClient)

		t.Run("NormalGet", func(t *testing.T) {
			userIdentity := "GetByEmail " + uuid.NewString()
			userPassword := "12345678"
			testUser := model.User{
				Name:  "TestUserRepositoryImpl_" + userIdentity,
				Email: "TestUserRepositoryImpl_" + userIdentity + "@gmail.com",
			}

			dummyUser, err := userRepo.CreateWithEmailAndPassword(testUser, userPassword)
			require.Nil(t, err)
			require.NotNil(t, dummyUser)
			require.Equal(t, testUser.Email, dummyUser.Email)

			testDummyUser, err := userRepo.GetByEmail(dummyUser.Email)
			assert.Nil(t, err)
			assert.NotNil(t, testDummyUser)
			assert.Equal(t, dummyUser.Id, testDummyUser.Id)
			assert.Equal(t, dummyUser.Email, testDummyUser.Email)
			assert.Equal(t, dummyUser.Name, testDummyUser.Name)
			assert.Equal(t, dummyUser.UserUuid, testDummyUser.UserUuid)
			assert.NotNil(t, dummyUser.CreatedAt, testDummyUser.CreatedAt)

			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(testDummyUser.Id)).Execute()
			require.Nil(t, err)
		})
	})
}
