package external

import (
	"testing"

	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestServiceProfile() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{returnSubject: "user"}
	api := NewServiceProfileServiceAPI(validator)

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

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

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		createReq := pb.CreateServiceProfileRequest{
			ServiceProfile: &pb.ServiceProfile{
				Name:                   "test-sp",
				OrganizationId:         org.ID,
				NetworkServerId:        n.ID,
				UlRate:                 100,
				UlBucketSize:           10,
				UlRatePolicy:           pb.RatePolicy_MARK,
				DlRate:                 200,
				DlBucketSize:           20,
				DlRatePolicy:           pb.RatePolicy_DROP,
				AddGwMetadata:          true,
				DevStatusReqFreq:       4,
				ReportDevStatusBattery: true,
				ReportDevStatusMargin:  true,
				DrMin:                  3,
				DrMax:                  5,
				PrAllowed:              true,
				HrAllowed:              true,
				RaAllowed:              true,
				NwkGeoLoc:              true,
				TargetPer:              10,
				MinGwDiversity:         3,
				GwsPrivate:             true,
			},
		}

		createResp, err := api.Create(context.Background(), &createReq)
		assert.NoError(err)

		assert.NotEqual("", createResp.Id)
		assert.NotEqual(uuid.Nil.String(), createResp.Id)

		// set network-server mock
		nsCreate := <-nsClient.CreateServiceProfileChan
		nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
			ServiceProfile: nsCreate.ServiceProfile,
		}

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			getResp, err := api.Get(context.Background(), &pb.GetServiceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			createReq.ServiceProfile.Id = createResp.Id
			assert.Equal(createReq.ServiceProfile, getResp.ServiceProfile)
		})

		t.Run("List", func(t *testing.T) {
			t.Run("As global admin", func(t *testing.T) {
				validator.returnUser = adminUser

				t.Run("No filters", func(t *testing.T) {
					assert := require.New(t)

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					assert.NoError(err)
					assert.EqualValues(1, listResp.TotalCount)
					assert.Len(listResp.Result, 1)
				})

				t.Run("Filter by organization ID", func(t *testing.T) {
					assert := require.New(t)

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit:          10,
						OrganizationId: org.ID,
					})
					assert.NoError(err)
					assert.EqualValues(1, listResp.TotalCount)
					assert.Len(listResp.Result, 1)

					listResp, err = api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit:          10,
						OrganizationId: org.ID + 1,
					})
					assert.NoError(err)
					assert.EqualValues(0, listResp.TotalCount)
					assert.Len(listResp.Result, 0)
				})

				t.Run("Filter by organization ID + network-server ID", func(t *testing.T) {
					assert := require.New(t)

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit:           10,
						OrganizationId:  org.ID,
						NetworkServerId: n.ID,
					})
					assert.NoError(err)
					assert.EqualValues(1, listResp.TotalCount)
					assert.Len(listResp.Result, 1)

					listResp, err = api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit:           10,
						OrganizationId:  org.ID,
						NetworkServerId: n.ID + 1,
					})
					assert.NoError(err)
					assert.EqualValues(0, listResp.TotalCount)
					assert.Len(listResp.Result, 0)
				})
			})

			t.Run("As organization user", func(t *testing.T) {
				assert := require.New(t)
				validator.returnUser = user

				user := storage.User{
					IsActive: true,
					Email:    "foo@bar.com",
				}

				err := storage.CreateUser(context.Background(), storage.DB(), &user)
				assert.NoError(err)
				assert.NoError(storage.CreateOrganizationUser(context.Background(), storage.DB(), org.ID, user.ID, false, false, false))

				t.Run("No filters", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = user

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					assert.NoError(err)
					assert.EqualValues(1, listResp.TotalCount)
					assert.Len(listResp.Result, 1)
				})

				t.Run("Invalid username", func(t *testing.T) {
					assert := require.New(t)

					user := storage.User{
						IsActive: true,
						Email:    "foo2@bar.com",
					}

					err := storage.CreateUser(context.Background(), storage.DB(), &user)

					validator.returnUser = user

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					assert.NoError(err)
					assert.EqualValues(0, listResp.TotalCount)
					assert.Len(listResp.Result, 0)
				})
			})

			t.Run("As API user", func(t *testing.T) {
				assert := require.New(t)
				validator.returnUser = user

				key := storage.APIKey{
					Name:    "foo@bar.com",
					IsAdmin: true,
				}
				_, err := storage.CreateAPIKey(context.Background(), storage.DB(), &key)
				assert.NoError(err)
				assert.NoError(storage.CreateOrganizationUser(context.Background(), storage.DB(), org.ID, user.ID, false, false, false))

				t.Run("No filters", func(t *testing.T) {
					assert := require.New(t)
					validator.returnAPIKeyID = key.ID

					listResp, err := api.List(context.Background(), &pb.ListServiceProfileRequest{
						Limit: 10,
					})
					assert.NoError(err)
					assert.EqualValues(1, listResp.TotalCount)
					assert.Len(listResp.Result, 1)
				})
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			updateReq := pb.UpdateServiceProfileRequest{
				ServiceProfile: &pb.ServiceProfile{
					Id:                     createResp.Id,
					Name:                   "updated-sp",
					OrganizationId:         org.ID,
					NetworkServerId:        n.ID,
					UlRate:                 200,
					UlBucketSize:           20,
					UlRatePolicy:           pb.RatePolicy_DROP,
					DlRate:                 300,
					DlBucketSize:           30,
					DlRatePolicy:           pb.RatePolicy_MARK,
					AddGwMetadata:          true,
					DevStatusReqFreq:       5,
					ReportDevStatusBattery: true,
					ReportDevStatusMargin:  true,
					DrMin:                  2,
					DrMax:                  4,
					PrAllowed:              true,
					HrAllowed:              true,
					RaAllowed:              true,
					NwkGeoLoc:              true,
					TargetPer:              20,
					MinGwDiversity:         4,
					GwsPrivate:             false,
				},
			}

			_, err := api.Update(context.Background(), &updateReq)
			assert.NoError(err)

			nsUpdate := <-nsClient.UpdateServiceProfileChan
			nsClient.GetServiceProfileResponse = ns.GetServiceProfileResponse{
				ServiceProfile: nsUpdate.ServiceProfile,
			}

			getResp, err := api.Get(context.Background(), &pb.GetServiceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(updateReq.ServiceProfile, getResp.ServiceProfile)
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteServiceProfileRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Delete(context.Background(), &pb.DeleteServiceProfileRequest{
				Id: createResp.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
