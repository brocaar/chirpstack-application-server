package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestOrganizationAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		config.C.PostgreSQL.DB = db
		test.MustResetDB(config.C.PostgreSQL.DB)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewOrganizationAPI(validator)
		userAPI := NewUserAPI(validator)

		Convey("When creating an organization with a bad name (spaces)", func() {
			validator.returnIsAdmin = true
			createReq := &pb.CreateOrganizationRequest{
				Name:            "organization name",
				DisplayName:     "Display Name",
				CanHaveGateways: true,
			}
			createResp, err := api.Create(ctx, createReq)
			So(err, ShouldNotBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp, ShouldBeNil)
		})

		Convey("When creating an organization as a global admin with a valid name", func() {
			validator.returnIsAdmin = true
			createReq := &pb.CreateOrganizationRequest{
				Name:            "orgName",
				DisplayName:     "Display Name",
				CanHaveGateways: true,
			}
			createResp, err := api.Create(ctx, createReq)
			So(err, ShouldBeNil)
			So(validator.validatorFuncs, ShouldHaveLength, 1)
			So(createResp, ShouldNotBeNil)

			Convey("Then the organization has been created", func() {
				org, err := api.Get(ctx, &pb.OrganizationRequest{
					Id: createResp.Id,
				})
				So(err, ShouldBeNil)
				So(org.Name, ShouldEqual, createReq.Name)
				So(org.DisplayName, ShouldEqual, createReq.DisplayName)
				So(org.CanHaveGateways, ShouldEqual, createReq.CanHaveGateways)

				orgs, err := api.List(ctx, &pb.ListOrganizationRequest{
					Limit:  10,
					Offset: 0,
				})
				So(err, ShouldBeNil)
				So(validator.validatorFuncs, ShouldHaveLength, 1)
				So(orgs, ShouldNotBeNil)
				// Default org is already in the database.
				So(orgs.Result, ShouldHaveLength, 2)
				orgId := int64(0)

				// If the length is not what was expected, then these checks
				// will, at best, be checking against some random data, and at
				// worst, crash the test.
				if 2 == len(orgs.Result) {
					So(orgs.Result[0].Name, ShouldEqual, createReq.Name)
					So(orgs.Result[0].DisplayName, ShouldEqual, createReq.DisplayName)
					So(orgs.Result[0].CanHaveGateways, ShouldEqual, createReq.CanHaveGateways)
					orgId = createResp.Id
				}

				Convey("When updating the organization", func() {
					updateOrg := &pb.UpdateOrganizationRequest{
						Id:              orgId,
						Name:            "anotherorg",
						DisplayName:     "Display Name 2",
						CanHaveGateways: false,
					}
					_, err := api.Update(ctx, updateOrg)
					So(err, ShouldBeNil)
					So(validator.validatorFuncs, ShouldHaveLength, 1)

					Convey("Then the organization has been updated", func() {
						orgUpd, err := api.Get(ctx, &pb.OrganizationRequest{
							Id: orgId,
						})
						So(err, ShouldBeNil)
						So(validator.validatorFuncs, ShouldHaveLength, 1)
						So(orgUpd.Name, ShouldResemble, updateOrg.Name)
						So(orgUpd.DisplayName, ShouldResemble, updateOrg.DisplayName)
						So(orgUpd.CanHaveGateways, ShouldResemble, updateOrg.CanHaveGateways)
					})

				})

				// Add a new user for adding to the organization.
				Convey("When adding a user", func() {
					userReq := &pb.AddUserRequest{
						Username:   "username",
						Password:   "pass^^ord",
						IsActive:   true,
						SessionTTL: 180,
						Email:      "foo@bar.com",
					}
					userResp, err := userAPI.Create(ctx, userReq)
					So(err, ShouldBeNil)

					validator.returnIsAdmin = false
					validator.returnUsername = userReq.Username

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
						addOrgUser := &pb.OrganizationUserRequest{
							Id:      orgId,
							UserID:  userResp.Id,
							IsAdmin: false,
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
								Id:     orgId,
								Limit:  10,
								Offset: 0,
							})
							So(err, ShouldBeNil)
							So(orgUsers.Result, ShouldHaveLength, 1)
							So(orgUsers.Result[0].Id, ShouldEqual, userResp.Id)
							So(orgUsers.Result[0].Username, ShouldEqual, userReq.Username)
							So(orgUsers.Result[0].IsAdmin, ShouldEqual, addOrgUser.IsAdmin)
						})

						Convey("When updating the user in the organization", func() {
							updOrgUser := &pb.OrganizationUserRequest{
								Id:      addOrgUser.Id,
								UserID:  addOrgUser.UserID,
								IsAdmin: !addOrgUser.IsAdmin,
							}
							_, err := api.UpdateUser(ctx, updOrgUser)
							So(err, ShouldBeNil)

							Convey("Then the user should be changed", func() {
								orgUsers, err := api.ListUsers(ctx, &pb.ListOrganizationUsersRequest{
									Id:     orgId,
									Limit:  10,
									Offset: 0,
								})
								So(err, ShouldBeNil)
								So(orgUsers, ShouldNotBeNil)
								if nil != orgUsers {
									So(orgUsers.Result, ShouldHaveLength, 1)
									if 1 == len(orgUsers.Result) {
										So(orgUsers.Result[0].Id, ShouldEqual, userResp.Id)
										So(orgUsers.Result[0].Username, ShouldEqual, userReq.Username)
										So(orgUsers.Result[0].IsAdmin, ShouldEqual, updOrgUser.IsAdmin)
									}
								}
							})

						})

						Convey("When removing the user from the organization", func() {
							delOrgUser := &pb.DeleteOrganizationUserRequest{
								Id:     addOrgUser.Id,
								UserID: addOrgUser.UserID,
							}
							_, err := api.DeleteUser(ctx, delOrgUser)
							So(err, ShouldBeNil)

							Convey("Then the user should be removed", func() {
								orgUsers, err := api.ListUsers(ctx, &pb.ListOrganizationUsersRequest{
									Id:     orgId,
									Limit:  10,
									Offset: 0,
								})
								So(err, ShouldBeNil)
								So(orgUsers, ShouldNotBeNil)
								So(orgUsers.Result, ShouldHaveLength, 0)
							})
						})
					})

					Convey("When deleting the organization", func() {
						validator.returnIsAdmin = true

						_, err := api.Delete(ctx, &pb.OrganizationRequest{
							Id: orgId,
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
