package storage

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		assert := require.New(t)

		u := User{Email: "admin"}
		assert.NoError(u.Validate())
	})

	t.Run("non admin", func(t *testing.T) {

		t.Run("invalid", func(t *testing.T) {
			assert := require.New(t)

			tests := []string{
				"nonadmin",
				"foo-bar.baz@example.com ",
			}

			for _, tst := range tests {
				u := User{Email: tst}
				assert.Equal(ErrInvalidEmail, u.Validate())
			}

		})

		t.Run("valid", func(t *testing.T) {
			assert := require.New(t)

			tests := []string{
				"foo-bar.baz@example.com",
			}

			for _, tst := range tests {
				u := User{Email: tst}
				assert.NoError(u.Validate())
			}
		})
	})
}

func (ts *StorageTestSuite) TestUser() {
	// Set a user secret so JWTs can be assigned
	jwtsecret = []byte("DoWahDiddy")

	ts.T().Run("Create with invalid password", func(t *testing.T) {
		assert := require.New(t)

		user := User{
			IsAdmin:    false,
			SessionTTL: 40,
			Email:      "foo@bar.com",
		}
		err := user.SetPasswordHash("bad")
		assert.Equal(ErrUserPasswordLength, errors.Cause(err))
	})

	ts.T().Run("Create with invalid email", func(t *testing.T) {
		assert := require.New(t)

		user := User{
			IsAdmin:    false,
			SessionTTL: 40,
			Email:      "foobar.com",
		}
		err := CreateUser(context.Background(), DB(), &user)

		assert.Equal(ErrInvalidEmail, errors.Cause(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)
		externalID := "ext-123"

		user := User{
			IsAdmin:       false,
			SessionTTL:    20,
			Email:         "foo@bar.com",
			EmailVerified: true,
			ExternalID:    &externalID,
		}
		password := "somepassword"
		assert.NoError(user.SetPasswordHash(password))

		err := CreateUser(context.Background(), DB(), &user)
		assert.NoError(err)

		t.Run("GetUser", func(t *testing.T) {
			assert := require.New(t)

			user2, err := GetUser(context.Background(), DB(), user.ID)
			assert.NoError(err)
			assert.Equal(user.Email, user2.Email)
			assert.Equal(user.IsAdmin, user2.IsAdmin)
			assert.Equal(user.SessionTTL, user2.SessionTTL)
		})

		t.Run("GetUserByEmail", func(t *testing.T) {
			assert := require.New(t)

			user2, err := GetUserByEmail(context.Background(), DB(), user.Email)
			assert.NoError(err)
			assert.Equal(user.Email, user2.Email)
			assert.Equal(user.IsAdmin, user2.IsAdmin)
			assert.Equal(user.SessionTTL, user2.SessionTTL)
		})

		t.Run("GetUserByExternalID", func(t *testing.T) {
			assert := require.New(t)

			user2, err := GetUserByExternalID(context.Background(), DB(), externalID)
			assert.NoError(err)
			assert.Equal(user.Email, user2.Email)
			assert.Equal(user.IsAdmin, user2.IsAdmin)
			assert.Equal(user.SessionTTL, user2.SessionTTL)
		})

		t.Run("GetUsers", func(t *testing.T) {
			assert := require.New(t)

			users, err := GetUsers(context.Background(), DB(), 10, 0)
			assert.NoError(err)
			assert.Len(users, 2)

			assert.Equal("admin", users[0].Email)
			assert.Equal("foo@bar.com", users[1].Email)
		})

		t.Run("GetUserCount", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetUserCount(context.Background(), DB())
			assert.NoError(err)
			assert.Equal(2, count)
		})

		t.Run("LoginUserByPassword", func(t *testing.T) {
			assert := require.New(t)

			jwt, err := LoginUserByPassword(context.Background(), DB(), user.Email, password)
			assert.NoError(err)
			assert.NotEqual("", jwt)
		})

		t.Run("GetUserToken", func(t *testing.T) {
			assert := require.New(t)

			token, err := GetUserToken(user)
			assert.NoError(err)
			assert.NotEqual("", token)
		})

		t.Run("Update password", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(user.SetPasswordHash("newrandompassword"))
			assert.NoError(UpdateUser(context.Background(), DB(), &user))

			_, err := LoginUserByPassword(context.Background(), DB(), user.Email, password)
			assert.Error(err)

			jwt, err := LoginUserByPassword(context.Background(), DB(), user.Email, "newrandompassword")
			assert.NoError(err)
			assert.NotEqual("", jwt)
		})

		t.Run("UpdateUser", func(t *testing.T) {
			assert := require.New(t)
			externalID := "test-123"

			user.Email = "updated@user.com"
			user.EmailVerified = false
			user.ExternalID = &externalID

			assert.NoError(UpdateUser(context.Background(), DB(), &user))

			user2, err := GetUser(context.Background(), DB(), user.ID)
			assert.NoError(err)
			assert.Equal(user.Email, user2.Email)
			assert.Equal(externalID, *user2.ExternalID)
		})

		t.Run("DeleteUser", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteUser(context.Background(), DB(), user.ID))

			_, err := GetUser(context.Background(), DB(), user.ID)
			assert.Equal(ErrDoesNotExist, errors.Cause(err))
		})
	})
}
