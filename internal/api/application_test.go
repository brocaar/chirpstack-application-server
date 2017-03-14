package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestApplicationAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}
		api := NewApplicationAPI(lsCtx, validator)
		apiuser := NewUserAPI(lsCtx, validator)

		Convey("When creating an application", func() {
			createResp, err := api.Create(ctx, &pb.CreateApplicationRequest{
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

			Convey("Then listing the applications as an admin returns a single item", func() {
				validator.returnIsAdmin = true

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
		})
	})
}
