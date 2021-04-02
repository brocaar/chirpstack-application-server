package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/test"
)

type testIntegrationSettings struct {
	URL string
	Key int64
}

func TestIntegration(t *testing.T) {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		t.Fatal(err)
	}

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	Convey("Given a clean database with an organization, network-server, service-profile, and application", t, func() {
		So(MigrateDown(DB().DB), ShouldBeNil)
		So(MigrateUp(DB().DB), ShouldBeNil)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(context.Background(), db, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(context.Background(), DB(), &n), ShouldBeNil)

		sp := ServiceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
		}
		So(CreateServiceProfile(context.Background(), DB(), &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		app := Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: spID,
		}
		So(CreateApplication(context.Background(), db, &app), ShouldBeNil)

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
			So(CreateIntegration(context.Background(), db, &intgr), ShouldBeNil)
			intgr.CreatedAt = intgr.CreatedAt.UTC().Truncate(time.Millisecond)
			intgr.UpdatedAt = intgr.UpdatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then it can be retrieved", func() {
				i, err := GetIntegration(context.Background(), db, intgr.ID)
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
				i, err := GetIntegrationByApplicationID(context.Background(), db, app.ID, "REST")
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
				ints, err := GetIntegrationsForApplicationID(context.Background(), db, app.ID)
				So(err, ShouldBeNil)

				So(ints, ShouldHaveLength, 1)
				So(ints[0].ID, ShouldEqual, intgr.ID)
			})

			Convey("Then it can be updated", func() {
				settings.URL = "http://foo.bar/updated"
				intgr.Settings, err = json.Marshal(settings)
				So(err, ShouldBeNil)
				So(UpdateIntegration(context.Background(), db, &intgr), ShouldBeNil)

				i, err := GetIntegration(context.Background(), db, intgr.ID)
				So(err, ShouldBeNil)

				var s testIntegrationSettings
				So(json.Unmarshal(i.Settings, &s), ShouldBeNil)
				So(s, ShouldResemble, settings)
			})

			Convey("Then it can be deleted", func() {
				So(DeleteIntegration(context.Background(), db, intgr.ID), ShouldBeNil)
				_, err := GetIntegration(context.Background(), db, intgr.ID)
				So(err, ShouldResemble, ErrDoesNotExist)
			})
		})
	})
}
