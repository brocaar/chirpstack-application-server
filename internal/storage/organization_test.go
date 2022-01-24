package storage

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/test"
)

func (ts *StorageTestSuite) TestOrganization() {
	assert := require.New(ts.T())

	conf := test.GetConfig()
	assert.NoError(Setup(conf))

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	ts.T().Run("Create organization with invalid name", func(t *testing.T) {
		assert := require.New(t)
		org := Organization{
			Name:        "invalid name",
			DisplayName: "invalid organization",
		}
		err := CreateOrganization(context.Background(), DB(), &org)
		assert.Equal(ErrOrganizationInvalidName, errors.Cause(err))
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		org := Organization{
			Name:            "test-organization",
			DisplayName:     "test organization",
			MaxGatewayCount: 10,
			MaxDeviceCount:  20,
		}
		assert.NoError(CreateOrganization(context.Background(), DB(), &org))
		org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
		org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			o, err := GetOrganization(context.Background(), DB(), org.ID, false)
			assert.NoError(err)
			o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
			o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
			assert.Equal(org, o)
		})

		t.Run("GetCount", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetOrganizationCount(context.Background(), DB(), OrganizationFilters{})
			assert.NoError(err)
			assert.Equal(2, count) // first org is created by migration
		})

		t.Run("GetOrganizations", func(t *testing.T) {
			assert := require.New(t)

			items, err := GetOrganizations(context.Background(), DB(), OrganizationFilters{
				Limit: 10,
			})
			assert.NoError(err)

			assert.Len(items, 2)

			items[1].CreatedAt = items[1].CreatedAt.Truncate(time.Millisecond).UTC()
			items[1].UpdatedAt = items[1].UpdatedAt.Truncate(time.Millisecond).UTC()
			assert.Equal(org, items[1])
		})

		t.Run("GetUserCount", func(t *testing.T) {
			assert := require.New(t)

			count, err := GetOrganizationUserCount(context.Background(), DB(), org.ID)
			assert.NoError(err)
			assert.Equal(0, count)
		})

		t.Run("CreateUser - admin", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(CreateOrganizationUser(context.Background(), DB(), org.ID, 1, false, false, false)) // admin user

			t.Run("GetUser", func(t *testing.T) {
				assert := require.New(t)

				u, err := GetOrganizationUser(context.Background(), DB(), org.ID, 1)
				assert.NoError(err)
				assert.EqualValues(1, u.UserID)
				assert.Equal("admin", u.Email)
				assert.False(u.IsAdmin)
				assert.False(u.IsDeviceAdmin)
				assert.False(u.IsGatewayAdmin)
			})

			t.Run("GetUserCount", func(t *testing.T) {
				assert := require.New(t)

				c, err := GetOrganizationUserCount(context.Background(), DB(), org.ID)
				assert.NoError(err)
				assert.Equal(1, c)

				users, err := GetOrganizationUsers(context.Background(), DB(), org.ID, 10, 0)
				assert.NoError(err)
				assert.Len(users, 1)
			})

			t.Run("UpdateUser", func(t *testing.T) {
				assert := require.New(t)

				assert.NoError(UpdateOrganizationUser(context.Background(), DB(), org.ID, 1, true, true, true)) // admin user, dev and gateway user

				u, err := GetOrganizationUser(context.Background(), DB(), org.ID, 1)
				assert.NoError(err)

				assert.EqualValues(1, u.UserID)
				assert.Equal("admin", u.Email)
				assert.True(u.IsAdmin)
				assert.True(u.IsDeviceAdmin)
				assert.True(u.IsGatewayAdmin)
			})

			t.Run("DeleteUser", func(t *testing.T) {
				assert := require.New(t)

				assert.NoError(DeleteOrganizationUser(context.Background(), DB(), org.ID, 1)) // admin user
				c, err := GetOrganizationUserCount(context.Background(), DB(), org.ID)
				assert.NoError(err)
				assert.Equal(0, c)
			})
		})

		t.Run("CreateUser - new user", func(t *testing.T) {
			assert := require.New(t)

			user := User{
				IsActive: true,
				Email:    "foo@bar.com",
			}
			err := CreateUser(context.Background(), DB(), &user)
			assert.NoError(err)

			t.Run("GetCountForUser", func(t *testing.T) {
				assert := require.New(t)

				c, err := GetOrganizationCount(context.Background(), DB(), OrganizationFilters{
					UserID: user.ID,
				})
				assert.NoError(err)
				assert.Equal(0, c)

				orgs, err := GetOrganizations(context.Background(), DB(), OrganizationFilters{
					UserID: user.ID,
					Limit:  10,
				})
				assert.NoError(err)
				assert.Len(orgs, 0)
			})

			t.Run("CreateUser", func(t *testing.T) {
				assert := require.New(t)

				assert.NoError(CreateOrganizationUser(context.Background(), DB(), org.ID, user.ID, false, true, true))

				c, err := GetOrganizationCount(context.Background(), DB(), OrganizationFilters{
					UserID: user.ID,
				})
				assert.NoError(err)
				assert.Equal(1, c)

				orgs, err := GetOrganizations(context.Background(), DB(), OrganizationFilters{
					UserID: user.ID,
					Limit:  10,
				})
				assert.NoError(err)
				assert.Len(orgs, 1)
				assert.Equal(org.ID, orgs[0].ID)
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			org.Name = "test-organization-updated"
			org.DisplayName = "test organization updated"
			org.CanHaveGateways = true
			org.MaxGatewayCount = 30
			org.MaxDeviceCount = 40
			assert.NoError(UpdateOrganization(context.Background(), DB(), &org))

			org.CreatedAt = org.CreatedAt.Truncate(time.Millisecond).UTC()
			org.UpdatedAt = org.UpdatedAt.Truncate(time.Millisecond).UTC()

			o, err := GetOrganization(context.Background(), DB(), org.ID, false)
			assert.NoError(err)
			o.CreatedAt = o.CreatedAt.Truncate(time.Millisecond).UTC()
			o.UpdatedAt = o.UpdatedAt.Truncate(time.Millisecond).UTC()
			assert.Equal(org, o)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteOrganization(context.Background(), DB(), org.ID))

			_, err := GetOrganization(context.Background(), DB(), org.ID, false)
			assert.Equal(ErrDoesNotExist, errors.Cause(err))
		})
	})
}
