package api

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lora-app-server/internal/test/testhandler"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestApplicationServerAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	nsClient := test.NewNetworkServerClient()

	common.DB = db
	common.RedisPool = storage.NewRedisPool(conf.RedisURL)
	common.NetworkServerPool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with bootstrap data node and api instance", t, func() {
		test.MustResetDB(common.DB)

		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{
				SupportsJoin: true,
			},
		}

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
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(common.DB, &sp), ShouldBeNil)

		dp := storage.DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(common.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
			Name:             "test-app",
		}
		So(storage.CreateApplication(common.DB, &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
		}
		So(storage.CreateDevice(common.DB, &d), ShouldBeNil)

		dc := storage.DeviceKeys{
			DevEUI: d.DevEUI,
			AppKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDeviceKeys(common.DB, &dc), ShouldBeNil)

		gw := storage.Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gw",
			Description:     "test gateway",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateGateway(common.DB, &gw), ShouldBeNil)

		h := testhandler.NewTestHandler()
		common.Handler = h

		ctx := context.Background()
		api := NewApplicationServerAPI()

		Convey("When calling HandleError", func() {
			_, err := api.HandleError(ctx, &as.HandleErrorRequest{
				DevEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Type:   as.ErrorType_DATA_UP_FCNT,
				Error:  "BOOM!",
				FCnt:   123,
			})
			So(err, ShouldBeNil)

			Convey("Then the error has been sent to the handler", func() {
				So(h.SendErrorNotificationChan, ShouldHaveLength, 1)
				So(<-h.SendErrorNotificationChan, ShouldResemble, handler.ErrorNotification{
					ApplicationID:   app.ID,
					ApplicationName: "test-app",
					NodeName:        "test-node",
					DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
					Type:            "DATA_UP_FCNT",
					Error:           "BOOM!",
					FCnt:            123,
				})
			})
		})

		Convey("Given the device is activated", func() {
			da := storage.DeviceActivation{
				DevEUI:  d.DevEUI,
				DevAddr: lorawan.DevAddr{},
				AppSKey: lorawan.AES128Key{},
				NwkSKey: lorawan.AES128Key{},
			}
			So(storage.CreateDeviceActivation(common.DB, &da), ShouldBeNil)

			Convey("When calling HandleUplinkData", func() {
				now := time.Now().UTC()
				mac := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

				req := as.HandleUplinkDataRequest{
					DevEUI: d.DevEUI[:],
					FCnt:   10,
					FPort:  3,
					Data:   []byte{1, 2, 3, 4},
					RxInfo: []*as.RXInfo{
						{
							Mac:       []byte{1, 2, 3, 4, 5, 6, 7, 8},
							Name:      "test-gateway",
							Latitude:  52.3740364,
							Longitude: 4.9144401,
							Altitude:  10,
							Time:      now.Format(time.RFC3339Nano),
							Rssi:      -60,
							LoRaSNR:   5,
						},
					},
					TxInfo: &as.TXInfo{
						Frequency: 868100000,
						DataRate: &as.DataRate{
							Modulation:   "LORA",
							BandWidth:    250,
							SpreadFactor: 5,
							Bitrate:      50000,
						},
						Adr:      true,
						CodeRate: "4/6",
					},
				}
				_, err := api.HandleUplinkData(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the expected payload was sent to the handler", func() {
					So(h.SendDataUpChan, ShouldHaveLength, 1)
					So(<-h.SendDataUpChan, ShouldResemble, handler.DataUpPayload{
						ApplicationID:   app.ID,
						ApplicationName: "test-app",
						NodeName:        "test-node",
						DevEUI:          d.DevEUI,
						RXInfo: []handler.RXInfo{
							{
								MAC:       mac,
								Name:      "test-gateway",
								Latitude:  52.3740364,
								Longitude: 4.9144401,
								Altitude:  10,
								Time:      &now,
								RSSI:      -60,
								LoRaSNR:   5,
							},
						},
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DataRate: handler.DataRate{
								Modulation:   "LORA",
								Bandwidth:    250,
								SpreadFactor: 5,
								Bitrate:      50000,
							},
							ADR:      true,
							CodeRate: "4/6",
						},
						FCnt:  10,
						FPort: 3,
						Data:  []byte{67, 216, 236, 205},
					})
				})
			})

			Convey("Given a device-queue mapping", func() {
				dqm := storage.DeviceQueueMapping{
					Reference: "test-1234",
					DevEUI:    d.DevEUI,
					FCnt:      10,
				}
				So(storage.CreateDeviceQueueMapping(common.DB, &dqm), ShouldBeNil)

				Convey("On HandleDownlinkACK (ack: true)", func() {
					_, err := api.HandleDownlinkACK(ctx, &as.HandleDownlinkACKRequest{
						DevEUI:       d.DevEUI[:],
						FCnt:         10,
						Acknowledged: true,
					})
					So(err, ShouldBeNil)

					Convey("Then the device-queue mapping has been removed", func() {
						_, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(common.DB, d.DevEUI, 10)
						So(err, ShouldEqual, storage.ErrDoesNotExist)
					})

					Convey("Then an ack (true) notification was sent to the handler", func() {
						So(h.SendACKNotificationChan, ShouldHaveLength, 1)
						So(<-h.SendACKNotificationChan, ShouldResemble, handler.ACKNotification{
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
							NodeName:        d.Name,
							DevEUI:          d.DevEUI,
							Reference:       dqm.Reference,
							Acknowledged:    true,
							FCnt:            10,
						})
					})
				})

				Convey("On HandleDownlinkACK (ack: false)", func() {
					_, err := api.HandleDownlinkACK(ctx, &as.HandleDownlinkACKRequest{
						DevEUI:       d.DevEUI[:],
						FCnt:         10,
						Acknowledged: false,
					})
					So(err, ShouldBeNil)

					Convey("Then the device-queue mapping has been removed", func() {
						_, err := storage.GetDeviceQueueMappingForDevEUIAndFCnt(common.DB, d.DevEUI, 10)
						So(err, ShouldEqual, storage.ErrDoesNotExist)
					})

					Convey("Then an ack (true) notification was sent to the handler", func() {
						So(h.SendACKNotificationChan, ShouldHaveLength, 1)
						So(<-h.SendACKNotificationChan, ShouldResemble, handler.ACKNotification{
							ApplicationID:   app.ID,
							ApplicationName: app.Name,
							NodeName:        d.Name,
							DevEUI:          d.DevEUI,
							Reference:       dqm.Reference,
							Acknowledged:    false,
							FCnt:            10,
						})
					})
				})
			})
		})
	})
}
