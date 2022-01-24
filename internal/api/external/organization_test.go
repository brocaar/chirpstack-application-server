package external

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestOrganization() {
	assert := require.New(ts.T())

	validator := &TestValidator{}
	api := NewOrganizationAPI(validator)

	validator.returnSubject = "user"

	adminUser := storage.User{
		Email:    "admin@user.com",
		IsActive: true,
		IsAdmin:  true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &adminUser))

	user := storage.User{
		Email:    "some@user.com",
		IsActive: true,
		IsAdmin:  false,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

	ts.T().Run("Create with invalid name", func(t *testing.T) {
		assert := require.New(t)

		validator.returnUser = adminUser
		createReq := &pb.CreateOrganizationRequest{
			Organization: &pb.Organization{
				Name:            "organization name",
				DisplayName:     "Display Name",
				CanHaveGateways: true,
			},
		}
		_, err := api.Create(context.Background(), createReq)
		assert.NotNil(err)
	})

	ts.T().Run("Create as global admin", func(t *testing.T) {
		assert := require.New(t)

		validator.returnUser = adminUser
		createReq := pb.CreateOrganizationRequest{
			Organization: &pb.Organization{
				Name:            "orgName",
				DisplayName:     "Display Name",
				CanHaveGateways: true,
				MaxGatewayCount: 2,
				MaxDeviceCount:  10,
			},
		}
		createResp, err := api.Create(context.Background(), &createReq)
		assert.Nil(err)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			org, err := api.Get(context.Background(), &pb.GetOrganizationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			createReq.Organization.Id = createResp.Id
			assert.Equal(createReq.Organization, org.Organization)
		})

		t.Run("List", func(t *testing.T) {
			assert := require.New(t)

			orgs, err := api.List(context.Background(), &pb.ListOrganizationRequest{
				Limit:  10,
				Offset: 0,
			})
			assert.NoError(err)

			// Default org is already in the database.
			assert.Len(orgs.Result, 2)

			assert.Equal(createReq.Organization.Name, orgs.Result[1].Name)
			assert.Equal(createReq.Organization.DisplayName, orgs.Result[1].DisplayName)
			assert.Equal(createReq.Organization.CanHaveGateways, orgs.Result[1].CanHaveGateways)
		})

		t.Run("As user", func(t *testing.T) {
			assert := require.New(t)

			user := storage.User{
				Email:    "foo@bar.com",
				IsActive: true,
			}
			assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

			validator.returnUser = user

			t.Run("User can not list organizations", func(t *testing.T) {
				assert := require.New(t)

				orgs, err := api.List(context.Background(), &pb.ListOrganizationRequest{
					Limit:  10,
					Offset: 0,
				})
				assert.NoError(err)

				assert.EqualValues(0, orgs.TotalCount)
				assert.Len(orgs.Result, 0)
			})

			t.Run("Add user to organization", func(t *testing.T) {
				addOrgUser := &pb.AddOrganizationUserRequest{
					OrganizationUser: &pb.OrganizationUser{
						OrganizationId: createResp.Id,
						Email:          user.Email,
						IsAdmin:        false,
						IsDeviceAdmin:  false,
						IsGatewayAdmin: false,
					},
				}
				_, err := api.AddUser(context.Background(), addOrgUser)
				assert.NoError(err)

				t.Run("List organizations for user", func(t *testing.T) {
					assert := require.New(t)

					validator.returnUser = user

					orgs, err := api.List(context.Background(), &pb.ListOrganizationRequest{
						Limit:  10,
						Offset: 0,
					})
					assert.NoError(err)

					assert.EqualValues(1, orgs.TotalCount)
					assert.Len(orgs.Result, 1)
				})

				t.Run("User is part of organization", func(t *testing.T) {
					assert := require.New(t)

					orgUsers, err := api.ListUsers(context.Background(), &pb.ListOrganizationUsersRequest{
						OrganizationId: createResp.Id,
						Limit:          10,
						Offset:         0,
					})
					assert.NoError(err)

					assert.Len(orgUsers.Result, 1)
					assert.Equal(user.ID, orgUsers.Result[0].UserId)
					assert.Equal(user.Email, orgUsers.Result[0].Email)
					assert.Equal(addOrgUser.OrganizationUser.IsAdmin, orgUsers.Result[0].IsAdmin)
				})

				t.Run("Update user", func(t *testing.T) {
					assert := require.New(t)

					updOrgUser := &pb.UpdateOrganizationUserRequest{
						OrganizationUser: &pb.OrganizationUser{
							OrganizationId: createResp.Id,
							UserId:         user.ID,
							IsAdmin:        !addOrgUser.OrganizationUser.IsAdmin,
							IsDeviceAdmin:  !addOrgUser.OrganizationUser.IsDeviceAdmin,
							IsGatewayAdmin: !addOrgUser.OrganizationUser.IsGatewayAdmin,
						},
					}
					_, err := api.UpdateUser(context.Background(), updOrgUser)
					assert.NoError(err)

					orgUsers, err := api.ListUsers(context.Background(), &pb.ListOrganizationUsersRequest{
						OrganizationId: createResp.Id,
						Limit:          10,
						Offset:         0,
					})
					assert.NoError(err)

					assert.Len(orgUsers.Result, 1)
					assert.Equal(user.ID, orgUsers.Result[0].UserId)
					assert.Equal(user.Email, orgUsers.Result[0].Email)
					assert.Equal(updOrgUser.OrganizationUser.IsAdmin, orgUsers.Result[0].IsAdmin)
					assert.Equal(updOrgUser.OrganizationUser.IsDeviceAdmin, orgUsers.Result[0].IsDeviceAdmin)
					assert.Equal(updOrgUser.OrganizationUser.IsGatewayAdmin, orgUsers.Result[0].IsGatewayAdmin)
				})

				t.Run("Remove user from organization", func(t *testing.T) {
					t.Run("Remove self", func(t *testing.T) {
						assert := require.New(t)
						validator.returnUser = user

						delOrgUser := &pb.DeleteOrganizationUserRequest{
							OrganizationId: createResp.Id,
							UserId:         user.ID,
						}
						_, err := api.DeleteUser(context.Background(), delOrgUser)
						assert.Equal(codes.InvalidArgument, grpc.Code(err))
					})

					t.Run("Remove other", func(t *testing.T) {
						assert := require.New(t)
						validator.returnUser = adminUser

						delOrgUser := &pb.DeleteOrganizationUserRequest{
							OrganizationId: createResp.Id,
							UserId:         user.ID,
						}
						_, err := api.DeleteUser(context.Background(), delOrgUser)
						assert.NoError(err)

						_, err = api.DeleteUser(context.Background(), delOrgUser)
						assert.Equal(codes.NotFound, grpc.Code(err))
					})
				})
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)
			validator.returnUser = adminUser

			updateOrg := &pb.UpdateOrganizationRequest{
				Organization: &pb.Organization{
					Id:              createResp.Id,
					Name:            "anotherorg",
					DisplayName:     "Display Name 2",
					CanHaveGateways: false,
					MaxGatewayCount: 4,
					MaxDeviceCount:  20,
				},
			}
			_, err := api.Update(context.Background(), updateOrg)
			assert.NoError(err)

			orgUpd, err := api.Get(context.Background(), &pb.GetOrganizationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			createReq.Organization.Id = createResp.Id
			assert.Equal(updateOrg.Organization, orgUpd.Organization)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)
			validator.returnUser = adminUser

			_, err := api.Delete(context.Background(), &pb.DeleteOrganizationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Delete(context.Background(), &pb.DeleteOrganizationRequest{
				Id: createResp.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
