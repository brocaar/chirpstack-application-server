package external

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	"github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownlinkQueueAPI(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database, an organization, application + node and api instance", t, func() {
		test.MustResetDB(storage.DB().DB)

		nsClient := mock.NewClient()
		nsClient.GetNextDownlinkFCntForDevEUIResponse = ns.GetNextDownlinkFCntForDevEUIResponse{
			FCnt: 12,
		}
		networkserver.SetPool(mock.NewPool(nsClient))

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDeviceQueueAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(storage.DB(), &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(storage.DB(), &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			Name:            "test-sp",
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
		}
		So(storage.CreateServiceProfile(storage.DB(), &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
		}
		So(storage.CreateDeviceProfile(storage.DB(), &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			ServiceProfileID: spID,
			Name:             "test-app",
		}
		So(storage.CreateApplication(storage.DB(), &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dpID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDevice(storage.DB(), &d), ShouldBeNil)

		da := storage.DeviceActivation{
			DevEUI:  d.DevEUI,
			DevAddr: lorawan.DevAddr{1, 2, 3, 4},
			AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDeviceActivation(storage.DB(), &da), ShouldBeNil)

		b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 12, []byte{1, 2, 3, 4})
		So(err, ShouldBeNil)

		Convey("Given a custom JS application codec", func() {
			app.PayloadCodec = codec.CustomJSType
			app.PayloadEncoderScript = `
				function Encode(fPort, obj) {
					return [
						obj.Bytes[3],
						obj.Bytes[2],
						obj.Bytes[1],
						obj.Bytes[0]
					];
				}
			`
			So(storage.UpdateApplication(storage.DB(), app), ShouldBeNil)

			Convey("When enqueueing a downlink queue item with raw JSON object", func() {
				resp, err := api.Enqueue(ctx, &pb.EnqueueDeviceQueueItemRequest{
					DeviceQueueItem: &pb.DeviceQueueItem{
						DevEui:     d.DevEUI.String(),
						FPort:      10,
						JsonObject: `{"Bytes": [4,3,2,1]}`,
					},
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request has been made to the network-server", func() {
					So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevEui:     d.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      10,
						},
					})
				})

				Convey("Then the expected response was returned", func() {
					So(resp, ShouldResemble, &pb.EnqueueDeviceQueueItemResponse{
						FCnt: 12,
					})
				})
			})
		})

		Convey("When enqueueing a downlink queue item", func() {
			resp, err := api.Enqueue(ctx, &pb.EnqueueDeviceQueueItemRequest{
				DeviceQueueItem: &pb.DeviceQueueItem{
					DevEui: d.DevEUI.String(),
					FPort:  10,
					Data:   []byte{1, 2, 3, 4},
				},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the expected request has been made to the network-server", func() {
				So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
				So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, ns.CreateDeviceQueueItemRequest{
					Item: &ns.DeviceQueueItem{
						DevEui:     d.DevEUI[:],
						FrmPayload: b,
						FCnt:       12,
						FPort:      10,
					},
				})
			})

			Convey("Then the expected response was returned", func() {
				So(resp, ShouldResemble, &pb.EnqueueDeviceQueueItemResponse{
					FCnt: 12,
				})
			})
		})

		Convey("Given a mocked device-queue item", func() {
			nsClient.GetDeviceQueueItemsForDevEUIResponse = ns.GetDeviceQueueItemsForDevEUIResponse{
				Items: []*ns.DeviceQueueItem{
					{
						DevEui:     d.DevEUI[:],
						FrmPayload: b,
						FCnt:       12,
						FPort:      10,
						Confirmed:  true,
					},
				},
			}

			Convey("Then list returns the expected item", func() {
				resp, err := api.List(ctx, &pb.ListDeviceQueueItemsRequest{
					DevEui: d.DevEUI.String(),
				})
				So(err, ShouldBeNil)
				So(resp.DeviceQueueItems, ShouldHaveLength, 1)
				So(resp.DeviceQueueItems[0], ShouldResemble, &pb.DeviceQueueItem{
					DevEui:    d.DevEUI.String(),
					Confirmed: true,
					FPort:     10,
					FCnt:      12,
					Data:      []byte{1, 2, 3, 4},
				})
			})
		})

		Convey("When calling Flush", func() {
			_, err := api.Flush(ctx, &pb.FlushDeviceQueueRequest{
				DevEui: d.DevEUI.String(),
			})
			So(err, ShouldBeNil)

			Convey("Then the expected request has been made to the network-server", func() {
				So(nsClient.FlushDeviceQueueForDevEUIChan, ShouldHaveLength, 1)
				So(<-nsClient.FlushDeviceQueueForDevEUIChan, ShouldResemble, ns.FlushDeviceQueueForDevEUIRequest{
					DevEui: d.DevEUI[:],
				})
			})
		})
	})
}
