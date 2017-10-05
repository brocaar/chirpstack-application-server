package api

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

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
	common.NetworkServer = nsClient
	common.RedisPool = storage.NewRedisPool(conf.RedisURL)

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

		dc := storage.DeviceCredentials{
			DevEUI: d.DevEUI,
			AppKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDeviceCredentials(common.DB, &dc), ShouldBeNil)

		gw := storage.Gateway{
			MAC:            lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			Name:           "test-gw",
			Description:    "test gateway",
			OrganizationID: org.ID,
		}
		So(storage.CreateGateway(common.DB, &gw), ShouldBeNil)

		h := testhandler.NewTestHandler()
		common.Handler = h

		ctx := context.Background()
		api := NewApplicationServerAPI()

		Convey("When calling HandleError", func() {
			_, err := api.HandleError(ctx, &as.HandleErrorRequest{
				DevEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				AppEUI: []byte{8, 7, 6, 5, 4, 3, 2, 1},
				Type:   as.ErrorType_DATA_UP_FCNT,
				Error:  "BOOM!",
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

			Convey("When calling HandleDataUp", func() {
				now := time.Now().UTC()
				mac := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

				req := as.HandleDataUpRequest{
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
				_, err := api.HandleDataUp(ctx, &req)
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

			Convey("Given a pending downlink queue item for this node", func() {
				qi := storage.DeviceQueueItem{
					DevEUI:    d.DevEUI,
					Reference: "abcd1234",
					Confirmed: true,
					Pending:   true,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDeviceQueueItem(common.DB, &qi), ShouldBeNil)

				Convey("Then it is removed when calling HandleDataDownACK", func() {
					_, err := api.HandleDataDownACK(ctx, &as.HandleDataDownACKRequest{
						DevEUI: d.DevEUI[:],
					})
					So(err, ShouldBeNil)

					_, err = storage.GetDeviceQueueItem(common.DB, qi.ID)
					So(err, ShouldNotBeNil)

					Convey("Then an ack notification was sent to the handler", func() {
						So(h.SendACKNotificationChan, ShouldHaveLength, 1)
						So(<-h.SendACKNotificationChan, ShouldResemble, handler.ACKNotification{
							ApplicationID:   app.ID,
							ApplicationName: "test-app",
							NodeName:        "test-node",
							DevEUI:          qi.DevEUI,
							Reference:       qi.Reference,
						})
					})
				})
			})

			Convey("Given a downlink queue item in the queue (confirmed=false)", func() {
				qi := storage.DeviceQueueItem{
					DevEUI:    d.DevEUI,
					Reference: "abcd1234",
					Confirmed: false,
					FPort:     1,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDeviceQueueItem(common.DB, &qi), ShouldBeNil)

				Convey("When calling GetDataDown", func() {
					resp, err := api.GetDataDown(ctx, &as.GetDataDownRequest{
						DevEUI:         d.DevEUI[:],
						MaxPayloadSize: 100,
						FCnt:           10,
					})
					So(err, ShouldBeNil)

					Convey("Then the expected response is returned", func() {
						da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
						So(err, ShouldBeNil)

						b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 10, resp.Data)
						So(err, ShouldBeNil)

						resp.Data = b

						So(resp, ShouldResemble, &as.GetDataDownResponse{
							Data:      qi.Data,
							Confirmed: false,
							FPort:     1,
							MoreData:  false,
						})
					})

					Convey("Then the item was removed from the queue", func() {
						size, err := storage.GetDeviceQueueItemCount(common.DB, d.DevEUI)
						So(err, ShouldBeNil)
						So(size, ShouldEqual, 0)
					})
				})
			})

			Convey("Given a downlink queue item in the queue (confirmed=true)", func() {
				qi := storage.DeviceQueueItem{
					DevEUI:    d.DevEUI,
					Reference: "abcd1234",
					Confirmed: true,
					FPort:     1,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDeviceQueueItem(common.DB, &qi), ShouldBeNil)

				Convey("When calling GetDataDown", func() {
					resp, err := api.GetDataDown(ctx, &as.GetDataDownRequest{
						DevEUI:         d.DevEUI[:],
						MaxPayloadSize: 100,
						FCnt:           10,
					})
					So(err, ShouldBeNil)

					Convey("Then the expected response is returned", func() {
						da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
						So(err, ShouldBeNil)

						b, err := lorawan.EncryptFRMPayload(da.AppSKey, false, da.DevAddr, 10, resp.Data)
						So(err, ShouldBeNil)

						resp.Data = b

						So(resp, ShouldResemble, &as.GetDataDownResponse{
							Data:      qi.Data,
							Confirmed: true,
							FPort:     1,
							MoreData:  false,
						})
					})

					Convey("Then the item was set to pending", func() {
						qi2, err := storage.GetDeviceQueueItem(common.DB, qi.ID)
						So(err, ShouldBeNil)
						So(qi2.Pending, ShouldBeTrue)
					})
				})
			})
		})

		Convey("Given a join-request", func() {
			phy := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinRequest,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinRequestPayload{
					DevEUI:   d.DevEUI,
					DevNonce: [2]byte{1, 2},
				},
			}
			So(phy.SetMIC(dc.AppKey), ShouldBeNil)

			b, err := phy.MarshalBinary()
			So(err, ShouldBeNil)

			Convey("Given the node is an ABP device", func() {
				nsClient.GetDeviceProfileResponse.DeviceProfile.SupportsJoin = false

				Convey("When calling JoinRequest", func() {
					req := as.JoinRequestRequest{
						PhyPayload: b,
						DevAddr:    []byte{1, 2, 3, 4},
						NetID:      []byte{1, 2, 3},
					}

					_, err := api.JoinRequest(ctx, &req)

					Convey("Then an error was returned", func() {
						So(err, ShouldResemble, grpc.Errorf(codes.FailedPrecondition, "node is ABP device"))
					})
				})
			})

			Convey("When calling JoinRequest without any additional channels", func() {
				req := as.JoinRequestRequest{
					PhyPayload: b,
					DevAddr:    []byte{1, 2, 3, 4},
					NetID:      []byte{1, 2, 3},
				}

				resp, err := api.JoinRequest(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the expected response is returned", func() {
					da, err := storage.GetLastDeviceActivationForDevEUI(common.DB, d.DevEUI)
					So(err, ShouldBeNil)

					So(resp.NwkSKey, ShouldResemble, da.NwkSKey[:])
					So(resp.RxDelay, ShouldEqual, uint32(dp.DeviceProfile.RXDelay1))
					So(resp.Rx1DROffset, ShouldEqual, uint32(dp.DeviceProfile.RXDROffset1))
					So(resp.Rx2DR, ShouldEqual, uint32(dp.DeviceProfile.RXDataRate2))

					var phy lorawan.PHYPayload
					So(phy.UnmarshalBinary(resp.PhyPayload), ShouldBeNil)

					So(phy.MHDR.MType, ShouldEqual, lorawan.JoinAccept)
					So(phy.DecryptJoinAcceptPayload(dc.AppKey), ShouldBeNil)
					ok, err := phy.ValidateMIC(dc.AppKey)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)

					jaPL, ok := phy.MACPayload.(*lorawan.JoinAcceptPayload)
					So(ok, ShouldBeTrue)

					So(jaPL.NetID, ShouldEqual, lorawan.NetID{1, 2, 3})
					So(jaPL.DLSettings, ShouldResemble, lorawan.DLSettings{
						RX2DataRate: uint8(dp.DeviceProfile.RXDataRate2),
						RX1DROffset: uint8(dp.DeviceProfile.RXDROffset1),
					})
					So(jaPL.RXDelay, ShouldEqual, uint8(dp.DeviceProfile.RXDelay1))
					So(jaPL.CFList, ShouldBeNil)
					So(jaPL.DevAddr[:], ShouldResemble, []byte{1, 2, 3, 4})

					Convey("Then the DevAddr of the node has been updated", func() {
						So(da.DevAddr[:], ShouldResemble, []byte{1, 2, 3, 4})
					})
				})

				Convey("Then a notification was sent to the handler", func() {
					So(h.SendJoinNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendJoinNotificationChan, ShouldResemble, handler.JoinNotification{
						ApplicationID:   app.ID,
						ApplicationName: "test-app",
						NodeName:        "test-node",
						DevAddr:         [4]byte{1, 2, 3, 4},
						DevEUI:          d.DevEUI,
					})
				})
			})

			Convey("When calling JoinRequest with additional channels", func() {
				req := as.JoinRequestRequest{
					PhyPayload: b,
					DevAddr:    []byte{1, 2, 3, 4},
					NetID:      []byte{1, 2, 3},
					CFList: []uint32{
						868400000,
						868500000,
						868600000,
					},
				}
				resp, err := api.JoinRequest(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the CFlist is set in the response", func() {
					var phy lorawan.PHYPayload
					So(phy.UnmarshalBinary(resp.PhyPayload), ShouldBeNil)

					So(phy.DecryptJoinAcceptPayload(dc.AppKey), ShouldBeNil)
					ok, err := phy.ValidateMIC(dc.AppKey)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)

					jaPL, ok := phy.MACPayload.(*lorawan.JoinAcceptPayload)
					So(ok, ShouldBeTrue)

					So(jaPL.CFList, ShouldResemble, &lorawan.CFList{
						868400000,
						868500000,
						868600000,
						0,
						0,
					})
				})

				Convey("Then a notification was sent to the handler", func() {
					So(h.SendJoinNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendJoinNotificationChan, ShouldResemble, handler.JoinNotification{
						ApplicationID:   app.ID,
						ApplicationName: "test-app",
						NodeName:        "test-node",
						DevAddr:         [4]byte{1, 2, 3, 4},
						DevEUI:          d.DevEUI,
					})
				})
			})
		})
	})
}
