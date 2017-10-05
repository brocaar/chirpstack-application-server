package storage

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan/backend"
)

func TestDeviceQueue(t *testing.T) {
	conf := test.GetConfig()
	db, err := OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db
	nsClient := test.NewNetworkServerClient()
	common.NetworkServer = nsClient

	Convey("Given a clean database with an organization, network-server, service-profile, device-profile, application and device", t, func() {
		test.MustResetDB(db)

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
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		dp := DeviceProfile{
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			Name:            "test-dp",
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(CreateDeviceProfile(common.DB, &dp), ShouldBeNil)

		app := Application{
			OrganizationID:   org.ID,
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
			Name:             "test",
		}
		So(CreateApplication(db, &app), ShouldBeNil)

		d := Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(CreateDevice(db, &d), ShouldBeNil)

		Convey("Given a set of four queue items", func() {
			queue := []*DeviceQueueItem{
				{DevEUI: d.DevEUI, Reference: "a", Data: []byte{1, 2, 3, 4, 5, 6, 7}},
				{DevEUI: d.DevEUI, Reference: "b", Data: []byte{1, 2, 3, 4, 5, 6}},
				{DevEUI: d.DevEUI, Reference: "c", Data: []byte{1, 2, 3, 4, 5}},
				{DevEUI: d.DevEUI, Reference: "d", Data: []byte{1, 2, 3, 4}},
			}
			for i := range queue {
				So(CreateDeviceQueueItem(db, queue[i]), ShouldBeNil)
				queue[i].CreatedAt = queue[i].CreatedAt.UTC().Truncate(time.Millisecond)
				queue[i].UpdatedAt = queue[i].UpdatedAt.UTC().Truncate(time.Millisecond)
			}

			Convey("When getting calling GetNextDeviceQueueItem with maxPayloadSize=7", func() {
				qi, err := GetNextDeviceQueueItem(db, d.DevEUI, 7)
				qi.CreatedAt = qi.CreatedAt.UTC().Truncate(time.Millisecond)
				qi.UpdatedAt = qi.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(err, ShouldBeNil)
				Convey("Then the first item is returned", func() {
					So(qi, ShouldResemble, queue[0])
				})
			})

			Convey("When getting calling GetNextDeviceQueueItem with maxPayloadSize=5", func() {
				qi, err := GetNextDeviceQueueItem(db, d.DevEUI, 5)
				qi.CreatedAt = qi.CreatedAt.UTC().Truncate(time.Millisecond)
				qi.UpdatedAt = qi.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(err, ShouldBeNil)
				Convey("Then the third item is returned", func() {
					So(qi, ShouldResemble, queue[2])
				})
			})

			Convey("When getting calling GetNextDeviceQueueItem with maxPayloadSize=3", func() {
				qi, err := GetNextDeviceQueueItem(db, d.DevEUI, 3)
				So(err, ShouldBeNil)
				Convey("Then no item is returned", func() {
					So(qi, ShouldBeNil)
				})
			})
		})

		Convey("When creating a downlink queue item", func() {
			qi := DeviceQueueItem{
				DevEUI:    d.DevEUI,
				Reference: "abcd1234-1",
				Confirmed: true,
				FPort:     10,
				Data:      []byte{1, 2, 3, 4},
			}
			So(CreateDeviceQueueItem(db, &qi), ShouldBeNil)
			qi.CreatedAt = qi.CreatedAt.UTC().Truncate(time.Millisecond)
			qi.UpdatedAt = qi.UpdatedAt.UTC().Truncate(time.Millisecond)

			Convey("Then GetDeviceQueueItemCount returns 1", func() {
				size, err := GetDeviceQueueItemCount(db, d.DevEUI)
				So(err, ShouldBeNil)
				So(size, ShouldEqual, 1)
			})

			Convey("Then the device-queue item can be retrieved", func() {
				qi2, err := GetDeviceQueueItem(db, qi.ID)
				qi2.CreatedAt = qi2.CreatedAt.UTC().Truncate(time.Millisecond)
				qi2.UpdatedAt = qi2.UpdatedAt.UTC().Truncate(time.Millisecond)
				So(err, ShouldBeNil)
				So(qi2, ShouldResemble, qi)
			})

			Convey("Then getting a pending device-queue item returns an error", func() {
				_, err := GetPendingDeviceQueueItem(db, d.DevEUI)
				So(err, ShouldNotBeNil)
			})

			Convey("When updating a device-queue item to pending", func() {
				qi.Pending = true
				So(UpdateDeviceQueueItem(db, &qi), ShouldBeNil)
				qi.UpdatedAt = qi.UpdatedAt.UTC().Truncate(time.Millisecond)

				Convey("Then the device-queue item has been updated", func() {
					qi2, err := GetDeviceQueueItem(db, qi.ID)
					So(err, ShouldBeNil)
					qi2.CreatedAt = qi2.CreatedAt.UTC().Truncate(time.Millisecond)
					qi2.UpdatedAt = qi2.UpdatedAt.UTC().Truncate(time.Millisecond)
					So(qi2, ShouldResemble, qi)
				})

				Convey("Then getting a pending device-queue item returns this item", func() {
					qi2, err := GetPendingDeviceQueueItem(db, d.DevEUI)
					So(err, ShouldBeNil)
					qi2.CreatedAt = qi2.CreatedAt.UTC().Truncate(time.Millisecond)
					qi2.UpdatedAt = qi2.UpdatedAt.UTC().Truncate(time.Millisecond)
					So(qi2, ShouldResemble, qi)
				})
			})

			Convey("When listing the queue items", func() {
				Convey("Then it returns one item for the existing DevEUI", func() {
					items, err := GetDeviceQueueItems(db, d.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					items[0].CreatedAt = items[0].CreatedAt.UTC().Truncate(time.Millisecond)
					items[0].UpdatedAt = items[0].UpdatedAt.UTC().Truncate(time.Millisecond)
					So(items[0], ShouldResemble, qi)
				})

				Convey("Then it returns zero items for a non-existing DevEUI", func() {
					items, err := GetDeviceQueueItems(db, [8]byte{8, 7, 6, 5, 4, 3, 2, 1})
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("When deleting the device-queue item", func() {
				So(DeleteDeviceQueueItem(db, qi.ID), ShouldBeNil)

				Convey("Then the item has been deleted", func() {
					_, err := GetDeviceQueueItem(db, qi.ID)
					So(err, ShouldNotBeNil)
				})
			})

			Convey("When deleting all queue items for a device", func() {
				So(DeleteDeviceQueueItemsForDevEUI(db, d.DevEUI), ShouldBeNil)

				Convey("Then the queue is empty", func() {
					items, err := GetDeviceQueueItems(db, d.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})
		})

	})
}
