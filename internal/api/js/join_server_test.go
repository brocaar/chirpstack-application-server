package js

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestJoinServerAPI(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database with a device", t, func() {
		test.MustResetDB(storage.DB().DB)
		test.MustFlushRedis(storage.RedisPool())

		nsClient := nsmock.NewClient()
		networkserver.SetPool(nsmock.NewPool(nsClient))

		h := mock.New()
		integration.SetIntegration(h)

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
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-sp",
		}
		So(storage.CreateServiceProfile(storage.DB(), &sp), ShouldBeNil)
		spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
		So(err, ShouldBeNil)

		dp := storage.DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
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
			Name:            "test-device",
			DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceProfileID: dpID,
		}
		So(storage.CreateDevice(storage.DB(), &d), ShouldBeNil)

		dk := storage.DeviceKeys{
			DevEUI:    d.DevEUI,
			NwkKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			JoinNonce: 65535,
		}
		So(storage.CreateDeviceKeys(storage.DB(), &dk), ShouldBeNil)

		Convey("Given a test-server", func() {
			api := JoinServerAPI{}
			server := httptest.NewServer(&api)
			defer server.Close()

			cFList := lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						868700000,
						868900000,
					},
				},
			}
			cFListB, err := cFList.MarshalBinary()
			So(err, ShouldBeNil)

			Convey("When making a JoinReq call", func() {
				jrPHY := lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.JoinRequest,
						Major: lorawan.LoRaWANR1,
					},
					MACPayload: &lorawan.JoinRequestPayload{
						DevEUI:   d.DevEUI,
						JoinEUI:  lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
						DevNonce: 258,
					},
				}
				So(jrPHY.SetUplinkJoinMIC(dk.NwkKey), ShouldBeNil)
				jrPHYBytes, err := jrPHY.MarshalBinary()
				So(err, ShouldBeNil)

				jaPHY := lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.JoinAccept,
						Major: lorawan.LoRaWANR1,
					},
					MACPayload: &lorawan.JoinAcceptPayload{
						JoinNonce: 65536,
						HomeNetID: lorawan.NetID{1, 2, 3},
						DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RXDelay: 1,
						CFList: &lorawan.CFList{
							CFListType: lorawan.CFListChannel,
							Payload: &lorawan.CFListChannelPayload{
								Channels: [5]uint32{
									868700000,
									868900000,
								},
							},
						},
					},
				}
				So(jaPHY.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 513, dk.NwkKey), ShouldBeNil)
				So(jaPHY.EncryptJoinAcceptPayload(dk.NwkKey), ShouldBeNil)
				jaPHYBytes, err := jaPHY.MarshalBinary()
				So(err, ShouldBeNil)
				So(jaPHYBytes, ShouldResemble, []byte{32, 38, 244, 178, 71, 240, 165, 215, 228, 106, 114, 14, 97, 200, 188, 203, 197, 23, 159, 69, 102, 225, 133, 237, 104, 137, 88, 155, 177, 169, 198, 140, 192})

				joinReqPayload := backend.JoinReqPayload{
					BasePayload: backend.BasePayload{
						ProtocolVersion: backend.ProtocolVersion1_0,
						SenderID:        "010203",
						ReceiverID:      "0807060504030201",
						TransactionID:   1234,
						MessageType:     backend.JoinReq,
					},
					MACVersion: "1.0.2",
					PHYPayload: backend.HEXBytes(jrPHYBytes),
					DevEUI:     d.DevEUI,
					DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						RX2DataRate: 5,
						RX1DROffset: 1,
					},
					RxDelay: 1,
					CFList:  backend.HEXBytes(cFListB),
				}
				joinReqPayloadJSON, err := json.Marshal(joinReqPayload)
				So(err, ShouldBeNil)

				req, err := http.NewRequest("POST", server.URL, bytes.NewReader(joinReqPayloadJSON))
				So(err, ShouldBeNil)

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				Convey("Then the expected response is returned", func() {
					var joinAnsPayload backend.JoinAnsPayload
					So(json.NewDecoder(resp.Body).Decode(&joinAnsPayload), ShouldBeNil)
					So(joinAnsPayload, ShouldResemble, backend.JoinAnsPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "0807060504030201",
							ReceiverID:      "010203",
							TransactionID:   1234,
							MessageType:     backend.JoinAns,
						},
						Result: backend.Result{
							ResultCode: backend.Success,
						},
						PHYPayload: backend.HEXBytes(jaPHYBytes),
						NwkSKey: &backend.KeyEnvelope{
							AESKey: []byte{223, 83, 195, 95, 48, 52, 204, 206, 208, 255, 53, 76, 112, 222, 4, 223},
						},
						AppSKey: &backend.KeyEnvelope{
							AESKey: []byte{146, 123, 156, 145, 17, 131, 207, 254, 76, 178, 255, 75, 117, 84, 95, 109},
						},
					})
				})

				Convey("Then the expected keys are stored", func() {
					dk, err := storage.GetDeviceKeys(storage.DB(), d.DevEUI)
					So(err, ShouldBeNil)

					So(dk.JoinNonce, ShouldEqual, 65536)
				})
			})

			Convey("When making a RejoinReq call", func() {
				rjPHY := lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.RejoinRequest,
						Major: lorawan.LoRaWANR1,
					},
					MACPayload: &lorawan.RejoinRequestType02Payload{
						RejoinType: lorawan.RejoinRequestType2,
						NetID:      lorawan.NetID{1, 2, 3},
						DevEUI:     d.DevEUI,
						RJCount0:   123,
					},
				}
				// no need to set the mic as this is validated by the network-server
				rjPHYBytes, err := rjPHY.MarshalBinary()
				So(err, ShouldBeNil)

				rejoinReqPayload := backend.RejoinReqPayload{
					BasePayload: backend.BasePayload{
						ProtocolVersion: backend.ProtocolVersion1_0,
						SenderID:        "010203",
						ReceiverID:      "0807060504030201",
						TransactionID:   1234,
						MessageType:     backend.RejoinReq,
					},
					MACVersion: "1.1.0",
					PHYPayload: backend.HEXBytes(rjPHYBytes),
					DevEUI:     d.DevEUI,
					DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						RX2DataRate: 5,
						RX1DROffset: 1,
					},
					RxDelay: 1,
					CFList:  backend.HEXBytes(cFListB),
				}
				rejoinReqPayloadJSON, err := json.Marshal(rejoinReqPayload)
				So(err, ShouldBeNil)

				req, err := http.NewRequest("POST", server.URL, bytes.NewReader(rejoinReqPayloadJSON))
				So(err, ShouldBeNil)

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)

				Convey("Then the expected response is returned", func() {
					var rejoinAnsPayload backend.RejoinAnsPayload
					So(json.NewDecoder(resp.Body).Decode(&rejoinAnsPayload), ShouldBeNil)
					So(rejoinAnsPayload, ShouldResemble, backend.RejoinAnsPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "0807060504030201",
							ReceiverID:      "010203",
							TransactionID:   1234,
							MessageType:     backend.RejoinAns,
						},
						Result: backend.Result{
							ResultCode: backend.Success,
						},
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: []byte{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
						},
						FNwkSIntKey: &backend.KeyEnvelope{
							AESKey: []byte{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: []byte{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
						},
						AppSKey: &backend.KeyEnvelope{
							AESKey: []byte{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
						},
						PHYPayload: backend.HEXBytes([]byte{32, 119, 168, 146, 89, 229, 41, 109, 112, 191, 64, 133, 175, 89, 101, 194, 76, 190, 109, 70, 29, 106, 9, 76, 214, 165, 255, 143, 250, 27, 248, 233, 75}),
					})
				})
			})
		})
	})
}
