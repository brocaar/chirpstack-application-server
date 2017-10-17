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
)

func TestApplicationAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with an organization and an api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		common.DB = db
		test.MustResetDB(common.DB)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewApplicationAPI(validator)
		apiuser := NewUserAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(common.DB, &org), ShouldBeNil)

		Convey("When creating an application", func() {
			createResp, err := api.Create(ctx, &pb.CreateApplicationRequest{
				OrganizationID:     org.ID,
				Name:               "test-app",
				Description:        "A test application",
				IsABP:              true,
				IsClassC:           true,
				RxDelay:            1,
				Rx1DROffset:        3,
				RxWindow:           pb.RXWindow_RX2,
				Rx2DR:              3,
				AdrInterval:        20,
				InstallationMargin: 5,
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
					OrganizationID:     org.ID,
					Id:                 createResp.Id,
					Name:               "test-app",
					Description:        "A test application",
					IsABP:              true,
					IsClassC:           true,
					RxDelay:            1,
					Rx1DROffset:        3,
					RxWindow:           pb.RXWindow_RX2,
					Rx2DR:              3,
					AdrInterval:        20,
					InstallationMargin: 5,
				})
			})

			Convey("Given an extra application belonging to a different organization", func() {
				org2 := storage.Organization{
					Name: "test-org-2",
				}
				So(storage.CreateOrganization(common.DB, &org2), ShouldBeNil)
				app2 := storage.Application{
					OrganizationID: org2.ID,
					Name:           "test-app-2",
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
							OrganizationID:     org.ID,
							Id:                 createResp.Id,
							Name:               "test-app",
							Description:        "A test application",
							IsABP:              true,
							IsClassC:           true,
							RxDelay:            1,
							Rx1DROffset:        3,
							RxWindow:           pb.RXWindow_RX2,
							Rx2DR:              3,
							AdrInterval:        20,
							InstallationMargin: 5,
						})
					})

					Convey("Then applications are only visible to users assigned to the application", func() {
						user := storage.User{Username: "testtest"}
						_, err := storage.CreateUser(common.DB, &user, "password123")
						So(err, ShouldBeNil)
						validator.returnIsAdmin = false
						validator.returnUsername = user.Username

						apps, err := api.List(ctx, &pb.ListApplicationRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 0)
						So(apps.Result, ShouldHaveLength, 0)

						So(storage.CreateUserForApplication(common.DB, createResp.Id, user.ID, false), ShouldBeNil)
						apps, err = api.List(ctx, &pb.ListApplicationRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 0)
						So(apps.Result, ShouldHaveLength, 0)
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

			Convey("When creating a user", func() {
				createUserReq := &pb.AddUserRequest{
					Username:   "username",
					Password:   "pass^^ord",
					IsAdmin:    true,
					IsActive:   true,
					SessionTTL: 180,
				}
				createRespUser, err := apiuser.Create(ctx, createUserReq)
				So(err, ShouldBeNil)
				So(createRespUser.Id, ShouldBeGreaterThan, 0)

				Convey("Then the user can be added to the application", func() {
					addReq := &pb.AddApplicationUserRequest{
						Id:      createResp.Id,
						UserID:  createRespUser.Id,
						IsAdmin: true,
					}
					noresp, err := api.AddUser(ctx, addReq)
					So(err, ShouldBeNil)
					So(noresp, ShouldNotBeNil)

					// Reused a lot below.
					getReq := &pb.ApplicationUserRequest{
						Id:     createResp.Id,
						UserID: createRespUser.Id,
					}

					Convey("Then listing the applications returns a single item", func() {
						validator.returnIsAdmin = false
						validator.returnUsername = createUserReq.Username

						apps, err := api.List(ctx, &pb.ListApplicationRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(validator.ctx, ShouldResemble, ctx)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(apps.Result, ShouldHaveLength, 1)
						So(apps.TotalCount, ShouldEqual, 1)
						So(apps.Result[0], ShouldResemble, &pb.GetApplicationResponse{
							OrganizationID:     org.ID,
							Id:                 createResp.Id,
							Name:               "test-app",
							Description:        "A test application",
							IsABP:              true,
							IsClassC:           true,
							RxDelay:            1,
							Rx1DROffset:        3,
							RxWindow:           pb.RXWindow_RX2,
							Rx2DR:              3,
							AdrInterval:        20,
							InstallationMargin: 5,
						})
					})

					Convey("Then the user can be accessed via application get", func() {
						getUserResp, err := api.GetUser(ctx, getReq)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(getUserResp.Username, ShouldEqual, createUserReq.Username)
						So(getUserResp.IsAdmin, ShouldEqual, createUserReq.IsAdmin)
					})

					Convey("Then the user can be accessed via get all users for application", func() {
						getUserList := &pb.ListApplicationUsersRequest{
							Id:     createResp.Id,
							Limit:  10,
							Offset: 0,
						}
						listAppResp, err := api.ListUsers(ctx, getUserList)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(listAppResp, ShouldNotBeNil)
						So(listAppResp.TotalCount, ShouldEqual, 1)
						So(listAppResp.Result, ShouldHaveLength, 1)
						So(listAppResp.Result[0].Username, ShouldEqual, createUserReq.Username)
						So(listAppResp.Result[0].IsAdmin, ShouldEqual, createUserReq.IsAdmin)
					})

					Convey("Then the user access to the application can be updated", func() {
						updReq := &pb.UpdateApplicationUserRequest{
							Id:      createResp.Id,
							UserID:  createRespUser.Id,
							IsAdmin: false,
						}
						empty, err := api.UpdateUser(ctx, updReq)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						Convey("Then the user can be accessed showing the new setting", func() {
							getUserResp, err := api.GetUser(ctx, getReq)
							So(err, ShouldBeNil)
							So(empty, ShouldNotBeNil)
							So(getUserResp.Username, ShouldEqual, createUserReq.Username)
							So(getUserResp.IsAdmin, ShouldEqual, updReq.IsAdmin)
						})
					})

					Convey("Then the user can be deleted from the application", func() {
						empty, err := api.DeleteUser(ctx, getReq)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(empty, ShouldNotBeNil)
						Convey("Then the user cannot be accessed via get", func() {
							getUserResp, err := api.GetUser(ctx, getReq)
							So(err, ShouldNotBeNil)
							So(getUserResp, ShouldBeNil)
						})
					})
				})
			})

			Convey("When updating the application", func() {
				_, err := api.Update(ctx, &pb.UpdateApplicationRequest{
					OrganizationID:     org.ID,
					Id:                 createResp.Id,
					Name:               "test-app-updated",
					Description:        "An updated test description",
					IsABP:              false,
					IsClassC:           true,
					RxDelay:            2,
					Rx1DROffset:        4,
					RxWindow:           pb.RXWindow_RX1,
					Rx2DR:              1,
					AdrInterval:        40,
					InstallationMargin: 10,
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
						OrganizationID:     org.ID,
						Id:                 createResp.Id,
						Name:               "test-app-updated",
						Description:        "An updated test description",
						IsABP:              false,
						IsClassC:           true,
						RxDelay:            2,
						Rx1DROffset:        4,
						RxWindow:           pb.RXWindow_RX1,
						Rx2DR:              1,
						AdrInterval:        40,
						InstallationMargin: 10,
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
