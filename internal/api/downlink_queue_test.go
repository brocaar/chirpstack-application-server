package api

import (
	"context"
	"testing"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownlinkQueueAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	common.DB = db

	Convey("Given a clean database, an organization, application + node and api instance", t, func() {
		test.MustResetDB(common.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{},
		}

		common.NetworkServerPool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDownlinkQueueAPI(validator)

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
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(common.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(storage.CreateApplication(common.DB, &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDevice(common.DB, &d), ShouldBeNil)

		da := storage.DeviceActivation{
			DevEUI:  d.DevEUI,
			DevAddr: lorawan.DevAddr{1, 2, 3, 4},
			AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			NwkSKey: lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		So(storage.CreateDeviceActivation(common.DB, &da), ShouldBeNil)

		Convey("Given the node is a class-c device", func() {
			nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
				DeviceProfile: &ns.DeviceProfile{
					SupportsClassC: true,
				},
			}

			nsClient.GetDeviceActivationResponse = ns.GetDeviceActivationResponse{
				FCntDown: 12,
			}

			Convey("When enqueueing a unconfirmed queue item", func() {
				_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
					DevEUI:    d.DevEUI.String(),
					Confirmed: false,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the payload was directly sent to the network-server", func() {
					b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
					So(err, ShouldBeNil)

					So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 1)
					So(<-nsClient.SendDownlinkDataChan, ShouldResemble, ns.SendDownlinkDataRequest{
						DevEUI:    d.DevEUI[:],
						Data:      b,
						Confirmed: false,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was not added to the queue", func() {
					items, err := storage.GetDeviceQueueItems(common.DB, d.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 0)
				})
			})

			Convey("When enqueueing a confirmed queue item", func() {
				_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
					DevEUI:    d.DevEUI.String(),
					Confirmed: true,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the payload was directly sent to the network-server", func() {
					b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
					So(err, ShouldBeNil)

					So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 1)
					So(<-nsClient.SendDownlinkDataChan, ShouldResemble, ns.SendDownlinkDataRequest{
						DevEUI:    d.DevEUI[:],
						Data:      b,
						Confirmed: true,
						FPort:     10,
						FCnt:      12,
					})
				})

				Convey("Then the item was added as pending item to the queue", func() {
					items, err := storage.GetDeviceQueueItems(common.DB, d.DevEUI)
					So(err, ShouldBeNil)
					So(items, ShouldHaveLength, 1)
					So(items[0].Pending, ShouldBeTrue)
				})
			})
		})

		Convey("When enqueueing a downlink queue item", func() {
			_, err := api.Enqueue(ctx, &pb.EnqueueDownlinkQueueItemRequest{
				DevEUI:    d.DevEUI.String(),
				Confirmed: true,
				FPort:     10,
				Data:      []byte{1, 2, 3, 4},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the queue contains a single item", func() {
				resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
					DevEUI: d.DevEUI.String(),
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				So(resp.Items, ShouldHaveLength, 1)
				So(resp.Items[0], ShouldResemble, &pb.DownlinkQueueItem{
					Id:        1,
					DevEUI:    d.DevEUI.String(),
					Confirmed: true,
					Pending:   false,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				})
			})

			Convey("Then nothing was sent to the network-server", func() {
				So(nsClient.SendDownlinkDataChan, ShouldHaveLength, 0)
			})

			Convey("When removing the queue item", func() {
				_, err := api.Delete(ctx, &pb.DeleteDownlinkQeueueItemRequest{
					DevEUI: d.DevEUI.String(),
					Id:     1,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the downlink queue item has been deleted", func() {
					resp, err := api.List(ctx, &pb.ListDownlinkQueueItemsRequest{
						DevEUI: d.DevEUI.String(),
					})
					So(err, ShouldBeNil)
					So(resp.Items, ShouldHaveLength, 0)
				})
			})
		})
	})
}
