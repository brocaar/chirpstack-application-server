package api

import (
	"testing"

	"github.com/brocaar/loraserver/api/ns"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestUserAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	config.C.PostgreSQL.DB = db

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}

		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewUserAPI(validator)
		apiInternal := NewInternalUserAPI(validator)

		Convey("When creating an user assigned to an organization", func() {
			org := storage.Organization{
				Name: "test-org",
			}
			So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

			createReq := pb.AddUserRequest{
				Username: "testuser",
				Password: "testpasswd",
				Organizations: []*pb.AddUserOrganization{
					{OrganizationID: org.ID, IsAdmin: true},
				},
				Email: "foo@bar.com",
			}
			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.Id, ShouldBeGreaterThan, 0)

			users, err := storage.GetOrganizationUsers(config.C.PostgreSQL.DB, org.ID, 10, 0)
			So(err, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
			So(users[0].UserID, ShouldEqual, createResp.Id)
			So(users[0].IsAdmin, ShouldBeTrue)
		})

		Convey("When creating an user", func() {
			validator.returnIsAdmin = true
			createReq := &pb.AddUserRequest{
				Username:   "username",
				Password:   "pass^^ord",
				IsAdmin:    true,
				SessionTTL: 180,
				Email:      "foo@bar.com",
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
				So(user.Email, ShouldEqual, createReq.Email)
				So(user.Note, ShouldEqual, createReq.Note)

				Convey("Then get all users returns 2 items (admin user already there)", func() {
					users, err := api.List(ctx, &pb.ListUserRequest{
						Limit:  10,
						Offset: 0,
					})
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)
					So(users.Result, ShouldHaveLength, 2)
					So(users.TotalCount, ShouldEqual, 2)
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
						Email:      "bar@foo.com",
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
						So(users.Result, ShouldHaveLength, 1)
						So(users.TotalCount, ShouldEqual, 1)
					})
				})
			})
		})
	})
}
