package external

import (
	"testing"

	"github.com/brocaar/loraserver/api/ns"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestUserAPI(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database and api instance", t, func() {
		test.MustResetDB(storage.DB().DB)

		nsClient := mock.NewClient()
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}
		networkserver.SetPool(mock.NewPool(nsClient))

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewUserAPI(validator)
		apiInternal := NewInternalUserAPI(validator)

		Convey("When creating an user assigned to an organization", func() {
			org := storage.Organization{
				Name: "test-org",
			}
			So(storage.CreateOrganization(storage.DB(), &org), ShouldBeNil)

			createReq := pb.CreateUserRequest{
				User: &pb.User{
					Username: "testuser",
					Email:    "foo@bar.com",
				},
				Password: "testpasswd",
				Organizations: []*pb.UserOrganization{
					{OrganizationId: org.ID, IsAdmin: true},
				},
			}
			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(createResp.Id, ShouldBeGreaterThan, 0)

			users, err := storage.GetOrganizationUsers(storage.DB(), org.ID, 10, 0)
			So(err, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
			So(users[0].UserID, ShouldEqual, createResp.Id)
			So(users[0].IsAdmin, ShouldBeTrue)
		})

		Convey("When creating an user", func() {
			validator.returnIsAdmin = true
			createReq := &pb.CreateUserRequest{
				User: &pb.User{
					Username:   "username",
					IsAdmin:    true,
					SessionTtl: 180,
					Email:      "foo@bar.com",
				},
				Password: "pass^^ord",
			}
			createResp, err := api.Create(ctx, createReq)
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp.Id, ShouldBeGreaterThan, 0)

			Convey("Then the user has been created", func() {
				user, err := api.Get(ctx, &pb.GetUserRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				createReq.User.Id = createResp.Id
				So(user.User, ShouldResemble, createReq.User)

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
						Username: createReq.User.Username,
						Password: createReq.Password,
					})
					So(err, ShouldBeNil)
					So(jwt, ShouldNotBeNil)
				})

				Convey("When updating the user", func() {
					updateReq := pb.UpdateUserRequest{
						User: &pb.User{
							Id:         createResp.Id,
							Username:   "anotheruser",
							SessionTtl: 300,
							IsAdmin:    false,
							Email:      "bar@foo.com",
						},
					}
					_, err := api.Update(ctx, &updateReq)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					Convey("Then the user has been updated", func() {
						userUpd, err := api.Get(ctx, &pb.GetUserRequest{
							Id: createResp.Id,
						})
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(userUpd.User, ShouldResemble, updateReq.User)
					})

					Convey("When updating the user's password", func() {
						updatePass := &pb.UpdateUserPasswordRequest{
							UserId:   createResp.Id,
							Password: "newpasstest",
						}
						_, err := api.UpdatePassword(ctx, updatePass)
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)

						Convey("Then the user can log in with the new password", func() {
							jwt, err := apiInternal.Login(ctx, &pb.LoginRequest{
								Username: updateReq.User.Username,
								Password: updatePass.Password,
							})
							So(err, ShouldBeNil)
							So(jwt, ShouldNotBeNil)
						})
					})
				})

				Convey("When deleting the user", func() {
					_, err := api.Delete(ctx, &pb.DeleteUserRequest{
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
