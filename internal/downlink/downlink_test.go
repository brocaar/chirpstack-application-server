package downlink

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestHandleDownlinkQueueItem(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	common.DB = db

	Convey("Given a clean database an organization, application + node", t, func() {
		test.MustResetDB(common.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetDeviceActivationResponse = ns.GetDeviceActivationResponse{
			FCntDown: 12,
		}
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}
		common.NetworkServerPool = test.NewNetworkServerPool(nsClient)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(common.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(common.DB, &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			Name:            "test-sp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(common.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(storage.CreateApplication(common.DB, &app), ShouldBeNil)

		device := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDevice(common.DB, &device), ShouldBeNil)

		da := storage.DeviceActivation{
			DevEUI:  [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		}
		So(storage.CreateDeviceActivation(common.DB, &da), ShouldBeNil)

		qi := storage.DeviceQueueItem{
			Reference: "test",
			DevEUI:    device.DevEUI,
			Confirmed: false,
			FPort:     10,
			Data:      []byte{1, 2, 3, 4},
		}

		b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
		So(err, ShouldBeNil)

		Convey("When calling HandleDownlinkQueueItem for a non class-c device", func() {
			So(HandleDownlinkQueueItem(device, &qi), ShouldBeNil)

			Convey("Then the item was added to the queue", func() {
				items, err := storage.GetDeviceQueueItems(common.DB, device.DevEUI)
				So(err, ShouldBeNil)
				So(items, ShouldHaveLength, 1)

				qi.CreatedAt = items[0].CreatedAt
				qi.UpdatedAt = items[0].UpdatedAt
				So(items[0], ShouldResemble, qi)
			})

			Convey("Then nothing was sent to the network-server", func() {
				So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 0)
			})
		})

		Convey("When calling HandleDownlinkQueueItem for a class-c device", func() {
			nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
				DeviceProfile: &ns.DeviceProfile{
					SupportsClassC: true,
				},
			}

			Convey("When the queue item is confirmed", func() {
				qi.Confirmed = true

				So(HandleDownlinkQueueItem(device, &qi), ShouldBeNil)

				Convey("Then the payload was sent to the network-server", func() {
					So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 1)
					So(<-nsClient.SendDownlinkDataChan, ShouldResemble, ns.SendDownlinkDataRequest{
						DevEUI:    device.DevEUI[:],
						Data:      b,
						Confirmed: true,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was added as pending to the queue", func() {
					items, err := storage.GetDeviceQueueItems(common.DB, device.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					So(items[0].Pending, ShouldBeTrue)
				})
			})

			Convey("When the queue item is unconfirmed", func() {
				So(HandleDownlinkQueueItem(device, &qi), ShouldBeNil)

				Convey("Then the payload was sent to the network-server", func() {
					So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 1)
					So(<-nsClient.SendDownlinkDataChan, ShouldResemble, ns.SendDownlinkDataRequest{
						DevEUI:    device.DevEUI[:],
						Data:      b,
						Confirmed: false,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the queue is empty", func() {
					items, err := storage.GetDeviceQueueItems(common.DB, device.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})
		})

	})
}
