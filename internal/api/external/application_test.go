package external

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/external/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mqtt"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
)

func (ts *APITestSuite) TestApplication() {
	assert := require.New(ts.T())

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	validator := &TestValidator{returnSubject: "user"}
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
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	org2 := storage.Organization{
		Name: "test-org-2",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org2))

	sp2 := storage.ServiceProfile{
		Name:            "test-sp2",
		NetworkServerID: n.ID,
		OrganizationID:  org2.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp2))
	spID2, err := uuid.FromBytes(sp2.ServiceProfile.Id)
	assert.NoError(err)

	user := storage.User{
		Email:    "foo@bar.com",
		IsActive: true,
		IsAdmin:  true,
	}
	assert.NoError(storage.CreateUser(context.Background(), storage.DB(), &user))

	ts.T().Run("Create with service-profile under different organization", func(t *testing.T) {
		assert := require.New(t)
		_, err := api.Create(context.Background(), &pb.CreateApplicationRequest{
			Application: &pb.Application{
				OrganizationId:       org.ID,
				Name:                 "test-app",
				Description:          "A test application",
				ServiceProfileId:     spID2.String(),
				PayloadCodec:         "CUSTOM_JS",
				PayloadEncoderScript: "Encode() {}",
				PayloadDecoderScript: "Decode() {}",
			},
		})
		assert.Equal(codes.InvalidArgument, grpc.Code(err))
	})

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

			app2 := storage.Application{
				OrganizationID:   org2.ID,
				Name:             "test-app-2",
				ServiceProfileID: spID,
			}
			assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app2))

			t.Run("List", func(t *testing.T) {
				t.Run("As global admin", func(t *testing.T) {
					assert := require.New(t)
					validator.returnUser = user

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
					validator.returnUser = user

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

		t.Run("Update with service-profile under different organization", func(t *testing.T) {
			assert := require.New(t)

			_, err := api.Update(context.Background(), &pb.UpdateApplicationRequest{
				Application: &pb.Application{
					Id:                   createResp.Id,
					Name:                 "test-app-updated",
					Description:          "An updated test description",
					ServiceProfileId:     spID2.String(),
					PayloadCodec:         "CUSTOM_JS",
					PayloadEncoderScript: "Encode2() {}",
					PayloadDecoderScript: "Decode2() {}",
				},
			})
			assert.Equal(codes.InvalidArgument, grpc.Code(err))
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
						EventEndpointUrl: "http://event",
						Marshaler:        pb.Marshaler_PROTOBUF,

						UplinkDataUrl:           "http://up",
						JoinNotificationUrl:     "http://join",
						AckNotificationUrl:      "http://ack",
						ErrorNotificationUrl:    "http://error",
						StatusNotificationUrl:   "http://status",
						LocationNotificationUrl: "http://location",
						TxAckNotificationUrl:    "http://txack",
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
							EventEndpointUrl: "http://event2",
							Marshaler:        pb.Marshaler_JSON,

							ApplicationId:           createResp.Id,
							UplinkDataUrl:           "http://up2",
							JoinNotificationUrl:     "http://join2",
							AckNotificationUrl:      "http://ack2",
							ErrorNotificationUrl:    "http://error",
							StatusNotificationUrl:   "http://status2",
							LocationNotificationUrl: "http://location2",
							TxAckNotificationUrl:    "http://txack2",
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
						Version:             pb.InfluxDBVersion_INFLUXDB_1,
						Db:                  "chirpstack",
						Username:            "username",
						Password:            "password",
						RetentionPolicyName: "DEFAULT",
						Precision:           pb.InfluxDBPrecision_MS,
						Token:               "test-token",
						Organization:        "test-org",
						Bucket:              "test-bucket",
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
							Version:             pb.InfluxDBVersion_INFLUXDB_2,
							Db:                  "chirpstack2",
							Username:            "username2",
							Password:            "password2",
							RetentionPolicyName: "CUSTOM",
							Precision:           pb.InfluxDBPrecision_S,
							Token:               "test-token-2",
							Organization:        "test-org-2",
							Bucket:              "test-bucket-2",
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

		t.Run("MyDevicesIntegration", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreateMyDevicesIntegrationRequest{
					Integration: &pb.MyDevicesIntegration{
						ApplicationId: createResp.Id,
						Endpoint:      "https://localhost:1234",
					},
				}
				_, err := api.CreateMyDevicesIntegration(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetMyDevicesIntegration(context.Background(), &pb.GetMyDevicesIntegrationRequest{
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
					assert.Equal(pb.IntegrationKind_MYDEVICES, resp.Result[0].Kind)
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateMyDevicesIntegrationRequest{
						Integration: &pb.MyDevicesIntegration{
							ApplicationId: createResp.Id,
							Endpoint:      "http://localhost:12345",
						},
					}
					_, err := api.UpdateMyDevicesIntegration(context.Background(), &updateReq)
					assert.NoError(err)

					i, err := api.GetMyDevicesIntegration(context.Background(), &pb.GetMyDevicesIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(updateReq.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteMyDevicesIntegration(context.Background(), &pb.DeleteMyDevicesIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					_, err = api.GetMyDevicesIntegration(context.Background(), &pb.GetMyDevicesIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("LoRaCloudIntegration", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreateLoRaCloudIntegrationRequest{
					Integration: &pb.LoRaCloudIntegration{
						ApplicationId:                createResp.Id,
						Geolocation:                  true,
						GeolocationToken:             "ab123",
						GeolocationBufferTtl:         10,
						GeolocationMinBufferSize:     4,
						GeolocationTdoa:              true,
						GeolocationRssi:              true,
						GeolocationGnss:              true,
						GeolocationGnssPayloadField:  "lr1110_gnss",
						GeolocationGnssUseRxTime:     true,
						GeolocationWifi:              true,
						GeolocationWifiPayloadField:  "access_points",
						Das:                          true,
						DasToken:                     "ba321",
						DasModemPort:                 199,
						DasGnssPort:                  198,
						DasGnssUseRxTime:             true,
						DasStreamingGeolocWorkaround: true,
					},
				}
				_, err := api.CreateLoRaCloudIntegration(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetLoRaCloudIntegration(context.Background(), &pb.GetLoRaCloudIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(createReq.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(pb.IntegrationKind_LORACLOUD, resp.Result[0].Kind)
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdateLoRaCloudIntegrationRequest{
						Integration: &pb.LoRaCloudIntegration{
							ApplicationId:                createResp.Id,
							Geolocation:                  true,
							GeolocationToken:             "123ab",
							GeolocationBufferTtl:         4,
							GeolocationMinBufferSize:     10,
							GeolocationTdoa:              true,
							GeolocationRssi:              true,
							GeolocationGnss:              true,
							GeolocationGnssPayloadField:  "lr1110_gnss_updated",
							GeolocationGnssUseRxTime:     true,
							GeolocationWifi:              true,
							GeolocationWifiPayloadField:  "access_points_updated",
							Das:                          true,
							DasToken:                     "321ba",
							DasModemPort:                 189,
							DasGnssPort:                  188,
							DasGnssUseRxTime:             false,
							DasStreamingGeolocWorkaround: false,
						},
					}

					_, err := api.UpdateLoRaCloudIntegration(context.Background(), &updateReq)
					assert.NoError(err)

					i, err := api.GetLoRaCloudIntegration(context.Background(), &pb.GetLoRaCloudIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(updateReq.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteLoRaCloudIntegration(context.Background(), &pb.DeleteLoRaCloudIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					_, err = api.GetLoRaCloudIntegration(context.Background(), &pb.GetLoRaCloudIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("GCPPubSub", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				req := pb.CreateGCPPubSubIntegrationRequest{
					Integration: &pb.GCPPubSubIntegration{
						ApplicationId:   createResp.Id,
						Marshaler:       pb.Marshaler_PROTOBUF,
						CredentialsFile: "file",
						ProjectId:       "test-project",
						TopicName:       "test-topic",
					},
				}
				_, err := api.CreateGCPPubSubIntegration(context.Background(), &req)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetGCPPubSubIntegration(context.Background(), &pb.GetGCPPubSubIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(&pb.IntegrationListItem{
						Kind: pb.IntegrationKind_GCP_PUBSUB,
					}, resp.Result[0])
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					req := pb.UpdateGCPPubSubIntegrationRequest{
						Integration: &pb.GCPPubSubIntegration{
							ApplicationId:   createResp.Id,
							Marshaler:       pb.Marshaler_JSON,
							CredentialsFile: "file-update",
							ProjectId:       "test-project-updated",
							TopicName:       "test-topic-updated",
						},
					}
					_, err := api.UpdateGCPPubSubIntegration(context.Background(), &req)
					assert.NoError(err)

					i, err := api.GetGCPPubSubIntegration(context.Background(), &pb.GetGCPPubSubIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteGCPPubSubIntegration(context.Background(), &pb.DeleteGCPPubSubIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					_, err = api.DeleteGCPPubSubIntegration(context.Background(), &pb.DeleteGCPPubSubIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("AWSSNS", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				req := pb.CreateAWSSNSIntegrationRequest{
					Integration: &pb.AWSSNSIntegration{
						ApplicationId:   createResp.Id,
						Marshaler:       pb.Marshaler_PROTOBUF,
						Region:          "eu-west-1",
						AccessKeyId:     "key-id-1",
						SecretAccessKey: "secret-key",
						TopicArn:        "test-topic",
					},
				}
				_, err := api.CreateAWSSNSIntegration(context.Background(), &req)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetAWSSNSIntegration(context.Background(), &pb.GetAWSSNSIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(&pb.IntegrationListItem{
						Kind: pb.IntegrationKind_AWS_SNS,
					}, resp.Result[0])
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					req := pb.UpdateAWSSNSIntegrationRequest{
						Integration: &pb.AWSSNSIntegration{
							ApplicationId:   createResp.Id,
							Marshaler:       pb.Marshaler_JSON,
							Region:          "eu-west-2",
							AccessKeyId:     "key-id-2",
							SecretAccessKey: "secret-key-updated",
							TopicArn:        "test-topic-updated",
						},
					}
					_, err := api.UpdateAWSSNSIntegration(context.Background(), &req)
					assert.NoError(err)

					i, err := api.GetAWSSNSIntegration(context.Background(), &pb.GetAWSSNSIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteAWSSNSIntegration(context.Background(), &pb.DeleteAWSSNSIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					_, err = api.DeleteAWSSNSIntegration(context.Background(), &pb.DeleteAWSSNSIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("AzureServiceBus", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				req := pb.CreateAzureServiceBusIntegrationRequest{
					Integration: &pb.AzureServiceBusIntegration{
						ApplicationId:    createResp.Id,
						Marshaler:        pb.Marshaler_PROTOBUF,
						ConnectionString: "test-connection-string",
						PublishName:      "test-queue",
					},
				}
				_, err := api.CreateAzureServiceBusIntegration(context.Background(), &req)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetAzureServiceBusIntegration(context.Background(), &pb.GetAzureServiceBusIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("List", func(t *testing.T) {
					assert := require.New(t)

					resp, err := api.ListIntegrations(context.Background(), &pb.ListIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					assert.EqualValues(1, resp.TotalCount)
					assert.Equal(&pb.IntegrationListItem{
						Kind: pb.IntegrationKind_AZURE_SERVICE_BUS,
					}, resp.Result[0])
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					req := pb.UpdateAzureServiceBusIntegrationRequest{
						Integration: &pb.AzureServiceBusIntegration{
							ApplicationId:    createResp.Id,
							Marshaler:        pb.Marshaler_JSON,
							ConnectionString: "test-connection-string-updated",
							PublishName:      "test-queue-updated",
						},
					}

					_, err := api.UpdateAzureServiceBusIntegration(context.Background(), &req)
					assert.NoError(err)

					i, err := api.GetAzureServiceBusIntegration(context.Background(), &pb.GetAzureServiceBusIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(req.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeleteAzureServiceBusIntegration(context.Background(), &pb.DeleteAzureServiceBusIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)

					_, err = api.DeleteAzureServiceBusIntegration(context.Background(), &pb.DeleteAzureServiceBusIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("PilotThings", func(t *testing.T) {
			t.Run("Create", func(t *testing.T) {
				assert := require.New(t)

				createReq := pb.CreatePilotThingsIntegrationRequest{
					Integration: &pb.PilotThingsIntegration{
						ApplicationId: createResp.Id,
						Server:        "http://localhost:1234",
						Token:         "very secure token",
					},
				}
				_, err := api.CreatePilotThingsIntegration(context.Background(), &createReq)
				assert.NoError(err)

				t.Run("Get", func(t *testing.T) {
					assert := require.New(t)

					i, err := api.GetPilotThingsIntegration(context.Background(), &pb.GetPilotThingsIntegrationRequest{
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
					assert.Equal(pb.IntegrationKind_PILOT_THINGS, resp.Result[0].Kind)
				})

				t.Run("Update", func(t *testing.T) {
					assert := require.New(t)

					updateReq := pb.UpdatePilotThingsIntegrationRequest{
						Integration: &pb.PilotThingsIntegration{
							ApplicationId: createResp.Id,
							Server:        "https://localhost:12345",
							Token:         "less secure token",
						},
					}
					_, err := api.UpdatePilotThingsIntegration(context.Background(), &updateReq)
					assert.NoError(err)

					i, err := api.GetPilotThingsIntegration(context.Background(), &pb.GetPilotThingsIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					assert.NoError(err)
					assert.Equal(updateReq.Integration, i.Integration)
				})

				t.Run("Delete", func(t *testing.T) {
					assert := require.New(t)

					_, err := api.DeletePilotThingsIntegration(context.Background(), &pb.DeletePilotThingsIntegrationRequest{ApplicationId: createResp.Id})
					assert.NoError(err)

					_, err = api.GetPilotThingsIntegration(context.Background(), &pb.GetPilotThingsIntegrationRequest{ApplicationId: createResp.Id})
					assert.Equal(codes.NotFound, grpc.Code(err))
				})
			})
		})

		t.Run("MQTT", func(t *testing.T) {
			t.Run("Generate certificate", func(t *testing.T) {
				assert := require.New(t)

				app, err := storage.GetApplication(context.Background(), ts.tx.Tx, createResp.Id)
				assert.NoError(err)
				assert.Len(app.MQTTTLSCert, 0)

				conf := test.GetConfig()
				conf.ApplicationServer.Integration.MQTT.Client.CACert = "../../test/ca_cert.pem"
				conf.ApplicationServer.Integration.MQTT.Client.CAKey = "../../test/ca_private.pem"
				conf.ApplicationServer.Integration.MQTT.Client.ClientCertLifetime = time.Hour

				assert.NoError(mqtt.Setup(conf))

				resp, err := api.GenerateMQTTIntegrationClientCertificate(context.Background(), &pb.GenerateMQTTIntegrationClientCertificateRequest{
					ApplicationId: createResp.Id,
				})
				assert.NoError(err)

				assert.NotEqual(0, len(resp.TlsCert))
				assert.NotEqual(0, len(resp.TlsKey))

				b, err := ioutil.ReadFile(conf.ApplicationServer.Integration.MQTT.Client.CACert)
				assert.NoError(err)

				assert.Equal(string(b), string(resp.CaCert))
				assert.NotNil(resp.ExpiresAt)

				app, err = storage.GetApplication(context.Background(), ts.tx.Tx, createResp.Id)
				assert.NoError(err)
				assert.NotEqual(0, len(app.MQTTTLSCert))
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
