package api

import (
	"testing"

	"github.com/gofrs/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestApplicationAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	nsClient := test.NewNetworkServerClient()

	config.C.PostgreSQL.DB = db
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with an organization and an api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewApplicationAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		Convey("When creating an application", func() {
			createResp, err := api.Create(ctx, &pb.CreateApplicationRequest{
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
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp.Id, ShouldBeGreaterThan, 0)

			Convey("Then the application has been created", func() {
				app, err := api.Get(ctx, &pb.GetApplicationRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(app, ShouldResemble, &pb.GetApplicationResponse{
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
				})
			})

			Convey("Given an extra application belonging to a different organization", func() {
				org2 := storage.Organization{
					Name: "test-org-2",
				}
				So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org2), ShouldBeNil)

				sp2 := storage.ServiceProfile{
					Name:            "test-sp2",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp2), ShouldBeNil)

				app2 := storage.Application{
					OrganizationID:   org2.ID,
					Name:             "test-app-2",
					ServiceProfileID: spID,
				}
				So(storage.CreateApplication(config.C.PostgreSQL.DB, &app2), ShouldBeNil)

				Convey("When listing all applications", func() {
					Convey("Then all applications are visible to an admin user", func() {
						validator.returnIsAdmin = true
						apps, err := api.List(ctx, &pb.ListApplicationRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 2)
						So(apps.Result, ShouldHaveLength, 2)
						So(apps.Result[0], ShouldResemble, &pb.ApplicationListItem{
							OrganizationId:     org.ID,
							Id:                 createResp.Id,
							Name:               "test-app",
							Description:        "A test application",
							ServiceProfileId:   spID.String(),
							ServiceProfileName: sp.Name,
						})
					})
				})

				Convey("When listing all applications as an admin given an organization ID", func() {
					validator.returnIsAdmin = true
					Convey("Then only the applications for that organization are returned", func() {
						apps, err := api.List(ctx, &pb.ListApplicationRequest{
							Limit:          10,
							Offset:         0,
							OrganizationId: org2.ID,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 1)
						So(apps.Result, ShouldHaveLength, 1)
						So(apps.Result[0].OrganizationId, ShouldEqual, org2.ID)
					})
				})
			})

			Convey("When updating the application", func() {
				_, err := api.Update(ctx, &pb.UpdateApplicationRequest{
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
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the application has been updated", func() {
					app, err := api.Get(ctx, &pb.GetApplicationRequest{
						Id: createResp.Id,
					})
					So(err, ShouldBeNil)
					So(app, ShouldResemble, &pb.GetApplicationResponse{
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
					})
				})
			})

			Convey("When deleting the application", func() {
				_, err := api.Delete(ctx, &pb.DeleteApplicationRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the application has been deleted", func() {
					apps, err := api.List(ctx, &pb.ListApplicationRequest{Limit: 10})
					So(err, ShouldBeNil)
					So(apps.TotalCount, ShouldEqual, 0)
					So(apps.Result, ShouldHaveLength, 0)
				})
			})

			Convey("When creating a HTTP integration", func() {
				req := pb.CreateHTTPIntegrationRequest{
					Integration: &pb.HTTPIntegration{
						ApplicationId: createResp.Id,
						Headers: []*pb.HTTPIntegrationHeader{
							{Key: "Foo", Value: "bar"},
						},
						UplinkDataUrl:        "http://up",
						JoinNotificationUrl:  "http://join",
						AckNotificationUrl:   "http://ack",
						ErrorNotificationUrl: "http://error",
					},
				}
				_, err := api.CreateHTTPIntegration(ctx, &req)
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the integration can be retrieved", func() {
					i, err := api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(req.Integration, ShouldResemble, i.Integration)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
				})

				Convey("Then the integrations can be listed", func() {
					resp, err := api.ListIntegrations(ctx, &pb.ListIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(resp.TotalCount, ShouldEqual, 1)
					So(resp.Result[0], ShouldResemble, &pb.IntegrationListItem{
						Kind: pb.IntegrationKind_HTTP,
					})
				})

				Convey("Then the integration can be updated", func() {
					req := pb.UpdateHTTPIntegrationRequest{
						Integration: &pb.HTTPIntegration{
							ApplicationId:        createResp.Id,
							UplinkDataUrl:        "http://up2",
							JoinNotificationUrl:  "http://join2",
							AckNotificationUrl:   "http://ack2",
							ErrorNotificationUrl: "http://error",
						},
					}
					_, err := api.UpdateHTTPIntegration(ctx, &req)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					i, err := api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(i.Integration, ShouldResemble, req.Integration)
				})

				Convey("Then the integration can be deleted", func() {
					_, err := api.DeleteHTTPIntegration(ctx, &pb.DeleteHTTPIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					_, err = api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})
			})

			Convey("When creating an InfluxDB integration", func() {
				createReq := pb.CreateInfluxDBIntegrationRequest{
					Integration: &pb.InfluxDBIntegration{
						ApplicationId:       createResp.Id,
						Endpoint:            "http://localhost:8086/write",
						Db:                  "loraserver",
						Username:            "username",
						Password:            "password",
						RetentionPolicyName: "DEFAULT",
						Precision:           pb.InfluxDBPrecision_MS,
					},
				}
				_, err := api.CreateInfluxDBIntegration(ctx, &createReq)
				So(err, ShouldBeNil)

				Convey("Then the integration can be retrieved", func() {
					i, err := api.GetInfluxDBIntegration(ctx, &pb.GetInfluxDBIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					So(err, ShouldBeNil)
					So(i.Integration, ShouldResemble, createReq.Integration)
				})

				Convey("Then the integrations can be listed", func() {
					resp, err := api.ListIntegrations(ctx, &pb.ListIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(resp.TotalCount, ShouldEqual, 1)
					So(resp.Result[0].Kind, ShouldEqual, pb.IntegrationKind_INFLUXDB)
				})

				Convey("Then the integration can be updated", func() {
					updateReq := pb.UpdateInfluxDBIntegrationRequest{
						Integration: &pb.InfluxDBIntegration{
							ApplicationId:       createResp.Id,
							Endpoint:            "http://localhost:8086/write2",
							Db:                  "loraserver2",
							Username:            "username2",
							Password:            "password2",
							RetentionPolicyName: "CUSTOM",
							Precision:           pb.InfluxDBPrecision_S,
						},
					}
					_, err := api.UpdateInfluxDBIntegration(ctx, &updateReq)
					So(err, ShouldBeNil)

					i, err := api.GetInfluxDBIntegration(ctx, &pb.GetInfluxDBIntegrationRequest{
						ApplicationId: createResp.Id,
					})
					So(err, ShouldBeNil)
					So(i.Integration, ShouldResemble, updateReq.Integration)
				})

				Convey("Then the integration can be deleted", func() {
					_, err := api.DeleteInfluxDBIntegration(ctx, &pb.DeleteInfluxDBIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					_, err = api.GetInfluxDBIntegration(ctx, &pb.GetInfluxDBIntegrationRequest{ApplicationId: createResp.Id})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})
			})
		})
	})
}
