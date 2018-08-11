package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/gofrs/uuid"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lora-app-server/internal/test/testhandler"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/common"
	gwPB "github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

func TestApplicationServerAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	nsClient := test.NewNetworkServerClient()

	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = storage.NewRedisPool(conf.RedisURL)
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

	Convey("Given a clean database with bootstrap data node and api instance", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
			DeviceProfile: &ns.DeviceProfile{
				SupportsJoin: true,
			},
		}

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
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := storage.DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)
		dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
		So(err, ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			ServiceProfileID: spID,
			Name:             "test-app",
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			Name:            "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceProfileID: dpID,
		}
		So(storage.CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

		dc := storage.DeviceKeys{
			DevEUI: d.DevEUI,
			NwkKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDeviceKeys(config.C.PostgreSQL.DB, &dc), ShouldBeNil)

		gw := storage.Gateway{
			MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:            "test-gw",
			Description:     "test gateway",
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
		}
		So(storage.CreateGateway(config.C.PostgreSQL.DB, &gw), ShouldBeNil)

		h := testhandler.NewTestHandler()
		config.C.ApplicationServer.Integration.Handler = h

		ctx := context.Background()
		api := NewApplicationServerAPI()

		Convey("When calling HandleError", func() {
			_, err := api.HandleError(ctx, &as.HandleErrorRequest{
				DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
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
					DeviceName:      "test-node",
					DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
					Type:            "DATA_UP_FCNT",
					Error:           "BOOM!",
					FCnt:            123,
				})
			})
		})

		Convey("Given the device is not activate but HandleUplinkDataRequest contains DeviceSecurityContext", func() {
			now := time.Now().UTC()

			req := as.HandleUplinkDataRequest{
				DevEui: d.DevEUI[:],
				FCnt:   10,
				FPort:  3,
				Dr:     6,
				Adr:    true,
				Data:   []byte{1, 2, 3, 4},
				RxInfo: []*gwPB.UplinkRXInfo{
					{
						GatewayId: gw.MAC[:],
						Rssi:      -60,
						LoraSnr:   5,
						Location: &gwPB.Location{
							Latitude:  52.3740364,
							Longitude: 4.9144401,
							Altitude:  10,
						},
					},
				},
				TxInfo: &gwPB.UplinkTXInfo{
					Frequency:  868100000,
					Modulation: common.Modulation_LORA,
					ModulationInfo: &gwPB.UplinkTXInfo_LoraModulationInfo{
						LoraModulationInfo: &gwPB.LoRaModulationInfo{
							Bandwidth:       250,
							SpreadingFactor: 5,
							CodeRate:        "4/6",
						},
					},
				},
				DeviceActivationContext: &as.DeviceActivationContext{
					DevAddr: []byte{1, 2, 3, 4},
					AppSKey: &common.KeyEnvelope{
						AesKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
					},
				},
			}

			req.RxInfo[0].Time, _ = ptypes.TimestampProto(now)

			Convey("When calling HandleUplinkData", func() {
				_, err := api.HandleUplinkData(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then a payload was sent to the handler", func() {
					So(h.SendDataUpChan, ShouldHaveLength, 1)
					pl := <-h.SendDataUpChan
					So(pl.Data, ShouldResemble, []byte{33, 99, 53, 13})
				})

				Convey("Then a join-notification was sent to the handler", func() {
					So(h.SendJoinNotificationChan, ShouldHaveLength, 1)
					pl := <-h.SendJoinNotificationChan
					So(pl.DevAddr, ShouldEqual, lorawan.DevAddr{1, 2, 3, 4})
				})
			})
		})

		Convey("Given the device is activated", func() {
			da := storage.DeviceActivation{
				DevEUI:  d.DevEUI,
				DevAddr: lorawan.DevAddr{},
				AppSKey: lorawan.AES128Key{},
			}
			So(storage.CreateDeviceActivation(config.C.PostgreSQL.DB, &da), ShouldBeNil)

			now := time.Now().UTC()
			mac := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

			req := as.HandleUplinkDataRequest{
				DevEui: d.DevEUI[:],
				FCnt:   10,
				FPort:  3,
				Dr:     6,
				Adr:    true,
				Data:   []byte{1, 2, 3, 4},
				RxInfo: []*gwPB.UplinkRXInfo{
					{
						GatewayId: gw.MAC[:],
						Rssi:      -60,
						LoraSnr:   5,
						Location: &gwPB.Location{
							Latitude:  52.3740364,
							Longitude: 4.9144401,
							Altitude:  10,
						},
					},
				},
				TxInfo: &gwPB.UplinkTXInfo{
					Frequency:  868100000,
					Modulation: common.Modulation_LORA,
					ModulationInfo: &gwPB.UplinkTXInfo_LoraModulationInfo{
						LoraModulationInfo: &gwPB.LoRaModulationInfo{
							Bandwidth:       250,
							SpreadingFactor: 5,
							CodeRate:        "4/6",
						},
					},
				},
			}

			req.RxInfo[0].Time, _ = ptypes.TimestampProto(now)

			Convey("When calling HandleUplinkData", func() {
				_, err := api.HandleUplinkData(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then a payload was sent to the handler", func() {
					So(h.SendDataUpChan, ShouldHaveLength, 1)
				})

				Convey("Then the device was updated", func() {
					d, err := storage.GetDevice(config.C.PostgreSQL.DB, d.DevEUI, false, true)
					So(err, ShouldBeNil)
					So(d.DeviceStatusBattery, ShouldBeNil)
					So(d.DeviceStatusMargin, ShouldBeNil)
					So(time.Now().Sub(*d.LastSeenAt), ShouldBeLessThan, time.Second)
				})
			})

			Convey("When calling HandleUplinkData (Custom JS codec configured)", func() {
				app.PayloadCodec = codec.CustomJSType
				app.PayloadDecoderScript = `
					function Decode(fPort, bytes) {
						return {
							"fPort": fPort,
							"firstByte": bytes[0]
						}
					}
				`
				So(storage.UpdateApplication(config.C.PostgreSQL.DB, app), ShouldBeNil)

				_, err := api.HandleUplinkData(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the object fields has been set in the payload", func() {
					So(h.SendDataUpChan, ShouldHaveLength, 1)
					pl := <-h.SendDataUpChan
					So(pl.Object, ShouldNotBeNil)
					b, err := json.Marshal(pl.Object)
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"fPort":3,"firstByte":67}`)
				})
			})

			Convey("When calling HandleUplinkData (no codec configured)", func() {
				_, err := api.HandleUplinkData(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the expected payload was sent to the handler", func() {
					So(h.SendDataUpChan, ShouldHaveLength, 1)
					So(<-h.SendDataUpChan, ShouldResemble, handler.DataUpPayload{
						ApplicationID:   app.ID,
						ApplicationName: "test-app",
						DeviceName:      "test-node",
						DevEUI:          d.DevEUI,
						RXInfo: []handler.RXInfo{
							{
								GatewayID: mac,
								Name:      "test-gw",
								Location: &handler.Location{
									Latitude:  52.3740364,
									Longitude: 4.9144401,
									Altitude:  10,
								},
								Time:    &now,
								RSSI:    -60,
								LoRaSNR: 5,
							},
						},
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DR:        6,
						},
						ADR:   true,
						FCnt:  10,
						FPort: 3,
						Data:  []byte{67, 216, 236, 205},
					})
				})
			})

			Convey("When calling SetDeviceStatus", func() {
				_, err := api.SetDeviceStatus(ctx, &as.SetDeviceStatusRequest{
					DevEui:  d.DevEUI[:],
					Margin:  10,
					Battery: 123,
				})
				So(err, ShouldBeNil)

				Convey("Then the expected payload was sent to the handler", func() {
					So(h.SendStatusNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendStatusNotificationChan, ShouldResemble, handler.StatusNotification{
						ApplicationID:   app.ID,
						ApplicationName: app.Name,
						DeviceName:      d.Name,
						DevEUI:          d.DevEUI,
						Margin:          10,
						Battery:         123,
					})
				})

				Convey("Then the device has been updated", func() {
					d, err := storage.GetDevice(db, d.DevEUI, false, true)
					So(err, ShouldBeNil)
					So(d.DeviceStatusMargin, ShouldNotBeNil)
					So(d.DeviceStatusBattery, ShouldNotBeNil)
					So(*d.DeviceStatusMargin, ShouldEqual, 10)
					So(*d.DeviceStatusBattery, ShouldEqual, 123)
				})
			})

			Convey("On HandleDownlinkACK (ack: true)", func() {
				_, err := api.HandleDownlinkACK(ctx, &as.HandleDownlinkACKRequest{
					DevEui:       d.DevEUI[:],
					FCnt:         10,
					Acknowledged: true,
				})
				So(err, ShouldBeNil)

				Convey("Then an ack (true) notification was sent to the handler", func() {
					So(h.SendACKNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendACKNotificationChan, ShouldResemble, handler.ACKNotification{
						ApplicationID:   app.ID,
						ApplicationName: app.Name,
						DeviceName:      d.Name,
						DevEUI:          d.DevEUI,
						Acknowledged:    true,
						FCnt:            10,
					})
				})
			})
		})

	})
}
