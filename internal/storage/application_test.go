package storage

import (
	"testing"

	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestApplication(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		Convey("When creating an application with an invalid name", func() {
			app := Application{
				Name: "i contain spaces",
			}
			err := CreateApplication(db, &app)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Cause(err), ShouldResemble, ErrApplicationInvalidName)
			})
		})

		Convey("When creating an application", func() {
			app := Application{
				Name:               "test-application",
				Description:        "A test application",
				RXDelay:            2,
				RX1DROffset:        3,
				RXWindow:           RX2,
				RX2DR:              3,
				ADRInterval:        20,
				InstallationMargin: 5,
				IsABP:              true,
				IsClassC:           true,
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

			Convey("When creating a user", func() {
				user := User{
					Username:   "username",
					IsAdmin:    false,
					IsActive:   true,
					SessionTTL: 20,
				}
				password := "somepassword"
				userId, err := CreateUser(db, &user, password)
				So(err, ShouldBeNil)

				Convey("Then the application count for the user is 0", func() {
					count, err := GetApplicationCountForUser(db, user.Username)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)

					apps, err := GetApplicationsForUser(db, user.Username, 10, 0)
					So(err, ShouldBeNil)
					So(apps, ShouldHaveLength, 0)
				})

				Convey("Then the user can be added to the application", func() {
					err := CreateUserForApplication(db, app.ID, userId, true)
					So(err, ShouldBeNil)
					Convey("Then the user count for the application is 1", func() {
						count, err := GetApplicationUsersCount(db, app.ID)
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)

						Convey("Then the application count for the user is 1", func() {
							count, err := GetApplicationCountForUser(db, user.Username)
							So(err, ShouldBeNil)
							So(count, ShouldEqual, 1)

							apps, err := GetApplicationsForUser(db, user.Username, 10, 0)
							So(err, ShouldBeNil)
							So(apps, ShouldHaveLength, 1)
						})
					})
					Convey("Then the user can be accessed via application get", func() {
						ua, err := GetUserForApplication(db, app.ID, userId)
						So(err, ShouldBeNil)
						So(ua.UserID, ShouldEqual, userId)
						So(ua.Username, ShouldEqual, user.Username)
						So(ua.IsAdmin, ShouldEqual, true)
						So(ua.CreatedAt, ShouldResemble, ua.UpdatedAt)
					})
					Convey("Then the user can be accessed via get all users for application", func() {
						uas, err := GetApplicationUsers(db, app.ID, 10, 0)
						So(err, ShouldBeNil)
						So(uas, ShouldNotBeNil)
						So(uas, ShouldHaveLength, 1)
						So(uas[0].UserID, ShouldEqual, userId)
						So(uas[0].Username, ShouldEqual, user.Username)
						So(uas[0].IsAdmin, ShouldEqual, true)
						So(uas[0].CreatedAt, ShouldResemble, uas[0].UpdatedAt)
					})
					Convey("Then the user access to the application can be updated", func() {
						err := UpdateUserForApplication(db, app.ID, userId, false)
						So(err, ShouldBeNil)
						Convey("Then the user can be accessed showing the new setting", func() {
							ua, err := GetUserForApplication(db, app.ID, userId)
							So(err, ShouldBeNil)
							So(ua.UserID, ShouldEqual, userId)
							So(ua.IsAdmin, ShouldEqual, false)
							So(ua.CreatedAt, ShouldNotResemble, ua.UpdatedAt)
						})
					})
					Convey("Then the user can be deleted from the application", func() {
						err := DeleteUserForApplication(db, app.ID, userId)
						So(err, ShouldBeNil)
						Convey("Then the user cannot be accessed via get", func() {
							ua, err := GetUserForApplication(db, app.ID, userId)
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
