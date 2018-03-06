package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/test/testhandler"

	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestJoinServerAPI(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db

	Convey("Given a clean database with a device", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		h := testhandler.NewTestHandler()
		config.C.ApplicationServer.Integration.Handler = h

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
			ServiceProfile:  backend.ServiceProfile{},
		}
		So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)

		dp := storage.DeviceProfile{
			OrganizationID:  org.ID,
			NetworkServerID: n.ID,
			Name:            "test-dp",
			DeviceProfile:   backend.DeviceProfile{},
		}
		So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)

		app := storage.Application{
			OrganizationID:   org.ID,
			ServiceProfileID: sp.ServiceProfile.ServiceProfileID,
			Name:             "test-app",
		}
		So(storage.CreateApplication(config.C.PostgreSQL.DB, &app), ShouldBeNil)

		d := storage.Device{
			ApplicationID:   app.ID,
			Name:            "test-device",
			DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceProfileID: dp.DeviceProfile.DeviceProfileID,
		}
		So(storage.CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

		dk := storage.DeviceKeys{
			DevEUI: d.DevEUI,
			AppKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}
		So(storage.CreateDeviceKeys(config.C.PostgreSQL.DB, &dk), ShouldBeNil)

		Convey("Given a test-server", func() {
			api := JoinServerAPI{}
			server := httptest.NewServer(&api)
			defer server.Close()

			Convey("When making a JoinReq call", func() {
				jrPHY := lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.JoinRequest,
						Major: lorawan.LoRaWANR1,
					},
					MACPayload: &lorawan.JoinRequestPayload{
						DevEUI:   d.DevEUI,
						AppEUI:   lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
						DevNonce: lorawan.DevNonce{1, 2},
					},
				}
				So(jrPHY.SetMIC(dk.AppKey), ShouldBeNil)
				jrPHYBytes, err := jrPHY.MarshalBinary()
				So(err, ShouldBeNil)

				jaPHY := lorawan.PHYPayload{
					MHDR: lorawan.MHDR{
						MType: lorawan.JoinAccept,
						Major: lorawan.LoRaWANR1,
					},
					MACPayload: &lorawan.JoinAcceptPayload{
						AppNonce: lorawan.AppNonce{1, 0, 0},
						NetID:    lorawan.NetID{1, 2, 3},
						DevAddr:  lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RXDelay: 1,
						CFList:  &lorawan.CFList{868700000, 868900000},
					},
				}
				So(jaPHY.SetMIC(dk.AppKey), ShouldBeNil)
				So(jaPHY.EncryptJoinAcceptPayload(dk.AppKey), ShouldBeNil)
				jaPHYBytes, err := jaPHY.MarshalBinary()
				So(err, ShouldBeNil)

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
					CFList:  &lorawan.CFList{868700000, 868900000},
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
							AESKey: lorawan.AES128Key{223, 83, 195, 95, 48, 52, 204, 206, 208, 255, 53, 76, 112, 222, 4, 223},
						},
					})
				})

				Convey("Then a join notification was sent", func() {
					So(h.SendJoinNotificationChan, ShouldHaveLength, 1)
					So(<-h.SendJoinNotificationChan, ShouldResemble, handler.JoinNotification{
						ApplicationID:   app.ID,
						ApplicationName: app.Name,
						DeviceName:      d.Name,
						DevEUI:          d.DevEUI,
						DevAddr:         lorawan.DevAddr{1, 2, 3, 4},
					})
				})
			})
		})
	})
}
