package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

type testIntegrationSettings struct {
	URL string
	Key int64
}

func TestIntegration(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with an application", t, func() {
		db, err := OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		app := Application{
			Name:           "test-app",
			OrganizationID: org.ID,
		}
		So(CreateApplication(db, &app), ShouldBeNil)

		Convey("When creating an integration", func() {
			settings := testIntegrationSettings{
				URL: "http://foo.bar/",
				Key: 12345,
			}
			intgr := Integration{
				ApplicationID: app.ID,
				Kind:          "REST",
			}
			intgr.Settings, err = json.Marshal(settings)
			So(err, ShouldBeNil)
			So(CreateIntegration(db, &intgr), ShouldBeNil)
			intgr.CreatedAt = intgr.CreatedAt.UTC().Truncate(time.Millisecond)
			intgr.UpdatedAt = intgr.UpdatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then it can be retrieved", func() {
				i, err := GetIntegration(db, intgr.ID)
				So(err, ShouldBeNil)

				var s testIntegrationSettings
				So(json.Unmarshal(i.Settings, &s), ShouldBeNil)
				So(s, ShouldResemble, settings)

				i.CreatedAt = i.CreatedAt.UTC().Truncate(time.Millisecond)
				i.UpdatedAt = i.UpdatedAt.UTC().Truncate(time.Millisecond)

				// set to nil because the order of keys might be different
				intgr.Settings = nil
				i.Settings = nil
				So(i, ShouldResemble, intgr)
			})

			Convey("Then it can be retrieved by the application ID", func() {
				i, err := GetIntegrationByApplicationID(db, app.ID, "REST")
				So(err, ShouldBeNil)

				var s testIntegrationSettings
				So(json.Unmarshal(i.Settings, &s), ShouldBeNil)
				So(s, ShouldResemble, settings)

				i.CreatedAt = i.CreatedAt.UTC().Truncate(time.Millisecond)
				i.UpdatedAt = i.UpdatedAt.UTC().Truncate(time.Millisecond)

				// set to nil because the order of keys might be different
				intgr.Settings = nil
				i.Settings = nil
				So(i, ShouldResemble, intgr)
			})

			Convey("Then it can be retrieved by the application id", func() {
				ints, err := GetIntegrationsForApplicationID(db, app.ID)
				So(err, ShouldBeNil)

				So(ints, ShouldHaveLength, 1)
				So(ints[0].ID, ShouldEqual, intgr.ID)
			})

			Convey("Then it can be updated", func() {
				settings.URL = "http://foo.bar/updated"
				intgr.Settings, err = json.Marshal(settings)
				So(err, ShouldBeNil)
				So(UpdateIntegration(db, &intgr), ShouldBeNil)

				i, err := GetIntegration(db, intgr.ID)
				So(err, ShouldBeNil)

				var s testIntegrationSettings
				So(json.Unmarshal(i.Settings, &s), ShouldBeNil)
				So(s, ShouldResemble, settings)
			})

			Convey("Then it can be deleted", func() {
				So(DeleteIntegration(db, intgr.ID), ShouldBeNil)
				_, err := GetIntegration(db, intgr.ID)
				So(err, ShouldResemble, ErrDoesNotExist)
			})
		})
	})
}
