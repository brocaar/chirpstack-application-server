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
				Name:        "test-application",
				Description: "A test application",
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
