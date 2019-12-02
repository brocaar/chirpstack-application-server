package as

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/lib/pq/hstore"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-api/go/as"
	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-api/go/common"
	gwPB "github.com/brocaar/chirpstack-api/go/gw"
	"github.com/brocaar/chirpstack-api/go/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/codec"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
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

		assert.Equal(pb.ErrorEvent{
			ApplicationId:   uint64(app.ID),
			ApplicationName: "test-app",
			DeviceName:      "test-node",
			DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Type:            pb.ErrorType_UPLINK_FCNT,
			Error:           "BOOM!",
			FCnt:            123,
			Tags: map[string]string{
				"foo": "bar",
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

			plJoin := <-h.SendJoinNotificationChan
			assert.Equal([]byte{1, 2, 3, 4}, plJoin.DevAddr)

			d, err := storage.GetDevice(context.Background(), storage.DB(), d.DevEUI, false, true)
			assert.NoError(err)
			assert.Equal(lorawan.DevAddr{0x01, 0x02, 0x03, 0x04}, d.DevAddr)
			assert.Equal(lorawan.AES128Key{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}, d.AppSKey)
		})

		t.Run("Activated device", func(t *testing.T) {
			assert := require.New(t)
			uplinkID, err := uuid.NewV4()
			assert.NoError(err)

			d.DevAddr = lorawan.DevAddr{}
			d.AppSKey = lorawan.AES128Key{}
			assert.NoError(storage.UpdateDevice(context.Background(), storage.DB(), &d, true))

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

				assert.Equal(pb.UplinkEvent{
					ApplicationId:   uint64(app.ID),
					ApplicationName: "test-app",
					DevEui:          d.DevEUI[:],
					DeviceName:      "test-node",
					RxInfo:          req.RxInfo,
					TxInfo:          req.TxInfo,
					Dr:              6,
					Adr:             true,
					FCnt:            10,
					FPort:           3,
					Data:            []byte{67, 216, 236, 205},
					ObjectJson:      "",
					Tags: map[string]string{
						"foo": "bar",
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
				assert.Equal(`{"fPort":3,"firstByte":67}`, pl.ObjectJson)
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
				assert.Equal(`{"fPort":4,"firstByte":68}`, pl.ObjectJson)
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
			StatusNotification     pb.StatusEvent
		}{
			{
				Name: "battery and margin",
				SetDeviceStatusRequest: as.SetDeviceStatusRequest{
					DevEui:       d.DevEUI[:],
					Margin:       10,
					Battery:      123,
					BatteryLevel: 25.50,
				},
				StatusNotification: pb.StatusEvent{
					ApplicationId:   uint64(app.ID),
					ApplicationName: app.Name,
					DeviceName:      d.Name,
					DevEui:          d.DevEUI[:],
					Margin:          10,
					BatteryLevel:    25.50,
					Tags: map[string]string{
						"foo": "bar",
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
				StatusNotification: pb.StatusEvent{
					ApplicationId:           uint64(app.ID),
					ApplicationName:         app.Name,
					DeviceName:              d.Name,
					DevEui:                  d.DevEUI[:],
					Margin:                  10,
					BatteryLevelUnavailable: true,
					Tags: map[string]string{
						"foo": "bar",
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
				StatusNotification: pb.StatusEvent{
					ApplicationId:       uint64(app.ID),
					ApplicationName:     app.Name,
					DeviceName:          d.Name,
					DevEui:              d.DevEUI[:],
					Margin:              10,
					ExternalPowerSource: true,
					Tags: map[string]string{
						"foo": "bar",
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

				assert.EqualValues(tst.StatusNotification.Margin, *d.DeviceStatusMargin)
				assert.EqualValues(tst.StatusNotification.ExternalPowerSource, d.DeviceStatusExternalPower)

				if tst.SetDeviceStatusRequest.BatteryLevelUnavailable || tst.SetDeviceStatusRequest.ExternalPowerSource {
					assert.Nil(d.DeviceStatusBattery)
				} else {
					assert.EqualValues(tst.StatusNotification.BatteryLevel, *d.DeviceStatusBattery)
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

		assert.Equal(pb.LocationEvent{
			ApplicationId:   uint64(app.ID),
			ApplicationName: app.Name,
			DeviceName:      d.Name,
			DevEui:          d.DevEUI[:],
			Location: &common.Location{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Source:    common.LocationSource_GEO_RESOLVER,
			},
			Tags: map[string]string{
				"foo": "bar",
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

		assert.Equal(pb.AckEvent{
			ApplicationId:   uint64(app.ID),
			ApplicationName: app.Name,
			DeviceName:      d.Name,
			DevEui:          d.DevEUI[:],
			Acknowledged:    true,
			FCnt:            10,
			Tags: map[string]string{
				"foo": "bar",
			},
		}, <-h.SendACKNotificationChan)
	})
}

func TestAS(t *testing.T) {
	suite.Run(t, new(ASTestSuite))
}
