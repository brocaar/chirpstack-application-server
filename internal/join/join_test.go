package join

import (
	"fmt"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/test/testhandler"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
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

	p := storage.NewRedisPool(conf.RedisURL)

	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = p

	Convey("Given a clean database with node", t, func() {
		test.MustResetDB(config.C.PostgreSQL.DB)
		test.MustFlushRedis(p)

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
			Name:            "test-device",
			DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceProfileID: dpID,
		}
		So(storage.CreateDevice(config.C.PostgreSQL.DB, &d), ShouldBeNil)

		dk := storage.DeviceKeys{
			DevEUI:    d.DevEUI,
			NwkKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			JoinNonce: 65535,
		}
		So(storage.CreateDeviceKeys(config.C.PostgreSQL.DB, &dk), ShouldBeNil)

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

		Convey("Given a set of tests for join-request", func() {
			validJRPHY := lorawan.PHYPayload{
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
			So(validJRPHY.SetUplinkJoinMIC(dk.NwkKey), ShouldBeNil)
			validJRPHYBytes, err := validJRPHY.MarshalBinary()
			So(err, ShouldBeNil)

			invalidMICJRPHY := lorawan.PHYPayload{
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
			So(invalidMICJRPHY.SetUplinkJoinMIC(lorawan.AES128Key{}), ShouldBeNil)
			invalidMICJRPHYBytes, err := invalidMICJRPHY.MarshalBinary()
			So(err, ShouldBeNil)

			validJAPHY := lorawan.PHYPayload{
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
			So(validJAPHY.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, dk.NwkKey), ShouldBeNil)
			So(validJAPHY.EncryptJoinAcceptPayload(dk.NwkKey), ShouldBeNil)
			validJAPHYBytes, err := validJAPHY.MarshalBinary()
			So(err, ShouldBeNil)

			validJAPHYLW11 := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinAccept,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinAcceptPayload{
					JoinNonce: 65536,
					HomeNetID: lorawan.NetID{1, 2, 3},
					DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						OptNeg:      true,
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
			jsIntKey, err := getJSIntKey(dk.NwkKey, d.DevEUI)
			So(err, ShouldBeNil)
			So(validJAPHYLW11.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, jsIntKey), ShouldBeNil)
			So(validJAPHYLW11.EncryptJoinAcceptPayload(dk.NwkKey), ShouldBeNil)
			validJAPHYLW11Bytes, err := validJAPHYLW11.MarshalBinary()
			So(err, ShouldBeNil)

			tests := []struct {
				Name            string
				PreRun          func() error
				RequestPayload  backend.JoinReqPayload
				ExpectedPayload backend.JoinAnsPayload
				ExpectedAppSKey lorawan.AES128Key
			}{
				{
					Name: "valid join-request (LoRaWAN 1.0)",
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
						CFList:  backend.HEXBytes(cFListB),
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
					ExpectedAppSKey: lorawan.AES128Key{146, 123, 156, 145, 17, 131, 207, 254, 76, 178, 255, 75, 117, 84, 95, 109},
				},
				{
					Name: "valid join-request (LoRaWAN 1.1)",
					RequestPayload: backend.JoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.JoinReq,
						},
						MACVersion: "1.1.0",
						PHYPayload: backend.HEXBytes(validJRPHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							OptNeg:      true,
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  backend.HEXBytes(cFListB),
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
						PHYPayload: backend.HEXBytes(validJAPHYLW11Bytes),
						FNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{83, 127, 138, 174, 137, 108, 121, 224, 21, 209, 2, 208, 98, 134, 53, 78},
						},
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{88, 148, 152, 153, 48, 146, 207, 219, 95, 210, 224, 42, 199, 81, 11, 241},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{152, 152, 40, 60, 79, 102, 235, 108, 111, 213, 22, 88, 130, 4, 108, 64},
						},
					},
					ExpectedAppSKey: lorawan.AES128Key{1, 98, 18, 21, 209, 202, 8, 254, 191, 12, 96, 44, 194, 173, 144, 250},
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
						CFList:  backend.HEXBytes(cFListB),
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
						CFList:  backend.HEXBytes(cFListB),
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
						CFList:  backend.HEXBytes(cFListB),
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
						da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
						So(err, ShouldBeNil)

						So(da.AppSKey, ShouldEqual, test.ExpectedAppSKey)
					}
				})
			}
		})

		Convey("Given a set of tests for rejoin-request", func() {
			jsIntKey, err := getJSIntKey(dk.NwkKey, d.DevEUI)
			So(err, ShouldBeNil)
			jsEncKey, err := getJSEncKey(dk.NwkKey, d.DevEUI)
			So(err, ShouldBeNil)

			rj0PHY := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.RejoinRequest,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.RejoinRequestType02Payload{
					RejoinType: lorawan.RejoinRequestType0,
					NetID:      lorawan.NetID{1, 2, 3},
					DevEUI:     d.DevEUI,
					RJCount0:   123,
				},
			}
			// no need to set the MIC as it is not validated by the js
			rj0PHYBytes, err := rj0PHY.MarshalBinary()
			So(err, ShouldBeNil)

			rj1PHY := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.RejoinRequest,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.RejoinRequestType1Payload{
					RejoinType: lorawan.RejoinRequestType1,
					JoinEUI:    lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
					DevEUI:     d.DevEUI,
					RJCount1:   123,
				},
			}
			So(rj1PHY.SetUplinkJoinMIC(jsIntKey), ShouldBeNil)
			rj1PHYBytes, err := rj1PHY.MarshalBinary()
			So(err, ShouldBeNil)

			rj2PHY := lorawan.PHYPayload{
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
			// no need to set the MIC as it is not validated by the js
			rj2PHYBytes, err := rj2PHY.MarshalBinary()
			So(err, ShouldBeNil)

			ja0PHY := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinAccept,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinAcceptPayload{
					JoinNonce: lorawan.JoinNonce(dk.JoinNonce + 1),
					HomeNetID: lorawan.NetID{1, 2, 3},
					DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						OptNeg:      true,
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
			So(ja0PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType0, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey), ShouldBeNil)
			So(ja0PHY.EncryptJoinAcceptPayload(jsEncKey), ShouldBeNil)
			ja0PHYBytes, err := ja0PHY.MarshalBinary()
			So(err, ShouldBeNil)

			ja1PHY := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinAccept,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinAcceptPayload{
					JoinNonce: lorawan.JoinNonce(dk.JoinNonce + 1),
					HomeNetID: lorawan.NetID{1, 2, 3},
					DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						OptNeg:      true,
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
			So(ja1PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType1, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey), ShouldBeNil)
			So(ja1PHY.EncryptJoinAcceptPayload(jsEncKey), ShouldBeNil)
			ja1PHYBytes, err := ja1PHY.MarshalBinary()
			So(err, ShouldBeNil)

			ja2PHY := lorawan.PHYPayload{
				MHDR: lorawan.MHDR{
					MType: lorawan.JoinAccept,
					Major: lorawan.LoRaWANR1,
				},
				MACPayload: &lorawan.JoinAcceptPayload{
					JoinNonce: lorawan.JoinNonce(dk.JoinNonce + 1),
					HomeNetID: lorawan.NetID{1, 2, 3},
					DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
					DLSettings: lorawan.DLSettings{
						OptNeg:      true,
						RX2DataRate: 5,
						RX1DROffset: 1,
					},
					RXDelay: 1,
				},
			}
			So(ja2PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType2, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey), ShouldBeNil)
			So(ja2PHY.EncryptJoinAcceptPayload(jsEncKey), ShouldBeNil)
			ja2PHYBytes, err := ja2PHY.MarshalBinary()
			So(err, ShouldBeNil)

			tests := []struct {
				Name            string
				PreRun          func() error
				RequestPayload  backend.RejoinReqPayload
				ExpectedPayload backend.RejoinAnsPayload
				ExpectedAppSKey lorawan.AES128Key
			}{
				{
					Name: "valid rejoin-request type 0",
					RequestPayload: backend.RejoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.RejoinReq,
						},
						MACVersion: "1.1.0",
						PHYPayload: backend.HEXBytes(rj0PHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							OptNeg:      true,
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  backend.HEXBytes(cFListB),
					},
					ExpectedPayload: backend.RejoinAnsPayload{
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
						PHYPayload: backend.HEXBytes(ja0PHYBytes),
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
						},
						FNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
						},
					},
					ExpectedAppSKey: lorawan.AES128Key{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
				{
					Name: "valid rejoin-request type 1",
					RequestPayload: backend.RejoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.RejoinReq,
						},
						MACVersion: "1.1.0",
						PHYPayload: backend.HEXBytes(rj1PHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							OptNeg:      true,
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
						CFList:  backend.HEXBytes(cFListB),
					},
					ExpectedPayload: backend.RejoinAnsPayload{
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
						PHYPayload: backend.HEXBytes(ja1PHYBytes),
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
						},
						FNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
						},
					},
					ExpectedAppSKey: lorawan.AES128Key{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
				{
					Name: "valid rejoin-request type 2",
					RequestPayload: backend.RejoinReqPayload{
						BasePayload: backend.BasePayload{
							ProtocolVersion: backend.ProtocolVersion1_0,
							SenderID:        "010203",
							ReceiverID:      "0807060504030201",
							TransactionID:   1234,
							MessageType:     backend.RejoinReq,
						},
						MACVersion: "1.1.0",
						PHYPayload: backend.HEXBytes(rj2PHYBytes),
						DevEUI:     d.DevEUI,
						DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
						DLSettings: lorawan.DLSettings{
							OptNeg:      true,
							RX2DataRate: 5,
							RX1DROffset: 1,
						},
						RxDelay: 1,
					},
					ExpectedPayload: backend.RejoinAnsPayload{
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
						PHYPayload: backend.HEXBytes(ja2PHYBytes),
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
						},
						FNwkSIntKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: lorawan.AES128Key{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
						},
					},
					ExpectedAppSKey: lorawan.AES128Key{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					if test.PreRun != nil {
						So(test.PreRun(), ShouldBeNil)
					}

					ans := HandleRejoinRequest(test.RequestPayload)
					So(ans, ShouldResemble, test.ExpectedPayload)

					if ans.Result.ResultCode == backend.Success {
						da, err := storage.GetLastDeviceActivationForDevEUI(config.C.PostgreSQL.DB, d.DevEUI)
						So(err, ShouldBeNil)

						So(da.AppSKey[:], ShouldResemble, test.ExpectedAppSKey[:])
					}
				})
			}
		})
	})
}
