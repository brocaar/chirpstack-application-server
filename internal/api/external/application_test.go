package external

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func (ts *APITestSuite) TestApplication() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{}
	api := NewApplicationAPI(validator)

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	sp := storage.ServiceProfile{
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)
		createResp, err := api.Create(context.Background(), &pb.CreateApplicationRequest{
			Application: &pb.Application{
				OrganizationId:       org.ID,
				Name:                 "test-app",
				Description:          "A test application",
				ServiceProfileId:     spID.String(),
				PayloadCodec:         "CUSTOM_JS",
				PayloadEncoderScript: "Encode() {}",
				PayloadDecoderScript: "Decode() {}",
			},
		})
		assert.NoError(err)
		assert.True(createResp.Id > 0)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			app, err := api.Get(context.Background(), &pb.GetApplicationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)
			assert.Equal(&pb.GetApplicationResponse{
				Application: &pb.Application{
					OrganizationId:       org.ID,
					Id:                   createResp.Id,
					Name:                 "test-app",
					Description:          "A test application",
					ServiceProfileId:     spID.String(),
					PayloadCodec:         "CUSTOM_JS",
					PayloadEncoderScript: "Encode() {}",
					PayloadDecoderScript: "Decode() {}",
				},
			}, app)
		})

		t.Run("Create application for different organization", func(t *testing.T) {
			assert := require.New(t)

			org2 := storage.Organization{
				Name: "test-org-2",
			}
			assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

			sp2 := storage.ServiceProfile{
				Name:            "test-sp2",
				NetworkServerID: n.ID,
				OrganizationID:  org.ID,
			}
			assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp2))

			app2 := storage.Application{
				OrganizationID:   org2.ID,
				Name:             "test-app-2",
				ServiceProfileID: spID,
			}
			assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app2))

			t.Run("List", func(t *testing.T) {
				t.Run("As global admin", func(t *testing.T) {
					assert := require.New(t)
					validator.returnIsAdmin = true

					apps, err := api.List(context.Background(), &pb.ListApplicationRequest{
						Limit:  10,
						Offset: 0,
					})
					assert.NoError(err)

					assert.EqualValues(2, apps.TotalCount)
					assert.Len(apps.Result, 2)
					assert.Equal(&pb.ApplicationListItem{
						OrganizationId:     org.ID,
						Id:                 createResp.Id,
						Name:               "test-app",
						Description:        "A test application",
						ServiceProfileId:   spID.String(),
						ServiceProfileName: sp.Name,
					}, apps.Result[0])
				})

				t.Run("As global admin - with org id filter", func(t *testing.T) {
					assert := require.New(t)
					validator.returnIsAdmin = true

					apps, err := api.List(context.Background(), &pb.ListApplicationRequest{
						Limit:          10,
						Offset:         0,
						OrganizationId: org2.ID,
					})
					assert.NoError(err)

					assert.EqualValues(1, apps.TotalCount)
					assert.Len(apps.Result, 1)
					assert.Equal(org2.ID, apps.Result[0].OrganizationId)
				})
			})
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Update(context.Background(), &pb.UpdateApplicationRequest{
				Application: &pb.Application{
					Id:                   createResp.Id,
					Name:                 "test-app-updated",
					Description:          "An updated test description",
					ServiceProfileId:     spID.String(),
					PayloadCodec:         "CUSTOM_JS",
					PayloadEncoderScript: "Encode2() {}",
					PayloadDecoderScript: "Decode2() {}",
				},
			})
			assert.NoError(err)

			app, err := api.Get(context.Background(), &pb.GetApplicationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			assert.Equal(&pb.GetApplicationResponse{
				Application: &pb.Application{
					OrganizationId:       org.ID,
					Id:                   createResp.Id,
					Name:                 "test-app-updated",
					Description:          "An updated test description",
					ServiceProfileId:     spID.String(),
					PayloadCodec:         "CUSTOM_JS",
					PayloadEncoderScript: "Encode2() {}",
					PayloadDecoderScript: "Decode2() {}",
				},
			}, app)
		})

		t.Run("HTTPIntegration", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				req := pb.CreateHTTPIntegrationRequest{
					Integration: &pb.HTTPIntegration{
						ApplicationId: createResp.Id,
						Headers: []*pb.HTTPIntegrationHeader{
							{Key: "Foo", Value: "bar"},
						},
						UplinkDataUrl:           "http://up",
						JoinNotificationUrl:     "http://join",
						AckNotificationUrl:      "http://ack",
						ErrorNotificationUrl:    "http://error",
						StatusNotificationUrl:   "http://status",
						LocationNotificationUrl: "http://location",
					},
				}
				_, err := api.CreateHTTPIntegration(context.Background(), &req)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetHTTPIntegration(context.Background(), &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(&pb.IntegrationListItem{
						Kind: pb.IntegrationKind_HTTP,
					}, resp.Result[0])
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					req := pb.UpdateHTTPIntegrationRequest{
						Integration: &pb.HTTPIntegration{
							ApplicationId:           createResp.Id,
							UplinkDataUrl:           "http://up2",
							JoinNotificationUrl:     "http://join2",
							AckNotificationUrl:      "http://ack2",
							ErrorNotificationUrl:    "http://error",
							StatusNotificationUrl:   "http://status2",
							LocationNotificationUrl: "http://location2",
						},
					}
					_, err := api.UpdateHTTPIntegration(context.Background(), &req)
					assert.NoError(err)

					i, err := api.GetHTTPIntegration(context.Background(), &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteHTTPIntegration(context.Background(), &pb.DeleteHTTPIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					_, err = api.GetHTTPIntegration(context.Background(), &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("InfluxDBIntegration", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreateInfluxDBIntegrationRequest{
					Integration: &pb.InfluxDBIntegration{
						ApplicationId:       createResp.Id,
						Endpoint:            "http://localhost:8086/write",
						Db:                  "chirpstack",
						Username:            "username",
						Password:            "password",
						RetentionPolicyName: "DEFAULT",
						Precision:           pb.InfluxDBPrecision_MS,
					},
				}
				_, err := api.CreateInfluxDBIntegration(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetInfluxDBIntegration(context.Background(), &pb.GetInfluxDBIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(createReq.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(pb.IntegrationKind_INFLUXDB, resp.Result[0].Kind)
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateInfluxDBIntegrationRequest{
						Integration: &pb.InfluxDBIntegration{
							ApplicationId:       createResp.Id,
							Endpoint:            "http://localhost:8086/write2",
							Db:                  "chirpstack2",
							Username:            "username2",
							Password:            "password2",
							RetentionPolicyName: "CUSTOM",
							Precision:           pb.InfluxDBPrecision_S,
						},
					}
					_, err := api.UpdateInfluxDBIntegration(context.Background(), &updateReq)
					assert.NoError(err)

					i, err := api.GetInfluxDBIntegration(context.Background(), &pb.GetInfluxDBIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(updateReq.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteInfluxDBIntegration(context.Background(), &pb.DeleteInfluxDBIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					_, err = api.GetInfluxDBIntegration(context.Background(), &pb.GetInfluxDBIntegrationRequest{ApplicationId: createResp.Id})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("ThingsBoardIntegration", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreateThingsBoardIntegrationRequest{
					Integration: &pb.ThingsBoardIntegration{
						ApplicationId: createResp.Id,
						Server:        "http://localhost:1234",
					},
				}
				_, err := api.CreateThingsBoardIntegration(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetThingsBoardIntegration(context.Background(), &pb.GetThingsBoardIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(createReq.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(pb.IntegrationKind_THINGSBOARD, resp.Result[0].Kind)
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateThingsBoardIntegrationRequest{
						Integration: &pb.ThingsBoardIntegration{
							ApplicationId: createResp.Id,
							Server:        "https://localhost:12345",
						},
					}
					_, err := api.UpdateThingsBoardIntegration(context.Background(), &updateReq)
					assert.NoError(err)

					i, err := api.GetThingsBoardIntegration(context.Background(), &pb.GetThingsBoardIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(updateReq.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteThingsBoardIntegration(context.Background(), &pb.DeleteThingsBoardIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					_, err = api.GetThingsBoardIntegration(context.Background(), &pb.GetThingsBoardIntegrationRequest{ApplicationId: createResp.Id})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Delete(context.Background(), &pb.DeleteApplicationRequest{
				Id: createResp.Id,
			})
			assert.NoError(err)

			_, err = api.Get(context.Background(), &pb.GetApplicationRequest{
				Id: createResp.Id,
			})
			assert.Equal(codes.NotFound, grpc.Code(err))
		})
	})
}
