package join

import (
	"fmt"
	"testing"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/test/testhandler"

	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
	. "github.com/smartystreets/goconvey/convey"
)

func TestJoin(t *testing.T) {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}

	config.C.PostgreSQL.DB = db

	Convey("Given a clean database with node", t, func() {
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

		Convey("Given a set of tests", func() {
			validJRPHY := lorawan.PHYPayload{
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
			So(validJRPHY.SetMIC(dk.AppKey), ShouldBeNil)
			validJRPHYBytes, err := validJRPHY.MarshalBinary()
			So(err, ShouldBeNil)

			invalidMICJRPHY := lorawan.PHYPayload{
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
			So(invalidMICJRPHY.SetMIC(lorawan.AES128Key{}), ShouldBeNil)
			invalidMICJRPHYBytes, err := invalidMICJRPHY.MarshalBinary()
			So(err, ShouldBeNil)

			validJAPHY := lorawan.PHYPayload{
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
			So(validJAPHY.SetMIC(dk.AppKey), ShouldBeNil)
			So(validJAPHY.EncryptJoinAcceptPayload(dk.AppKey), ShouldBeNil)
			validJAPHYBytes, err := validJAPHY.MarshalBinary()
			So(err, ShouldBeNil)

			tests := []struct {
				Name            string
				PreRun          func() error
				RequestPayload  backend.JoinReqPayload
				ExpectedPayload backend.JoinAnsPayload
			}{
				{
					Name: "valid join-request",
					RequestPayload: backend.JoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.JoinReq,
						},
						MACVersion: "1.0.2",
						PHYPayload: backend.HEXBytes(validJRPHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  &lorawan.CFList{868700000, 868900000},
					},
					ExpectedPayload: backend.JoinAnsPayload{
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
						PHYPayload: backend.HEXBytes(validJAPHYBytes),
						NwkSKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{223, 83, 195, 95, 48, 52, 204, 206, 208, 255, 53, 76, 112, 222, 4, 223},
						},
					},
				},
				{
					Name: "join-request with invalid mic",
					RequestPayload: backend.JoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.JoinReq,
						},
						MACVersion: "1.0.2",
						PHYPayload: backend.HEXBytes(invalidMICJRPHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  &lorawan.CFList{868700000, 868900000},
					},
					ExpectedPayload: backend.JoinAnsPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "0807060504030201",
							ReceiverID:      "010203",
							TransactionID:   1234,
							MessageType:     backend.JoinAns,
						},
						Result: backend.Result{
							ResultCode:  backend.MICFailed,
							Description: "invalid mic",
						},
					},
				},
				{
					Name: "join-request for unknown device",
					RequestPayload: backend.JoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.JoinReq,
						},
						MACVersion: "1.0.2",
						PHYPayload: backend.HEXBytes(validJRPHYBytes),
						DevEUI:     lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  &lorawan.CFList{868700000, 868900000},
					},
					ExpectedPayload: backend.JoinAnsPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "0807060504030201",
							ReceiverID:      "010203",
							TransactionID:   1234,
							MessageType:     backend.JoinAns,
						},
						Result: backend.Result{
							ResultCode:  backend.UnknownDevEUI,
							Description: "get device error: object does not exist",
						},
					},
				},
				{
					Name: "join-request for device without keys",
					PreRun: func() error {
						return storage.DeleteDeviceKeys(config.C.PostgreSQL.DB, d.DevEUI)
					},
					RequestPayload: backend.JoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.JoinReq,
						},
						MACVersion: "1.0.2",
						PHYPayload: backend.HEXBytes(validJRPHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  &lorawan.CFList{868700000, 868900000},
					},
					ExpectedPayload: backend.JoinAnsPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "0807060504030201",
							ReceiverID:      "010203",
							TransactionID:   1234,
							MessageType:     backend.JoinAns,
						},
						Result: backend.Result{
							ResultCode:  backend.UnknownDevEUI,
							Description: "get device-keys error: object does not exist",
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					if test.PreRun != nil {
						So(test.PreRun(), ShouldBeNil)
					}

					ans := HandleJoinRequest(test.RequestPayload)
					So(ans, ShouldResemble, test.ExpectedPayload)

					if ans.Result.ResultCode == backend.Success {
						_, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
						So(err, ShouldBeNil)
					}
				})
			}
		})
	})
}
