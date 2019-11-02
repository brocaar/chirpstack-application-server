package as

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/lib/pq/hstore"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/common"
	gwPB "github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/brocaar/lorawan"
)

type ASTestSuite struct {
	suite.Suite
}

func (ts *ASTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)
	test.MustFlushRedis(storage.RedisPool())
}

func (ts *ASTestSuite) TestApplicationServer() {
	assert := require.New(ts.T())

	nsClient := nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(nsClient))

	nsClient.GetDeviceProfileResponse = ns.GetDeviceProfileResponse{
		DeviceProfile: &ns.DeviceProfile{
			SupportsJoin: true,
		},
	}

	org := storage.Organization{
		Name: "test-as-org",
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

	app := storage.Application{
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
		Name:             "test-app",
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	d := storage.Device{
		ApplicationID:   app.ID,
		Name:            "test-node",
		DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		DeviceProfileID: dpID,
		Tags: hstore.Hstore{
			Map: map[string]sql.NullString{
				"foo": sql.NullString{String: "bar", Valid: true},
			},
		},
		Variables: hstore.Hstore{
			Map: map[string]sql.NullString{
				"secret_token": sql.NullString{String: "secret value", Valid: true},
			},
		},
	}
	assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &d))

	dc := storage.DeviceKeys{
		DevEUI: d.DevEUI,
		NwkKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(storage.CreateDeviceKeys(context.Background(), storage.DB(), &dc))

	gw := storage.Gateway{
		MAC:             lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Name:            "test-gw",
		Description:     "test gateway",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gw))

	h := mock.New()
	integration.SetIntegration(h)

	ctx := context.Background()
	api := NewApplicationServerAPI()

	ts.T().Run("HandleError", func(t *testing.T) {
		assert := require.New(t)

		_, err := api.HandleError(ctx, &as.HandleErrorRequest{
			DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Type:   as.ErrorType_DATA_UP_FCNT,
			Error:  "BOOM!",
			FCnt:   123,
		})
		assert.NoError(err)

		assert.Equal(integration.ErrorNotification{
			ApplicationID:   app.ID,
			ApplicationName: "test-app",
			DeviceName:      "test-node",
			DevEUI:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Type:            "DATA_UP_FCNT",
			Error:           "BOOM!",
			FCnt:            123,
			Tags: map[string]string{
				"foo": "bar",
			},
			Variables: map[string]string{
				"secret_token": "secret value",
			},
		}, <-h.SendErrorNotificationChan)
	})

	ts.T().Run("HandleUplinkDataRequest", func(t *testing.T) {
		t.Run("With DeviceSecurityContext", func(t *testing.T) {
			assert := require.New(t)

			now := time.Now().UTC()
			uplinkID, err := uuid.NewV4()
			assert.NoError(err)

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
						UplinkId:  uplinkID[:],
						Rssi:      -60,
						LoraSnr:   5,
						Location: &common.Location{
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

			_, err = api.HandleUplinkData(ctx, &req)
			assert.NoError(err)

			plData := <-h.SendDataUpChan
			assert.Equal([]byte{33, 99, 53, 13}, plData.Data)
			assert.Equal(uplinkID, plData.RXInfo[0].UplinkID)

			plJoin := <-h.SendJoinNotificationChan
			assert.Equal(lorawan.DevAddr{1, 2, 3, 4}, plJoin.DevAddr)
			assert.Equal(uplinkID, plData.RXInfo[0].UplinkID)
		})

		t.Run("Activated device", func(t *testing.T) {
			assert := require.New(t)
			uplinkID, err := uuid.NewV4()
			assert.NoError(err)

			da := storage.DeviceActivation{
				DevEUI:  d.DevEUI,
				DevAddr: lorawan.DevAddr{},
				AppSKey: lorawan.AES128Key{},
			}
			assert.NoError(storage.CreateDeviceActivation(context.Background(), storage.DB(), &da))

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
						UplinkId:  uplinkID[:],
						Rssi:      -60,
						LoraSnr:   5,
						Location: &common.Location{
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

			t.Run("No codec", func(t *testing.T) {
				assert := require.New(t)

				_, err := api.HandleUplinkData(ctx, &req)
				assert.NoError(err)

				d, err := storage.GetDevice(context.Background(), storage.DB(), d.DevEUI, false, false)
				assert.NoError(err)
				assert.InDelta(time.Now().UnixNano(), d.LastSeenAt.UnixNano(), float64(time.Second))

				assert.Equal(integration.DataUpPayload{
					ApplicationID:   app.ID,
					ApplicationName: "test-app",
					DeviceName:      "test-node",
					DevEUI:          d.DevEUI,
					RXInfo: []integration.RXInfo{
						{
							GatewayID: mac,
							UplinkID:  uplinkID,
							Name:      "test-gw",
							Location: &integration.Location{
								Latitude:  52.3740364,
								Longitude: 4.9144401,
								Altitude:  10,
							},
							Time:    &now,
							RSSI:    -60,
							LoRaSNR: 5,
						},
					},
					TXInfo: integration.TXInfo{
						Frequency: 868100000,
						DR:        6,
					},
					ADR:   true,
					FCnt:  10,
					FPort: 3,
					Data:  []byte{67, 216, 236, 205},
					Tags: map[string]string{
						"foo": "bar",
					},
					Variables: map[string]string{
						"secret_token": "secret value",
					},
				}, <-h.SendDataUpChan)
			})

			t.Run("JS codec", func(t *testing.T) {
				assert := require.New(t)

				app.PayloadCodec = codec.CustomJSType
				app.PayloadDecoderScript = `
					function Decode(fPort, bytes) {
						return {
							"fPort": fPort,
							"firstByte": bytes[0]
						}
					}
				`
				assert.NoError(storage.UpdateApplication(context.Background(), storage.DB(), app))

				_, err := api.HandleUplinkData(ctx, &req)
				assert.NoError(err)

				pl := <-h.SendDataUpChan
				assert.NotNil(pl.Object)
				b, err := json.Marshal(pl.Object)
				assert.NoError(err)
				assert.Equal(`{"fPort":3,"firstByte":67}`, string(b))
			})

			t.Run("JS codec on device-profile", func(t *testing.T) {
				assert := require.New(t)

				dp.PayloadCodec = codec.CustomJSType
				dp.PayloadDecoderScript = `
					function Decode(fPort, bytes) {
						return {
							"fPort": fPort + 1,
							"firstByte": bytes[0] + 1
						}
					}
				`
				assert.NoError(storage.UpdateDeviceProfile(context.Background(), storage.DB(), &dp))

				_, err := api.HandleUplinkData(ctx, &req)
				assert.NoError(err)

				pl := <-h.SendDataUpChan
				assert.NotNil(pl.Object)
				b, err := json.Marshal(pl.Object)
				assert.NoError(err)
				assert.Equal(`{"fPort":4,"firstByte":68}`, string(b))
			})
		})
	})

	ts.T().Run("HandleGatewayStats", func(t *testing.T) {
		assert := require.New(t)

		now := time.Now()
		nowPB, err := ptypes.TimestampProto(now)
		assert.NoError(err)

		assert.NoError(storage.SetAggregationIntervals([]storage.AggregationInterval{storage.AggregationMinute}))
		storage.SetMetricsTTL(time.Minute, time.Minute, time.Minute, time.Minute)

		stats := as.HandleGatewayStatsRequest{
			GatewayId: gw.MAC[:],
			Time:      nowPB,
			Location: &common.Location{
				Latitude:  1.1234,
				Longitude: 2.1234,
				Altitude:  3,
			},
			RxPacketsReceived:   10,
			RxPacketsReceivedOk: 9,
			TxPacketsReceived:   8,
			TxPacketsEmitted:    7,
		}
		_, err = api.HandleGatewayStats(ctx, &stats)
		assert.NoError(err)

		start := time.Now().Truncate(time.Minute)
		end := time.Now()

		metrics, err := storage.GetMetrics(context.Background(), storage.RedisPool(), storage.AggregationMinute, "gw:"+gw.MAC.String(), start, end)
		assert.NoError(err)
		assert.Len(metrics, 1)

		assert.Equal(map[string]float64{
			"rx_count":    10,
			"rx_ok_count": 9,
			"tx_count":    8,
			"tx_ok_count": 7,
		}, metrics[0].Metrics)
		assert.Equal(start.UTC(), metrics[0].Time.UTC())
	})

	ts.T().Run("SetDeviceStatus", func(t *testing.T) {
		tests := []struct {
			Name                   string
			SetDeviceStatusRequest as.SetDeviceStatusRequest
			StatusNotification     integration.StatusNotification
		}{
			{
				Name: "battery and margin",
				SetDeviceStatusRequest: as.SetDeviceStatusRequest{
					DevEui:       d.DevEUI[:],
					Margin:       10,
					Battery:      123,
					BatteryLevel: 25.50,
				},
				StatusNotification: integration.StatusNotification{
					ApplicationID:   app.ID,
					ApplicationName: app.Name,
					DeviceName:      d.Name,
					DevEUI:          d.DevEUI,
					Margin:          10,
					Battery:         123,
					BatteryLevel:    25.50,
					Tags: map[string]string{
						"foo": "bar",
					},
					Variables: map[string]string{
						"secret_token": "secret value",
					},
				},
			},
			{
				Name: "battery unavailable and margin",
				SetDeviceStatusRequest: as.SetDeviceStatusRequest{
					DevEui:                  d.DevEUI[:],
					Margin:                  10,
					BatteryLevelUnavailable: true,
				},
				StatusNotification: integration.StatusNotification{
					ApplicationID:           app.ID,
					ApplicationName:         app.Name,
					DeviceName:              d.Name,
					DevEUI:                  d.DevEUI,
					Margin:                  10,
					BatteryLevelUnavailable: true,
					Tags: map[string]string{
						"foo": "bar",
					},
					Variables: map[string]string{
						"secret_token": "secret value",
					},
				},
			},
			{
				Name: "external power and margin",
				SetDeviceStatusRequest: as.SetDeviceStatusRequest{
					DevEui:              d.DevEUI[:],
					Margin:              10,
					ExternalPowerSource: true,
				},
				StatusNotification: integration.StatusNotification{
					ApplicationID:       app.ID,
					ApplicationName:     app.Name,
					DeviceName:          d.Name,
					DevEUI:              d.DevEUI,
					Margin:              10,
					ExternalPowerSource: true,
					Tags: map[string]string{
						"foo": "bar",
					},
					Variables: map[string]string{
						"secret_token": "secret value",
					},
				},
			},
		}

		for _, tst := range tests {
			t.Run(tst.Name, func(t *testing.T) {
				assert := require.New(t)

				_, err := api.SetDeviceStatus(ctx, &tst.SetDeviceStatusRequest)
				assert.NoError(err)
				assert.Equal(tst.StatusNotification, <-h.SendStatusNotificationChan)

				d, err := storage.GetDevice(context.Background(), storage.DB(), d.DevEUI, false, false)
				assert.NoError(err)

				assert.Equal(tst.StatusNotification.Margin, *d.DeviceStatusMargin)
				assert.Equal(tst.StatusNotification.ExternalPowerSource, d.DeviceStatusExternalPower)

				if tst.SetDeviceStatusRequest.BatteryLevelUnavailable || tst.SetDeviceStatusRequest.ExternalPowerSource {
					assert.Nil(d.DeviceStatusBattery)
				} else {
					assert.Equal(tst.StatusNotification.BatteryLevel, *d.DeviceStatusBattery)
				}
			})
		}
	})

	ts.T().Run("SetDeviceLocation", func(t *testing.T) {
		assert := require.New(t)

		_, err := api.SetDeviceLocation(ctx, &as.SetDeviceLocationRequest{
			DevEui: d.DevEUI[:],
			Location: &common.Location{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Source:    common.LocationSource_GEO_RESOLVER,
			},
		})
		assert.NoError(err)

		assert.Equal(integration.LocationNotification{
			ApplicationID:   app.ID,
			ApplicationName: app.Name,
			DeviceName:      d.Name,
			DevEUI:          d.DevEUI,
			Location: integration.Location{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
			},
			Tags: map[string]string{
				"foo": "bar",
			},
			Variables: map[string]string{
				"secret_token": "secret value",
			},
		}, <-h.SendLocationNotificationChan)

		d, err := storage.GetDevice(context.Background(), storage.DB(), d.DevEUI, false, true)
		assert.NoError(err)
		assert.Equal(1.123, *d.Latitude)
		assert.Equal(2.123, *d.Longitude)
		assert.Equal(3.123, *d.Altitude)
	})

	ts.T().Run("HandleDownlinkACK", func(t *testing.T) {
		_, err := api.HandleDownlinkACK(ctx, &as.HandleDownlinkACKRequest{
			DevEui:       d.DevEUI[:],
			FCnt:         10,
			Acknowledged: true,
		})
		assert.NoError(err)

		assert.Equal(integration.ACKNotification{
			ApplicationID:   app.ID,
			ApplicationName: app.Name,
			DeviceName:      d.Name,
			DevEUI:          d.DevEUI,
			Acknowledged:    true,
			FCnt:            10,
			Tags: map[string]string{
				"foo": "bar",
			},
			Variables: map[string]string{
				"secret_token": "secret value",
			},
		}, <-h.SendACKNotificationChan)
	})
}

func TestAS(t *testing.T) {
	suite.Run(t, new(ASTestSuite))
}
