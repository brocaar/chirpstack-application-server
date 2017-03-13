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

func TestUserAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		ctx := context.Background()
		lsCtx := common.Context{DB: db}
		validator := &TestValidator{}
		api := NewUserAPI(lsCtx, validator)
		apiInternal := NewInternalUserAPI(lsCtx, validator)

		Convey("When creating a user", func() {
			validator.returnIsAdmin = true
			createReq := &pb.AddUserRequest{
				Username:   "username",
				Password:   "pass^^ord",
				IsAdmin:    true,
				SessionTTL: 180,
			}
			createResp, err := api.Create(ctx, createReq)
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp.Id, ShouldBeGreaterThan, 0)

			Convey("Then the user has been created", func() {
				user, err := api.Get(ctx, &pb.UserRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(user.Username, ShouldResemble, createReq.Username)
				So(user.SessionTTL, ShouldResemble, createReq.SessionTTL)
				So(user.IsAdmin, ShouldResemble, createReq.IsAdmin)

				Convey("Then get all users returns a single item", func() {
					users, err := api.List(ctx, &pb.ListUserRequest{
						Limit:  10,
						Offset: 0,
					})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
					So(users.Result, ShouldHaveLength, 1)
					So(users.TotalCount, ShouldEqual, 1)
					So(users.Result[0].Username, ShouldResemble, user.Username)
				})

				Convey("Then login in succeeds", func() {
					jwt, err := apiInternal.Login(ctx, &pb.LoginRequest{
						Username: createReq.Username,
						Password: createReq.Password,
					})
					So(err, ShouldBeNil)
					So(jwt, ShouldNotBeNil)
				})

				Convey("When updating the user", func() {
					updateUser := &pb.UpdateUserRequest{
						Id:         createResp.Id,
						Username:   "anotheruser",
						SessionTTL: 300,
						IsAdmin:    false,
					}
					_, err := api.Update(ctx, updateUser)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					Convey("Then the user has been updated", func() {
						userUpd, err := api.Get(ctx, &pb.UserRequest{
							Id: createResp.Id,
						})
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(userUpd.Username, ShouldResemble, updateUser.Username)
						So(userUpd.SessionTTL, ShouldResemble, updateUser.SessionTTL)
						So(userUpd.IsAdmin, ShouldResemble, updateUser.IsAdmin)
					})

					Convey("When updating the user's password", func() {
						updatePass := &pb.UpdateUserPasswordRequest{
							Id:       createResp.Id,
							Password: "newpasstest",
						}
						_, err := api.UpdatePassword(ctx, updatePass)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)

						Convey("Then the user can log in with the new password", func() {
							jwt, err := apiInternal.Login(ctx, &pb.LoginRequest{
								Username: updateUser.Username,
								Password: updatePass.Password,
							})
							So(err, ShouldBeNil)
							So(jwt, ShouldNotBeNil)
						})
					})
				})

				Convey("When deleting the user", func() {
					_, err := api.Delete(ctx, &pb.UserRequest{
						Id: createResp.Id,
					})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					Convey("Then the user has been deleted", func() {
						users, err := api.List(ctx, &pb.ListUserRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)
						So(users.Result, ShouldHaveLength, 0)
						So(users.TotalCount, ShouldEqual, 0)
					})
				})
			})
		})
	})
}
