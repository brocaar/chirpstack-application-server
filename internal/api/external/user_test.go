package external

import (
	"testing"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestUser() {
	nsClient := mock.NewClient()
	nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
		DeviceProfile: &ns.DeviceProfile{},
	}
	networkserver.SetPool(mock.NewPool(nsClient))

	ctx := context.Background()
	validator := &TestValidator{}
	api := NewUserAPI(validator)
	apiInternal := NewInternalUserAPI(validator)

	ts.T().Run("Create user assigned to organization", func(t *testing.T) {
		assert := require.New(t)
		validator.returnIsAdmin = true

		org := storage.Organization{
			Name: "test-org",
		}
		assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

		createReq := pb.CreateUserRequest{
			User: &pb.User{
				Username: "testuser",
				Email:    "foo@bar.com",
			},
			Password: "testpasswd",
			Organizations: []*pb.UserOrganization{
				{OrganizationId: org.ID, IsAdmin: true},
			},
		}
		createResp, err := api.Create(ctx, &createReq)
		assert.NoError(err)
		assert.True(createResp.Id > 0)

		users, err := storage.GetOrganizationUsers(context.Background(), storage.DB(), org.ID, 10, 0)
		assert.NoError(err)
		assert.Len(users, 1)
		assert.Equal(createResp.Id, users[0].UserID)
		assert.True(users[0].IsAdmin)
	})

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)
		validator.returnIsAdmin = true

		createReq := &pb.CreateUserRequest{
			User: &pb.User{
				Username:   "username",
				IsAdmin:    true,
				SessionTtl: 180,
				Email:      "foo@bar.com",
			},
			Password: "pass^^ord",
		}
		createResp, err := api.Create(context.Background(), createReq)
		assert.NoError(err)
		assert.True(createResp.Id > 0)
		createReq.User.Id = createResp.Id

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			user, err := api.Get(ctx, &pb.GetUserRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(createReq.User, user.User)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			users, err := api.List(ctx, &pb.ListUserRequest{
				Limit:  10,
				Offset: 0,
			})
			assert.NoError(err)
			assert.Len(users.Result, 3) // 2 created users + admin user
			assert.EqualValues(3, users.TotalCount)
		})

		t.Run("Login", func(t *testing.T) {
			assert := require.New(t)

			jwt, err := apiInternal.Login(ctx, &pb.LoginRequest{
				Username: createReq.User.Username,
				Password: createReq.Password,
			})
			assert.NoError(err)
			assert.NotEqual("", jwt)
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			updateReq := pb.UpdateUserRequest{
				User: &pb.User{
					Id:         createResp.Id,
					Username:   "anotheruser",
					SessionTtl: 300,
					IsAdmin:    false,
					Email:      "bar@foo.com",
				},
			}
			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			userUpd, err := api.Get(ctx, &pb.GetUserRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(updateReq.User, userUpd.User)
		})

		t.Run("UpdatePassword", func(t *testing.T) {
			assert := require.New(t)

			updatePass := &pb.UpdateUserPasswordRequest{
				UserId:   createResp.Id,
				Password: "newpasstest",
			}
			_, err := api.UpdatePassword(ctx, updatePass)
			assert.NoError(err)

			jwt, err := apiInternal.Login(ctx, &pb.LoginRequest{
				Username: "anotheruser",
				Password: updatePass.Password,
			})
			assert.NoError(err)
			assert.NotEqual("", jwt)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(ctx, &pb.DeleteUserRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			users, err := api.List(ctx, &pb.ListUserRequest{
				Limit:  10,
				Offset: 0,
			})
			assert.NoError(err)
			assert.Len(users.Result, 2)
			assert.EqualValues(2, users.TotalCount)
		})
	})

}
