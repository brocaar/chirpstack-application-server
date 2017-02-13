package api

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lora-app-server/internal/test/testhandler"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestApplicationServerAPI(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a clean database with application + node and api instance", t, func() {
		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

		app := storage.Application{
			Name: "test-app",
		}
		So(storage.CreateApplication(db, &app), ShouldBeNil)
		node := storage.Node{
			ApplicationID:      app.ID,
			Name:               "test-node",
			DevEUI:             [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:             [8]byte{8, 7, 6, 5, 4, 3, 2, 1},
			AppKey:             [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			RXWindow:           storage.RX2,
			RXDelay:            1,
			RX1DROffset:        2,
			RX2DR:              3,
			RelaxFCnt:          true,
			ADRInterval:        20,
			InstallationMargin: 5,
		}
		So(storage.CreateNode(db, node), ShouldBeNil)

		h := testhandler.NewTestHandler()

		ctx := context.Background()
		lsCtx := common.Context{
			DB:      db,
			Handler: h,
		}

		api := NewApplicationServerAPI(lsCtx)

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

		Convey("Given a join-request", func() {
			phy := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinRequest,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinRequestPayload{
					AppEUI:   node.AppEUI,
					DevEUI:   node.DevEUI,
					DevNonce: [2]byte{1, 2},
				},
			}
			So(phy.SetMIC(node.AppKey), ShouldBeNil)

			b, err := phy.MarshalBinary()
			So(err, ShouldBeNil)

			Convey("When calling HandleDataUp", func() {
				now := time.Now().UTC()
				mac := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

				req := as.HandleDataUpRequest{
					DevEUI: node.DevEUI[:],
					AppEUI: node.AppEUI[:],
					FCnt:   10,
					FPort:  3,
					Data:   []byte{1, 2, 3, 4},
					RxInfo: []*as.RXInfo{
						{
							Mac:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
							Time:    now.Format(time.RFC3339Nano),
							Rssi:    -60,
							LoRaSNR: 5,
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
						DevEUI:          node.DevEUI,
						RXInfo: []handler.RXInfo{
							{
								MAC:     mac,
								Time:    &now,
								RSSI:    -60,
								LoRaSNR: 5,
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

			Convey("Given the node is an ABP device", func() {
				node.IsABP = true
				So(storage.UpdateNode(db, node), ShouldBeNil)

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

			Convey("When calling JoinRequest", func() {

				req := as.JoinRequestRequest{
					PhyPayload: b,
					DevAddr:    []byte{1, 2, 3, 4},
					NetID:      []byte{1, 2, 3},
				}

				resp, err := api.JoinRequest(ctx, &req)
				So(err, ShouldBeNil)

				Convey("Then the expected response is returned", func() {
					node, err := storage.GetNode(db, node.DevEUI)
					So(err, ShouldBeNil)

					So(resp.NwkSKey, ShouldResemble, node.NwkSKey[:])
					So(resp.RxDelay, ShouldEqual, uint32(node.RXDelay))
					So(resp.Rx1DROffset, ShouldEqual, uint32(node.RX1DROffset))
					So(resp.CFList, ShouldHaveLength, 0)
					So(resp.RxWindow, ShouldEqual, as.RXWindow_RX2)
					So(resp.Rx2DR, ShouldEqual, uint32(node.RX2DR))
					So(resp.RelaxFCnt, ShouldBeTrue)
					So(resp.InstallationMargin, ShouldEqual, 5)
					So(resp.AdrInterval, ShouldEqual, 20)

					var phy lorawan.PHYPayload
					So(phy.UnmarshalBinary(resp.PhyPayload), ShouldBeNil)

					So(phy.MHDR.MType, ShouldEqual, lorawan.JoinAccept)
					So(phy.DecryptJoinAcceptPayload(node.AppKey), ShouldBeNil)
					ok, err := phy.ValidateMIC(node.AppKey)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)

					jaPL, ok := phy.MACPayload.(*lorawan.JoinAcceptPayload)
					So(ok, ShouldBeTrue)

					So(jaPL.NetID, ShouldEqual, [3]byte{1, 2, 3})
					So(jaPL.DLSettings, ShouldResemble, lorawan.DLSettings{
						RX2DataRate: node.RX2DR,
						RX1DROffset: node.RX1DROffset,
					})
					So(jaPL.RXDelay, ShouldEqual, node.RXDelay)
					So(jaPL.CFList, ShouldBeNil)
					So(jaPL.DevAddr[:], ShouldResemble, []byte{1, 2, 3, 4})

					Convey("Then the DevAddr of the node has been updated", func() {
						So(node.DevAddr[:], ShouldResemble, []byte{1, 2, 3, 4})
					})
				})

				Convey("Then a notification was sent to the handler", func() {
					So(h.SendJoinNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendJoinNotificationChan, ShouldResemble, handler.JoinNotification{
						ApplicationID:   app.ID,
						ApplicationName: "test-app",
						NodeName:        "test-node",
						DevAddr:         [4]byte{1, 2, 3, 4},
						DevEUI:          node.DevEUI,
					})
				})
			})

			Convey("Given the node as a CFList with three channels", func() {
				cl := storage.ChannelList{
					Name: "test list",
					Channels: []int64{
						868400000,
						868500000,
						868600000,
					},
				}
				So(storage.CreateChannelList(db, &cl), ShouldBeNil)

				node.ChannelListID = &cl.ID
				So(storage.UpdateNode(db, node), ShouldBeNil)

				Convey("When calling JoinRequest", func() {
					req := as.JoinRequestRequest{
						PhyPayload: b,
						DevAddr:    []byte{1, 2, 3, 4},
						NetID:      []byte{1, 2, 3},
					}

					resp, err := api.JoinRequest(ctx, &req)
					So(err, ShouldBeNil)

					Convey("Then the CFlist is set in the response", func() {
						node, err := storage.GetNode(db, node.DevEUI)
						So(err, ShouldBeNil)

						var phy lorawan.PHYPayload
						So(phy.UnmarshalBinary(resp.PhyPayload), ShouldBeNil)

						So(phy.DecryptJoinAcceptPayload(node.AppKey), ShouldBeNil)
						ok, err := phy.ValidateMIC(node.AppKey)
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
							DevEUI:          node.DevEUI,
						})
					})
				})
			})

			Convey("Given a pending downlink queue item for this node", func() {
				qi := storage.DownlinkQueueItem{
					DevEUI:    node.DevEUI,
					Reference: "abcd1234",
					Confirmed: true,
					Pending:   true,
					FPort:     10,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDownlinkQueueItem(db, &qi), ShouldBeNil)

				Convey("Then it is removed when calling HandleDataDownACK", func() {
					_, err := api.HandleDataDownACK(ctx, &as.HandleDataDownACKRequest{
						DevEUI: node.DevEUI[:],
					})
					So(err, ShouldBeNil)

					_, err = storage.GetDownlinkQueueItem(db, qi.ID)
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
				qi := storage.DownlinkQueueItem{
					DevEUI:    node.DevEUI,
					Reference: "abcd1234",
					Confirmed: false,
					FPort:     1,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDownlinkQueueItem(db, &qi), ShouldBeNil)

				Convey("When calling GetDataDown", func() {
					resp, err := api.GetDataDown(ctx, &as.GetDataDownRequest{
						DevEUI:         node.DevEUI[:],
						MaxPayloadSize: 100,
						FCnt:           10,
					})
					So(err, ShouldBeNil)

					Convey("Then the expected response is returned", func() {
						b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, 10, resp.Data)
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
						size, err := storage.GetDownlinkQueueSize(db, node.DevEUI)
						So(err, ShouldBeNil)
						So(size, ShouldEqual, 0)
					})
				})
			})

			Convey("Given a downlink queue item in the queue (confirmed=true)", func() {
				qi := storage.DownlinkQueueItem{
					DevEUI:    node.DevEUI,
					Reference: "abcd1234",
					Confirmed: true,
					FPort:     1,
					Data:      []byte{1, 2, 3, 4},
				}
				So(storage.CreateDownlinkQueueItem(db, &qi), ShouldBeNil)

				Convey("When calling GetDataDown", func() {
					resp, err := api.GetDataDown(ctx, &as.GetDataDownRequest{
						DevEUI:         node.DevEUI[:],
						MaxPayloadSize: 100,
						FCnt:           10,
					})
					So(err, ShouldBeNil)

					Convey("Then the expected response is returned", func() {
						b, err := lorawan.EncryptFRMPayload(node.AppSKey, false, node.DevAddr, 10, resp.Data)
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
						qi2, err := storage.GetDownlinkQueueItem(db, qi.ID)
						So(err, ShouldBeNil)
						So(qi2.Pending, ShouldBeTrue)
					})
				})
			})
		})
	})
}
