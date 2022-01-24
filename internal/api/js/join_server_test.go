package js

import (
	"bytes"
	"context"
	"crypto/aes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/lib/pq/hstore"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

func TestJoinServerAPI(t *testing.T) {
	assert := require.New(t)
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	assert.NoError(storage.MigrateDown(storage.DB().DB))
	assert.NoError(storage.MigrateUp(storage.DB().DB))
	storage.RedisClient().FlushAll(context.Background())

	nsClient := nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(nsClient))

	h := mock.New()
	integration.SetMockIntegration(h)

	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	n := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test-ns:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &n))

	sp := storage.ServiceProfile{
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
		Name:            "test-sp",
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	dp := storage.DeviceProfile{
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
		Name:            "test-dp",
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &dp))
	dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
		Name:             "test-app",
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	d := storage.Device{
		ApplicationID:   app.ID,
		Name:            "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DeviceProfileID: dpID,
		Variables: hstore.Hstore{
			Map: map[string]sql.NullString{
				"home_netid": sql.NullString{
					String: "010203",
					Valid:  true,
				},
			},
		},
	}
	assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &d))

	dk := storage.DeviceKeys{
		DevEUI:    d.DevEUI,
		AppKey:    lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 9, 4, 3, 2, 1},
		NwkKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		JoinNonce: 65535,
	}
	assert.NoError(storage.CreateDeviceKeys(context.Background(), storage.DB(), &dk))

	api, err := getHandler(config.Config{})
	assert.NoError(err)
	server := httptest.NewServer(api)
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
	assert.NoError(err)

	t.Run("JoinReq LW10", func(t *testing.T) {
		assert := require.New(t)

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
		assert.NoError(jrPHY.SetUplinkJoinMIC(dk.NwkKey))
		jrPHYBytes, err := jrPHY.MarshalBinary()
		assert.NoError(err)

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
		assert.NoError(jaPHY.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, dk.NwkKey))
		assert.NoError(jaPHY.EncryptJoinAcceptPayload(dk.NwkKey))
		jaPHYBytes, err := jaPHY.MarshalBinary()
		assert.NoError(err)
		assert.Equal(jaPHYBytes, []byte{32, 38, 244, 178, 71, 240, 165, 215, 228, 106, 114, 14, 97, 200, 188, 203, 197, 23, 159, 69, 102, 225, 133, 237, 104, 137, 88, 155, 177, 169, 198, 140, 192})

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
		assert.NoError(err)

		req, err := http.NewRequest("POST", server.URL, bytes.NewReader(joinReqPayloadJSON))
		assert.NoError(err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		t.Run("Response", func(t *testing.T) {
			assert := require.New(t)

			var joinAnsPayload backend.JoinAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&joinAnsPayload))
			assert.Equal(joinAnsPayload, backend.JoinAnsPayload{
				BasePayloadResult: backend.BasePayloadResult{
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

		t.Run("Keys", func(t *testing.T) {
			assert := require.New(t)

			dk, err := storage.GetDeviceKeys(context.Background(), storage.DB(), d.DevEUI)
			assert.NoError(err)

			assert.Equal(dk.JoinNonce, 65536)
		})
	})

	assert.NoError(storage.UpdateDeviceKeys(context.Background(), storage.DB(), &dk))

	t.Run("JoinReq LW11", func(t *testing.T) {
		assert := require.New(t)

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
		assert.NoError(jrPHY.SetUplinkJoinMIC(dk.NwkKey))
		jrPHYBytes, err := jrPHY.MarshalBinary()
		assert.NoError(err)

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
					OptNeg:      true,
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
		assert.NoError(err)

		assert.NoError(jaPHY.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, jsIntKey))
		assert.NoError(jaPHY.EncryptJoinAcceptPayload(dk.NwkKey))
		jaPHYBytes, err := jaPHY.MarshalBinary()
		assert.NoError(err)
		assert.Equal(jaPHYBytes, []byte{0x20, 0xa4, 0xc8, 0x99, 0x79, 0x5a, 0x42, 0xf5, 0xd1, 0x60, 0x67, 0xa8, 0x22, 0xc2, 0x1b, 0x37, 0x9a, 0xf8, 0xaa, 0xa3, 0x1, 0xa9, 0xd, 0x1c, 0x60, 0xd4, 0xcb, 0x79, 0x2c, 0xb8, 0xee, 0xb0, 0xa3})

		joinReqPayload := backend.JoinReqPayload{
			BasePayload: backend.BasePayload{
				ProtocolVersion: backend.ProtocolVersion1_0,
				SenderID:        "010203",
				ReceiverID:      "0807060504030201",
				TransactionID:   1234,
				MessageType:     backend.JoinReq,
			},
			MACVersion: "1.1.0",
			PHYPayload: backend.HEXBytes(jrPHYBytes),
			DevEUI:     d.DevEUI,
			DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				RX2DataRate: 5,
				RX1DROffset: 1,
				OptNeg:      true,
			},
			RxDelay: 1,
			CFList:  backend.HEXBytes(cFListB),
		}
		joinReqPayloadJSON, err := json.Marshal(joinReqPayload)
		assert.NoError(err)

		req, err := http.NewRequest("POST", server.URL, bytes.NewReader(joinReqPayloadJSON))
		assert.NoError(err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		t.Run("Response", func(t *testing.T) {
			assert := require.New(t)

			var joinAnsPayload backend.JoinAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&joinAnsPayload))
			assert.Equal(backend.JoinAnsPayload{
				BasePayloadResult: backend.BasePayloadResult{
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
				},
				PHYPayload: backend.HEXBytes(jaPHYBytes),
				SNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{0x58, 0x94, 0x98, 0x99, 0x30, 0x92, 0xcf, 0xdb, 0x5f, 0xd2, 0xe0, 0x2a, 0xc7, 0x51, 0x0b, 0xf1},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{0x1c, 0xee, 0xb6, 0x88, 0x07, 0xa9, 0x15, 0xa4, 0xac, 0xd5, 0x18, 0xb3, 0x8b, 0x20, 0xce, 0x03},
				},
				FNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{0x53, 0x7f, 0x8a, 0xae, 0x89, 0x6c, 0x79, 0xe0, 0x15, 0xd1, 0x02, 0xd0, 0x62, 0x86, 0x35, 0x4e},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					AESKey: []byte{0x98, 0x98, 0x28, 0x3c, 0x4f, 0x66, 0xeb, 0x6c, 0x6f, 0xd5, 0x16, 0x58, 0x82, 0x04, 0x6c, 0x40},
				},
			}, joinAnsPayload)
		})

		t.Run("Keys", func(t *testing.T) {
			assert := require.New(t)

			dk, err := storage.GetDeviceKeys(context.Background(), storage.DB(), d.DevEUI)
			assert.NoError(err)

			assert.Equal(dk.JoinNonce, 65536)
		})
	})

	assert.NoError(storage.UpdateDeviceKeys(context.Background(), storage.DB(), &dk))

	t.Run("RejoinReq", func(t *testing.T) {
		assert := require.New(t)

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
		assert.NoError(err)

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
		assert.NoError(err)

		req, err := http.NewRequest("POST", server.URL, bytes.NewReader(rejoinReqPayloadJSON))
		assert.NoError(err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		t.Run("Response", func(t *testing.T) {
			assert := require.New(t)

			var rejoinAnsPayload backend.RejoinAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&rejoinAnsPayload))
			assert.Equal(rejoinAnsPayload, backend.RejoinAnsPayload{
				BasePayloadResult: backend.BasePayloadResult{
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

	t.Run("HomeNSReq", func(t *testing.T) {
		t.Run("Known DevEUI", func(t *testing.T) {
			assert := require.New(t)

			homeNSReq := backend.HomeNSReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.HomeNSReq,
				},
				DevEUI: d.DevEUI,
			}
			homeNSReqJSON, err := json.Marshal(homeNSReq)
			assert.NoError(err)

			req, err := http.NewRequest("POST", server.URL, bytes.NewReader(homeNSReqJSON))
			assert.NoError(err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(err)

			var homeNSAns backend.HomeNSAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&homeNSAns))
			assert.Equal(backend.Success, homeNSAns.Result.ResultCode)
			assert.Equal(lorawan.NetID{1, 2, 3}, homeNSAns.HNetID)
		})

		t.Run("Know DevEUI, no home_netid variable", func(t *testing.T) {
			assert := require.New(t)

			d.Variables = hstore.Hstore{}
			assert.NoError(storage.UpdateDevice(context.Background(), storage.DB(), &d, true))

			homeNSReq := backend.HomeNSReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.HomeNSReq,
				},
				DevEUI: d.DevEUI,
			}
			homeNSReqJSON, err := json.Marshal(homeNSReq)
			assert.NoError(err)

			req, err := http.NewRequest("POST", server.URL, bytes.NewReader(homeNSReqJSON))
			assert.NoError(err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(err)

			var homeNSAns backend.HomeNSAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&homeNSAns))
			assert.Equal(backend.Success, homeNSAns.Result.ResultCode)
			assert.Equal(lorawan.NetID{0, 0, 0}, homeNSAns.HNetID)
		})

		t.Run("Unknown DevEUI", func(t *testing.T) {
			assert := require.New(t)

			homeNSReq := backend.HomeNSReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.HomeNSReq,
				},
				DevEUI: lorawan.EUI64{3, 3, 3, 3, 3, 3, 3, 3},
			}
			homeNSReqJSON, err := json.Marshal(homeNSReq)
			assert.NoError(err)

			req, err := http.NewRequest("POST", server.URL, bytes.NewReader(homeNSReqJSON))
			assert.NoError(err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(err)

			var homeNSAns backend.HomeNSAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&homeNSAns))
			assert.Equal(backend.UnknownDevEUI, homeNSAns.Result.ResultCode)
		})
	})
}

func getJSIntKey(nwkKey lorawan.AES128Key, devEUI lorawan.EUI64) (lorawan.AES128Key, error) {
	return getJSKey(0x06, devEUI, nwkKey)
}

func getJSKey(typ byte, devEUI lorawan.EUI64, nwkKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)

	b[0] = typ

	devB, err := devEUI.MarshalBinary()
	if err != nil {
		return key, err
	}
	copy(b[1:9], devB[:])

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)
	return key, nil
}
