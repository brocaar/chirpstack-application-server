package join

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestJoin(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	Convey("Given a clean database with node", t, func() {
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

		config.C.JoinServer.KEK.ASKEKLabel = ""
		config.C.JoinServer.KEK.Set = nil

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
							AESKey: []byte{223, 83, 195, 95, 48, 52, 204, 206, 208, 255, 53, 76, 112, 222, 4, 223},
						},
						AppSKey: &backend.KeyEnvelope{
							AESKey: []byte{146, 123, 156, 145, 17, 131, 207, 254, 76, 178, 255, 75, 117, 84, 95, 109},
						},
					},
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
							AESKey: []byte{83, 127, 138, 174, 137, 108, 121, 224, 21, 209, 2, 208, 98, 134, 53, 78},
						},
						SNwkSIntKey: &backend.KeyEnvelope{
							AESKey: []byte{88, 148, 152, 153, 48, 146, 207, 219, 95, 210, 224, 42, 199, 81, 11, 241},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							AESKey: []byte{152, 152, 40, 60, 79, 102, 235, 108, 111, 213, 22, 88, 130, 4, 108, 64},
						},
						AppSKey: &backend.KeyEnvelope{
							AESKey: []byte{1, 98, 18, 21, 209, 202, 8, 254, 191, 12, 96, 44, 194, 173, 144, 250},
						},
					},
				},
				{
					PreRun: func() error {
						config.C.JoinServer.KEK.ASKEKLabel = "lora-app-server"
						config.C.JoinServer.KEK.Set = []struct {
							Label string `mapstructure:"label"`
							KEK   string `mapstructure:"kek"`
						}{
							{
								Label: "010203",
								KEK:   "00000000000000000000000000000000",
							},
							{
								Label: "lora-app-server",
								KEK:   "00000000000000000000000000000000",
							},
						}

						return nil
					},
					Name: "valid join-request (LoRaWAN 1.1) with KEK",
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
							KEKLabel: "010203",
							AESKey:   []byte{87, 85, 230, 195, 36, 30, 231, 230, 100, 111, 15, 254, 135, 120, 122, 0, 44, 249, 228, 176, 131, 73, 143, 0},
						},
						SNwkSIntKey: &backend.KeyEnvelope{
							KEKLabel: "010203",
							AESKey:   []byte{246, 176, 184, 31, 61, 48, 41, 18, 85, 145, 192, 176, 184, 141, 118, 201, 59, 72, 172, 164, 4, 22, 133, 211},
						},
						NwkSEncKey: &backend.KeyEnvelope{
							KEKLabel: "010203",
							AESKey:   []byte{78, 225, 236, 219, 189, 151, 82, 239, 109, 226, 140, 65, 233, 189, 174, 37, 39, 206, 241, 242, 2, 127, 157, 247},
						},
						AppSKey: &backend.KeyEnvelope{
							KEKLabel: "lora-app-server",
							AESKey:   []byte{248, 215, 201, 250, 55, 176, 209, 198, 53, 78, 109, 184, 225, 157, 157, 122, 180, 229, 199, 88, 30, 159, 30, 32},
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
							Description: "get device-keys error: object does not exist",
						},
					},
				},
				{
					Name: "join-request for device without keys",
					PreRun: func() error {
						return storage.DeleteDeviceKeys(storage.DB(), d.DevEUI)
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
					},
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
					},
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
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					if test.PreRun != nil {
						So(test.PreRun(), ShouldBeNil)
					}

					ans := HandleRejoinRequest(test.RequestPayload)
					So(ans, ShouldResemble, test.ExpectedPayload)
				})
			}
		})
	})
}
