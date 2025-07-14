package repository

import (
	"cashier-api/helper/client"
	"cashier-api/model"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepositoryImpl(t *testing.T) {

	var supabaseClient = client.CreateSupabaseClient()
	t.Run("EmailAndPasswordRegister", func(t *testing.T) {
		userRepositoryImpl := NewUserRepositoryImpl(supabaseClient)

		t.Run("NormalRegister", func(t *testing.T) {
			testUser := model.User{
				Name:  "Test_EmailAndPasswordRegister_NormalRegister 1",
				Email: "TestEmailAndPasswordRegisterNormalRegister@gmail.com",
			}
			password := "password123"
			newCreatedTestUser, err := userRepositoryImpl.EmailAndPasswordRegister(testUser, password)

			require.Nil(t, err)
			require.NotNil(t, newCreatedTestUser)
			require.Equal(t, testUser.Name, newCreatedTestUser.Name)
			require.Equal(t, testUser.Email, newCreatedTestUser.Email)
			require.NotNil(t, newCreatedTestUser.CreatedAt)

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

		t.Run("DuplicateId", func(t *testing.T) {
			testUser := model.User{
				Name:  "Test_EmailAndPasswordRegister_NormalRegister 1",
				Email: "TestEmailAndPasswordRegisterNormalRegister@gmail.com",
			}
			password := "password123"
			newCreatedTestUser, err := userRepositoryImpl.EmailAndPasswordRegister(testUser, password)

			require.Nil(t, err)
			require.NotNil(t, newCreatedTestUser)
			require.Equal(t, testUser.Name, newCreatedTestUser.Name)
			require.Equal(t, testUser.Email, newCreatedTestUser.Email)
			require.NotNil(t, newCreatedTestUser.CreatedAt)

			// user create using the same Id
			testUser.Id = newCreatedTestUser.Id

			duplicateUser, err := userRepositoryImpl.EmailAndPasswordRegister(testUser, password)
			require.NotNil(t, err)
			require.Nil(t, duplicateUser)
			require.Equal(t, "(23505) duplicate key value violates unique constraint \"user_email_key\"", err.Error())

			// Only delete the first test user, because the second user should not be created
			_, _, err = supabaseClient.From(UserTable).Delete("", "").Eq("id", strconv.Itoa(newCreatedTestUser.Id)).Execute()
			require.Nil(t, err)
		})
	})
}
