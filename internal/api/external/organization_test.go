package external

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestOrganizationAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instance", t, func() {
		if err := storage.Setup(conf); err != nil {
			t.Fatal(err)
		}
		test.MustResetDB(storage.DB().DB)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewOrganizationAPI(validator)
		userAPI := NewUserAPI(validator)

		Convey("When creating an organization with a bad name (spaces)", func() {
			validator.returnIsAdmin = true
			createReq := &pb.CreateOrganizationRequest{
				Organization: &pb.Organization{
					Name:            "organization name",
					DisplayName:     "Display Name",
					CanHaveGateways: true,
				},
			}
			createResp, err := api.Create(ctx, createReq)
			So(err, ShouldNotBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp, ShouldBeNil)
		})

		Convey("When creating an organization as a global admin with a valid name", func() {
			validator.returnIsAdmin = true
			createReq := pb.CreateOrganizationRequest{
				Organization: &pb.Organization{
					Name:            "orgName",
					DisplayName:     "Display Name",
					CanHaveGateways: true,
				},
			}
			createResp, err := api.Create(ctx, &createReq)
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp, ShouldNotBeNil)

			Convey("Then the organization has been created", func() {
				org, err := api.Get(ctx, &pb.GetOrganizationRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)

				createReq.Organization.Id = createResp.Id
				So(org.Organization, ShouldResemble, createReq.Organization)

				orgs, err := api.List(ctx, &pb.ListOrganizationRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(orgs, ShouldNotBeNil)
				// Default org is already in the database.
				So(orgs.Result, ShouldHaveLength, 2)

				So(orgs.Result[0].Name, ShouldEqual, createReq.Organization.Name)
				So(orgs.Result[0].DisplayName, ShouldEqual, createReq.Organization.DisplayName)
				So(orgs.Result[0].CanHaveGateways, ShouldEqual, createReq.Organization.CanHaveGateways)

				Convey("When updating the organization", func() {
					updateOrg := &pb.UpdateOrganizationRequest{
						Organization: &pb.Organization{
							Id:              createResp.Id,
							Name:            "anotherorg",
							DisplayName:     "Display Name 2",
							CanHaveGateways: false,
						},
					}
					_, err := api.Update(ctx, updateOrg)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					Convey("Then the organization has been updated", func() {
						orgUpd, err := api.Get(ctx, &pb.GetOrganizationRequest{
							Id: createResp.Id,
						})
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)

						createReq.Organization.Id = createResp.Id
						So(orgUpd.Organization, ShouldResemble, updateOrg.Organization)
					})

				})

				// Add a new user for adding to the organization.
				Convey("When adding a user", func() {
					userReq := &pb.CreateUserRequest{
						User: &pb.User{
							Username:   "username",
							IsActive:   true,
							SessionTtl: 180,
							Email:      "foo@bar.com",
						},
						Password: "pass^^ord",
					}
					userResp, err := userAPI.Create(ctx, userReq)
					So(err, ShouldBeNil)

					validator.returnIsAdmin = false
					validator.returnUsername = userReq.User.Username

					Convey("When listing the organizations for the user", func() {
						orgs, err := api.List(ctx, &pb.ListOrganizationRequest{
							Limit:  10,
							Offset: 0,
						})
						So(err, ShouldBeNil)

						Convey("Then the user should not see any organizations", func() {
							So(orgs.TotalCount, ShouldEqual, 0)
							So(orgs.Result, ShouldHaveLength, 0)
						})
					})

					Convey("When adding the user to the organization", func() {
						addOrgUser := &pb.AddOrganizationUserRequest{
							OrganizationUser: &pb.OrganizationUser{
								OrganizationId: createResp.Id,
								UserId:         userResp.Id,
								IsAdmin:        false,
							},
						}
						_, err := api.AddUser(ctx, addOrgUser)
						So(err, ShouldBeNil)

						Convey("When listing the organizations for the user", func() {
							orgs, err := api.List(ctx, &pb.ListOrganizationRequest{
								Limit:  10,
								Offset: 0,
							})
							So(err, ShouldBeNil)

							Convey("Then the user should see the organization", func() {
								So(orgs.TotalCount, ShouldEqual, 1)
								So(orgs.Result, ShouldHaveLength, 1)
							})
						})

						Convey("Then the user should be part of the organization", func() {
							orgUsers, err := api.ListUsers(ctx, &pb.ListOrganizationUsersRequest{
								OrganizationId: createResp.Id,
								Limit:          10,
								Offset:         0,
							})
							So(err, ShouldBeNil)
							So(orgUsers.Result, ShouldHaveLength, 1)
							So(orgUsers.Result[0].UserId, ShouldEqual, userResp.Id)
							So(orgUsers.Result[0].Username, ShouldEqual, userReq.User.Username)
							So(orgUsers.Result[0].IsAdmin, ShouldEqual, addOrgUser.OrganizationUser.IsAdmin)
						})

						Convey("When updating the user in the organization", func() {
							updOrgUser := &pb.UpdateOrganizationUserRequest{
								OrganizationUser: &pb.OrganizationUser{
									OrganizationId: createResp.Id,
									UserId:         addOrgUser.OrganizationUser.UserId,
									IsAdmin:        !addOrgUser.OrganizationUser.IsAdmin,
								},
							}
							_, err := api.UpdateUser(ctx, updOrgUser)
							So(err, ShouldBeNil)

							Convey("Then the user should be changed", func() {
								orgUsers, err := api.ListUsers(ctx, &pb.ListOrganizationUsersRequest{
									OrganizationId: createResp.Id,
									Limit:          10,
									Offset:         0,
								})
								So(err, ShouldBeNil)
								So(orgUsers, ShouldNotBeNil)
								So(orgUsers.Result, ShouldHaveLength, 1)
								So(orgUsers.Result[0].UserId, ShouldEqual, userResp.Id)
								So(orgUsers.Result[0].Username, ShouldEqual, userReq.User.Username)
								So(orgUsers.Result[0].IsAdmin, ShouldEqual, updOrgUser.OrganizationUser.IsAdmin)
							})

						})

						Convey("When removing the user from the organization", func() {
							delOrgUser := &pb.DeleteOrganizationUserRequest{
								OrganizationId: createResp.Id,
								UserId:         addOrgUser.OrganizationUser.UserId,
							}
							_, err := api.DeleteUser(ctx, delOrgUser)
							So(err, ShouldBeNil)

							Convey("Then the user should be removed", func() {
								orgUsers, err := api.ListUsers(ctx, &pb.ListOrganizationUsersRequest{
									OrganizationId: createResp.Id,
									Limit:          10,
									Offset:         0,
								})
								So(err, ShouldBeNil)
								So(orgUsers, ShouldNotBeNil)
								So(orgUsers.Result, ShouldHaveLength, 0)
							})
						})
					})

					Convey("When deleting the organization", func() {
						validator.returnIsAdmin = true

						_, err := api.Delete(ctx, &pb.DeleteOrganizationRequest{
							Id: createResp.Id,
						})
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)

						Convey("Then the organization has been deleted", func() {
							orgs, err := api.List(ctx, &pb.ListOrganizationRequest{
								Limit:  10,
								Offset: 0,
							})
							So(err, ShouldBeNil)
							So(orgs.Result, ShouldHaveLength, 1)
							So(orgs.TotalCount, ShouldEqual, 1)
						})
					})
				})
			})
		})
	})
}
