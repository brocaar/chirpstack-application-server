package storage

import (
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestApplication(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with an organization, network-server and service-profile", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		sp := ServiceProfile{
			Name:            "test-service-profile",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

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
				OrganizationID:       org.ID,
				ServiceProfileID:     sp.ServiceProfile.ServiceProfileID,
				Name:                 "test-application",
				Description:          "A test application",
				PayloadCodec:         "CUSTOM_JS",
				PayloadEncoderScript: "Encode() {}",
				PayloadDecoderScript: "Decode() {}",
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
				So(apps[0].ID, ShouldEqual, app.ID)
				So(apps[0].ServiceProfileName, ShouldEqual, sp.Name)

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
				So(apps[0].ID, ShouldEqual, app.ID)
				So(apps[0].ServiceProfileName, ShouldEqual, sp.Name)
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
