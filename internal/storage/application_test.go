package storage

import (
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
)

func TestApplication(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db
	nsClient := test.NewNetworkServerClient()
	common.NetworkServer = nsClient

	Convey("Given a clean database with an organization, network-server and service-profile", t, func() {
		test.MustResetDB(common.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(common.DB, &n), ShouldBeNil)

		sp := ServiceProfile{
			Name:            "test-service-profile",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		Convey("When creating an application with an invalid name", func() {
			app := Application{
				OrganizationID:   org.ID,
				ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
				Name:             "i contain spaces",
			}
			err := CreateApplication(db, &app)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrApplicationInvalidName)
			})
		})

		Convey("When creating an application", func() {
			app := Application{
				OrganizationID:   org.ID,
				ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
				Name:             "test-application",
				Description:      "A test application",
			}
			So(CreateApplication(db, &app), ShouldBeNil)

			Convey("It can be get by id", func() {
				app2, err := GetApplication(db, app.ID)
				So(err, ShouldBeNil)
				So(app2, ShouldResemble, app)
			})

			Convey("Then get applications returns a single application", func() {
				apps, err := GetApplications(db, 10, 0)
				So(err, ShouldBeNil)
				So(apps, ShouldHaveLength, 1)
				So(apps[0], ShouldResemble, app)
			})

			Convey("Then get application count returns 1", func() {
				count, err := GetApplicationCount(db)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then the application count for the organization returns 1", func() {
				count, err := GetApplicationCountForOrganizationID(db, org.ID)
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("Then listing the applications for the organization returns the expected application", func() {
				apps, err := GetApplicationsForOrganizationID(db, org.ID, 10, 0)
				So(err, ShouldBeNil)
				So(apps, ShouldHaveLength, 1)
				So(apps[0], ShouldResemble, app)
			})

			Convey("When creating a user", func() {
				user := User{
					Username:   "username",
					IsAdmin:    false,
					IsActive:   true,
					SessionTTL: 20,
				}
				password := "somepassword"
				userID, err := CreateUser(db, &user, password)
				So(err, ShouldBeNil)

				Convey("Then the application count for the user is 0", func() {
					count, err := GetApplicationCountForUser(db, user.Username, org.ID)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)

					apps, err := GetApplicationsForUser(db, user.Username, org.ID, 10, 0)
					So(err, ShouldBeNil)
					So(apps, ShouldHaveLength, 0)
				})

				Convey("When adding the user to the organization", func() {
					err := CreateOrganizationUser(db, org.ID, user.ID, true)
					So(err, ShouldBeNil)

					Convey("Then the application count for the user is 1", func() {
						count, err := GetApplicationCountForUser(db, user.Username, org.ID)
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)

						apps, err := GetApplicationsForUser(db, user.Username, org.ID, 10, 0)
						So(err, ShouldBeNil)
						So(apps, ShouldHaveLength, 1)
					})
				})

				Convey("When adding the user to the application", func() {
					err := CreateUserForApplication(db, app.ID, userID, true)
					So(err, ShouldBeNil)

					Convey("Then the user count for the application is 1", func() {
						count, err := GetApplicationUsersCount(db, app.ID)
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)

						Convey("Then the application count for the user is 1", func() {
							count, err := GetApplicationCountForUser(db, user.Username, org.ID)
							So(err, ShouldBeNil)
							So(count, ShouldEqual, 1)

							apps, err := GetApplicationsForUser(db, user.Username, org.ID, 10, 0)
							So(err, ShouldBeNil)
							So(apps, ShouldHaveLength, 1)
						})
					})

					Convey("Then the user can be accessed via application get", func() {
						ua, err := GetUserForApplication(db, app.ID, userID)
						So(err, ShouldBeNil)
						So(ua.UserID, ShouldEqual, userID)
						So(ua.Username, ShouldEqual, user.Username)
						So(ua.IsAdmin, ShouldEqual, true)
						So(ua.CreatedAt, ShouldResemble, ua.UpdatedAt)
					})
					Convey("Then the user can be accessed via get all users for application", func() {
						uas, err := GetApplicationUsers(db, app.ID, 10, 0)
						So(err, ShouldBeNil)
						So(uas, ShouldNotBeNil)
						So(uas, ShouldHaveLength, 1)
						So(uas[0].UserID, ShouldEqual, userID)
						So(uas[0].Username, ShouldEqual, user.Username)
						So(uas[0].IsAdmin, ShouldEqual, true)
						So(uas[0].CreatedAt, ShouldResemble, uas[0].UpdatedAt)
					})
					Convey("Then the user access to the application can be updated", func() {
						err := UpdateUserForApplication(db, app.ID, userID, false)
						So(err, ShouldBeNil)
						Convey("Then the user can be accessed showing the new setting", func() {
							ua, err := GetUserForApplication(db, app.ID, userID)
							So(err, ShouldBeNil)
							So(ua.UserID, ShouldEqual, userID)
							So(ua.IsAdmin, ShouldEqual, false)
							So(ua.CreatedAt, ShouldNotResemble, ua.UpdatedAt)
						})
					})
					Convey("Then the user can be deleted from the application", func() {
						err := DeleteUserForApplication(db, app.ID, userID)
						So(err, ShouldBeNil)
						Convey("Then the user cannot be accessed via get", func() {
							ua, err := GetUserForApplication(db, app.ID, userID)
							So(err, ShouldNotBeNil)
							So(ua, ShouldBeNil)
						})
					})
				})
			})

			Convey("When updating the application", func() {
				app.Description = "some new description"
				So(UpdateApplication(db, app), ShouldBeNil)

				Convey("Then the application has been updated", func() {
					app2, err := GetApplication(db, app.ID)
					So(err, ShouldBeNil)
					So(app2, ShouldResemble, app)
				})
			})

			Convey("When deleting the application", func() {
				So(DeleteApplication(db, app.ID), ShouldBeNil)

				Convey("Then the application count returns 0", func() {
					count, err := GetApplicationCount(db)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})
			})
		})
	})
}
