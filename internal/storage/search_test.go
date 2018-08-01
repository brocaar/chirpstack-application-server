package storage

import (
	"testing"

	"github.com/gofrs/uuid"

	"github.com/brocaar/lorawan"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSearch(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with test-data", t, func() {
		test.MustResetDB(db)

		u := User{
			Username: "testuser",
			Email:    "test@example.com",
		}
		_, err := CreateUser(db, &u, "testpw")
		So(err, ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(db, &n), ShouldBeNil)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(db, &org), ShouldBeNil)

		sp := ServiceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
		}
		So(CreateServiceProfile(db, &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
		}
		So(CreateDeviceProfile(db, &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		a := Application{
			Name:             "test-app",
			OrganizationID:   org.ID,
			ServiceProfileID: spID,
		}
		So(CreateApplication(db, &a), ShouldBeNil)

		gw := Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gateway",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(CreateGateway(db, &gw), ShouldBeNil)

		d := Device{
			DevEUI:          lorawan.EUI64{2, 3, 4, 5, 6, 7, 8, 9},
			Name:            "test-device",
			ApplicationID:   a.ID,
			DeviceProfileID: dpID,
		}
		So(CreateDevice(db, &d), ShouldBeNil)

		Convey("When the user is not global admin, this does not return any results", func() {
			queries := []string{
				"test",
				"org",
				"app",
				"010203",
				"020304",
				"device",
			}

			for _, q := range queries {
				res, err := GlobalSearch(db, u.Username, false, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, 0)
			}
		})

		Convey("When the user is global admin, this returns results", func() {
			queries := map[string]int{
				"test":   4,
				"org":    1,
				"app":    1,
				"010203": 1,
				"020304": 2,
				"device": 1,
				"dev":    1,
				"gatew":  1,
			}

			for q, c := range queries {
				res, err := GlobalSearch(db, u.Username, true, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, c)
			}
		})

		Convey("When the user is part of the organization, this returns results", func() {
			So(CreateOrganizationUser(db, org.ID, u.ID, false), ShouldBeNil)

			queries := map[string]int{
				"test":   4,
				"org":    1,
				"app":    1,
				"010203": 1,
				"020304": 2,
				"device": 1,
				"dev":    1,
				"gatew":  1,
			}

			for q, c := range queries {
				res, err := GlobalSearch(db, u.Username, false, q, 10, 0)
				So(err, ShouldBeNil)
				So(res, ShouldHaveLength, c)
			}
		})
	})
}
