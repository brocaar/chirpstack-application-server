package api

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan/backend"
)

func TestApplicationAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	nsClient := test.NewNetworkServerClient()

	common.DB = db
	common.NetworkServer = nsClient

	Convey("Given a clean database with an organization and an api instance", t, func() {
		test.MustResetDB(common.DB)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewApplicationAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(common.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(common.DB, &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		Convey("When creating an application", func() {
			createResp, err := api.Create(ctx, &pb.CreateApplicationRequest{
				OrganizationID:   org.ID,
				Name:             "test-app",
				Description:      "A test application",
				ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
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
					OrganizationID:   org.ID,
					Id:               createResp.Id,
					Name:             "test-app",
					Description:      "A test application",
					ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
				})
			})

			Convey("Given an extra application belonging to a different organization", func() {
				org2 := storage.Organization{
					Name: "test-org-2",
				}
				So(storage.CreateOrganization(common.DB, &org2), ShouldBeNil)

				sp2 := storage.ServiceProfile{
					Name:            "test-sp2",
					NetworkServerID: n.ID,
					OrganizationID:  org.ID,
				}
				So(storage.CreateServiceProfile(common.DB, &sp2), ShouldBeNil)

				app2 := storage.Application{
					OrganizationID:   org2.ID,
					Name:             "test-app-2",
					ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
				}
				So(storage.CreateApplication(common.DB, &app2), ShouldBeNil)

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
						So(apps.Result[0], ShouldResemble, &pb.GetApplicationResponse{
							OrganizationID:   org.ID,
							Id:               createResp.Id,
							Name:             "test-app",
							Description:      "A test application",
							ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
						})
					})
				})

				Convey("When listing all applications as an admin given an organization ID", func() {
					validator.returnIsAdmin = true
					Convey("Then only the applications for that organization are returned", func() {
						apps, err := api.List(ctx, &pb.ListApplicationRequest{
							Limit:          10,
							Offset:         0,
							OrganizationID: org2.ID,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 1)
						So(apps.Result, ShouldHaveLength, 1)
						So(apps.Result[0].OrganizationID, ShouldEqual, org2.ID)
					})
				})
			})

			Convey("When updating the application", func() {
				_, err := api.Update(ctx, &pb.UpdateApplicationRequest{
					OrganizationID:   org.ID,
					Id:               createResp.Id,
					Name:             "test-app-updated",
					Description:      "An updated test description",
					ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
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
						OrganizationID:   org.ID,
						Id:               createResp.Id,
						Name:             "test-app-updated",
						Description:      "An updated test description",
						ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
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
				integration := pb.HTTPIntegration{
					Id: createResp.Id,
					Headers: []*pb.HTTPIntegrationHeader{
						{Key: "Foo", Value: "bar"},
					},
					DataUpURL:            "http://up",
					JoinNotificationURL:  "http://join",
					AckNotificationURL:   "http://ack",
					ErrorNotificationURL: "http://error",
				}
				_, err := api.CreateHTTPIntegration(ctx, &integration)
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the integration can be retrieved", func() {
					i, err := api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{Id: createResp.Id})
					So(err, ShouldBeNil)
					So(*i, ShouldResemble, integration)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
				})

				Convey("Then the integrations can be listed", func() {
					resp, err := api.ListIntegrations(ctx, &pb.ListIntegrationRequest{Id: createResp.Id})
					So(err, ShouldBeNil)
					So(resp.Kinds, ShouldResemble, []pb.IntegrationKind{pb.IntegrationKind_HTTP})
				})

				Convey("Then the integration can be updated", func() {
					integration.DataUpURL = "http://up2"
					integration.JoinNotificationURL = "http://join2"
					integration.AckNotificationURL = "http://ack2"
					integration.ErrorNotificationURL = "http://error"
					_, err := api.UpdateHTTPIntegration(ctx, &integration)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					i, err := api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{Id: createResp.Id})
					So(err, ShouldBeNil)
					So(*i, ShouldResemble, integration)
				})

				Convey("Then the integration can be deleted", func() {
					_, err := api.DeleteHTTPIntegration(ctx, &pb.DeleteIntegrationRequest{Id: createResp.Id})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					_, err = api.GetHTTPIntegration(ctx, &pb.GetHTTPIntegrationRequest{Id: createResp.Id})
					So(err, ShouldNotBeNil)
					So(grpc.Code(err), ShouldEqual, codes.NotFound)
				})
			})
		})
	})
}
