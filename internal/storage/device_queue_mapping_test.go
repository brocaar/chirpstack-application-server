package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test"
)

func TestDeviceQueueMapping(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	nsClient := test.NewNetworkServerClient()
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database and a device", t, func() {
		test.MustResetDB(db)

		org := Organization{
			Name: "test-org",
		}
		So(CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		sp := ServiceProfile{
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			Name:            "test-sp",
		}
		So(CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

		dp := DeviceProfile{
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			Name:            "test-dp",
		}
		So(CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		app := Application{
			OrganizationID:   org.ID,
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
			Name:             "test-app",
		}
		So(CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		d := Device{
			Name:            "test-device",
			DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
		}
		So(CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

		Convey("When creating a device-queue mapping", func() {
			dqm := DeviceQueueMapping{
				Reference: "test-123",
				DevEUI:    d.DevEUI,
				FCnt:      10,
			}
			So(CreateDeviceQueueMapping(config.C.PostgreSQL.DB, &dqm), ShouldBeNil)
			dqm.CreatedAt = dqm.CreatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then GetDeviceQueueMappingForDevEUIAndFCnt with the same FCnt returns it", func() {
				dqmGet, err := GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, d.DevEUI, 10)
				So(err, ShouldBeNil)
				dqmGet.CreatedAt = dqmGet.CreatedAt.UTC().Truncate(time.Millisecond)
				So(dqmGet, ShouldResemble, dqm)
			})

			Convey("Then DeleteDeviceQueueMapping deletes the mapping", func() {
				So(DeleteDeviceQueueMapping(config.C.PostgreSQL.DB, dqm.ID), ShouldBeNil)
				_, err := GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, d.DevEUI, 10)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

			Convey("Then FlushDeviceQueueMappingForDevEUI flushes all mappings", func() {
				So(FlushDeviceQueueMappingForDevEUI(config.C.PostgreSQL.DB, d.DevEUI), ShouldBeNil)
				_, err := GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, d.DevEUI, 10)
				So(err, ShouldEqual, ErrDoesNotExist)
			})

		})

		Convey("Testing GetDeviceQueueMappingForDevEUIAndFCnt with different frame-counters", func() {
			tests := []struct {
				Name     string
				FCnt     uint32
				Mappings []DeviceQueueMapping

				ExpectedReference string
				ExpectedError     error
			}{
				{
					Name: "frame-counter equal",
					FCnt: 10,
					Mappings: []DeviceQueueMapping{
						{DevEUI: d.DevEUI, FCnt: 10, Reference: "test-1"},
					},
					ExpectedReference: "test-1",
				},
				{
					Name: "two mappings, first one 'expired', second matching",
					FCnt: 10,
					Mappings: []DeviceQueueMapping{
						{DevEUI: d.DevEUI, FCnt: 9, Reference: "test-1"},
						{DevEUI: d.DevEUI, FCnt: 10, Reference: "test-2"},
					},
					ExpectedReference: "test-2",
				},
				{
					Name: "one mapping with FCnt + 1 returns does not exist",
					FCnt: 4294967295,
					Mappings: []DeviceQueueMapping{
						{DevEUI: d.DevEUI, FCnt: 0, Reference: "test-1"}, // note: 4294967295 + 1 = 0 (uint32)
					},
					ExpectedError: ErrDoesNotExist,
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					// create mappings
					for i := range test.Mappings {
						So(CreateDeviceQueueMapping(config.C.PostgreSQL.DB, &test.Mappings[i]), ShouldBeNil)
					}

					dqm, err := GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, d.DevEUI, test.FCnt)
					if test.ExpectedError != nil {
						So(err, ShouldResemble, test.ExpectedError)
						return
					}

					So(err, ShouldEqual, nil)
					So(dqm.Reference, ShouldResemble, test.ExpectedReference)
				})
			}
		})
	})
}
