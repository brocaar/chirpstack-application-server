package api

import (
	"context"
	"testing"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
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

	config.C.PostgreSQL.DB = db

	Convey("Given a clean database, an organization, application + node and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		nsClient.GetNextDownlinkFCntForDevEUIResponse = ns.GetNextDownlinkFCntForDevEUIResponse{
			FCnt: 12,
		}
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		ctx := context.Background()
		validator := &TestValidator{}
		api := NewDeviceQueueAPI(validator)

		org := storage.Organization{
			Name: "test-org",
		}
		So(storage.CreateOrganization(config.C.PostgreSQL.DB, &org), ShouldBeNil)

		n := storage.NetworkServer{
			Name:   "test-ns",
			Server: "test-ns:1234",
		}
		So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

		sp := storage.ServiceProfile{
			Name:            "test-sp",
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

		dp := storage.DeviceProfile{
			Name:            "test-dp",
			NetworkServerID: n.ID,
			OrganizationID:  org.ID,
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			Name:             "test-app",
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

		da := storage.DeviceActivation{
			DevEUI:  d.DevEUI,
			DevAddr: lorawan.DevAddr{1, 2, 3, 4},
			AppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			NwkSKey: lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		So(storage.CreateDeviceActivation(config.C.PostgreSQL.DB, &da), ShouldBeNil)

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
			So(storage.UpdateApplication(config.C.PostgreSQL.DB, app), ShouldBeNil)

			Convey("When enqueueing a downlink queue item with raw JSON object", func() {
				_, err := api.Enqueue(ctx, &pb.EnqueueDeviceQueueItemRequest{
					DevEUI:     d.DevEUI.String(),
					FPort:      10,
					JsonObject: `{"Bytes": [4,3,2,1]}`,
				})
				So(err, ShouldBeNil)
				So(validator.ctx, ShouldResemble, ctx)
				So(validator.validatorFuncs, ShouldHaveLength, 1)

				Convey("Then the expected request has been made to the network-server", func() {
					So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
					So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, ns.CreateDeviceQueueItemRequest{
						Item: &ns.DeviceQueueItem{
							DevEUI:     d.DevEUI[:],
							FrmPayload: b,
							FCnt:       12,
							FPort:      10,
						},
					})
				})
			})
		})

		Convey("When enqueueing a downlink queue item", func() {
			_, err := api.Enqueue(ctx, &pb.EnqueueDeviceQueueItemRequest{
				DevEUI: d.DevEUI.String(),
				FPort:  10,
				Data:   []byte{1, 2, 3, 4},
			})
			So(err, ShouldBeNil)
			So(validator.ctx, ShouldResemble, ctx)
			So(validator.validatorFuncs, ShouldHaveLength, 1)

			Convey("Then the expected request has been made to the network-server", func() {
				So(nsClient.CreateDeviceQueueItemChan, ShouldHaveLength, 1)
				So(<-nsClient.CreateDeviceQueueItemChan, ShouldResemble, ns.CreateDeviceQueueItemRequest{
					Item: &ns.DeviceQueueItem{
						DevEUI:     d.DevEUI[:],
						FrmPayload: b,
						FCnt:       12,
						FPort:      10,
					},
				})
			})
		})

		Convey("Given a mocked device-queue item", func() {
			nsClient.GetDeviceQueueItemsForDevEUIResponse = ns.GetDeviceQueueItemsForDevEUIResponse{
				Items: []*ns.DeviceQueueItem{
					{
						DevEUI:     d.DevEUI[:],
						FrmPayload: b,
						FCnt:       12,
						FPort:      10,
						Confirmed:  true,
					},
				},
			}

			Convey("Then list returns the expected item", func() {
				resp, err := api.List(ctx, &pb.ListDeviceQueueItemsRequest{
					DevEUI: d.DevEUI.String(),
				})
				So(err, ShouldBeNil)
				So(resp.Items, ShouldHaveLength, 1)
				So(resp.Items[0], ShouldResemble, &pb.DeviceQueueItem{
					DevEUI:    d.DevEUI.String(),
					Confirmed: true,
					FPort:     10,
					FCnt:      12,
					Data:      []byte{1, 2, 3, 4},
				})
			})
		})

		Convey("Given a device-queue mapping", func() {
			dqm := storage.DeviceQueueMapping{
				DevEUI:    d.DevEUI,
				Reference: "test-123",
				FCnt:      12,
			}
			So(storage.CreateDeviceQueueMapping(config.C.PostgreSQL.DB, &dqm), ShouldBeNil)

			Convey("When calling Flush", func() {
				_, err := api.Flush(ctx, &pb.FlushDeviceQueueRequest{
					DevEUI: d.DevEUI.String(),
				})
				So(err, ShouldBeNil)

				Convey("Then the expected request has been made to the network-server", func() {
					So(nsClient.FlushDeviceQueueForDevEUIChan, ShouldHaveLength, 1)
					So(<-nsClient.FlushDeviceQueueForDevEUIChan, ShouldResemble, ns.FlushDeviceQueueForDevEUIRequest{
						DevEUI: d.DevEUI[:],
					})
				})

				Convey("Then the device-queue mapping has been removed", func() {
					_, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(config.C.PostgreSQL.DB, d.DevEUI, 12)
					So(err, ShouldEqual, storage.ErrDoesNotExist)
				})
			})
		})
	})
}
