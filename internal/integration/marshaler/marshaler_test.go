package marshaler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type MarshalerTestSuite struct {
	suite.Suite
}

func (ts *MarshalerTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	storage.RedisClient().FlushAll(context.Background())
	assert.NoError(storage.MigrateDown(storage.DB().DB))
	assert.NoError(storage.MigrateUp(storage.DB().DB))
}

func (ts *MarshalerTestSuite) GetUplinkEvent() integration.UplinkEvent {
	assert := require.New(ts.T())

	now := time.Now().UTC()
	nowPB, err := ptypes.TimestampProto(now)
	assert.NoError(err)

	tenSeconds := time.Second * 10
	tenSecondsPB := ptypes.DurationProto(tenSeconds)

	return integration.UplinkEvent{
		ApplicationId:     123,
		ApplicationName:   "test-application",
		DeviceName:        "test-device",
		DeviceProfileName: "test-profile",
		DeviceProfileId:   "f293e453-6d9c-4a22-8c4d-99b2dbe4e94f",
		DevEui:            []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				Time:              nowPB,
				TimeSinceGpsEpoch: tenSecondsPB,
				Rssi:              110,
				LoraSnr:           5.6,
				Channel:           10,
				RfChain:           1,
				Board:             0,
				Antenna:           0,
				Location: &common.Location{
					Latitude:  1.1234,
					Longitude: 2.1234,
					Altitude:  3.1,
				},
			},
		},
		TxInfo: &gw.UplinkTXInfo{
			Frequency:  868100000,
			Modulation: common.Modulation_LORA,
			ModulationInfo: &gw.UplinkTXInfo_LoraModulationInfo{
				LoraModulationInfo: &gw.LoRaModulationInfo{
					Bandwidth:       125,
					SpreadingFactor: 12,
					CodeRate:        "3/4",
				},
			},
		},
		Adr:        true,
		Dr:         3,
		FCnt:       123,
		FPort:      101,
		Data:       []byte{0x01, 0x02, 0x03, 0x04},
		ObjectJson: `{"foo":"bar"}`,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetJoinEvent() integration.JoinEvent {
	assert := require.New(ts.T())

	now := time.Now().UTC()
	nowPB, err := ptypes.TimestampProto(now)
	assert.NoError(err)

	tenSeconds := time.Second * 10
	tenSecondsPB := ptypes.DurationProto(tenSeconds)

	return integration.JoinEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		DevAddr:         []byte{0x01, 0x02, 0x03, 0x4},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				Time:              nowPB,
				TimeSinceGpsEpoch: tenSecondsPB,
				Rssi:              110,
				LoraSnr:           5.6,
				Channel:           10,
				RfChain:           1,
				Board:             0,
				Antenna:           0,
				Location: &common.Location{
					Latitude:  1.1234,
					Longitude: 2.1234,
					Altitude:  3.1,
				},
			},
		},
		TxInfo: &gw.UplinkTXInfo{
			Frequency:  868100000,
			Modulation: common.Modulation_LORA,
			ModulationInfo: &gw.UplinkTXInfo_LoraModulationInfo{
				LoraModulationInfo: &gw.LoRaModulationInfo{
					Bandwidth:       125,
					SpreadingFactor: 12,
					CodeRate:        "3/4",
				},
			},
		},
		Dr: 3,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetErrorEvent() integration.ErrorEvent {
	return integration.ErrorEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Type:            integration.ErrorType_UPLINK_CODEC,
		Error:           "function error",
		FCnt:            110,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetAckEvent() integration.AckEvent {
	return integration.AckEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Acknowledged:    true,
		FCnt:            123,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetStatusEvent() integration.StatusEvent {
	return integration.StatusEvent{
		ApplicationId:           123,
		ApplicationName:         "test-application",
		DeviceName:              "test-device",
		DevEui:                  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Margin:                  10,
		ExternalPowerSource:     true,
		BatteryLevelUnavailable: true,
		BatteryLevel:            55,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetLocationEvent() integration.LocationEvent {
	return integration.LocationEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  100,
			Source:    common.LocationSource_GEO_RESOLVER_TDOA,
			Accuracy:  10,
		},
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) GetTxAckEvent() integration.TxAckEvent {
	return integration.TxAckEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		FCnt:            123,
		Tags: map[string]string{
			"test": "tag",
		},
		GatewayId: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		TxInfo: &gw.DownlinkTXInfo{
			Frequency:  868100000,
			Board:      0,
			Antenna:    0,
			Modulation: common.Modulation_LORA,
			ModulationInfo: &gw.DownlinkTXInfo_LoraModulationInfo{
				LoraModulationInfo: &gw.LoRaModulationInfo{
					Bandwidth:       125,
					SpreadingFactor: 12,
					CodeRate:        "3/4",
				},
			},
		},
	}
}

func (ts *MarshalerTestSuite) GetIntegrationEvent() integration.IntegrationEvent {
	return integration.IntegrationEvent{
		ApplicationId:   123,
		ApplicationName: "test-application",
		DeviceName:      "test-device",
		DevEui:          []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		IntegrationName: "foo",
		EventType:       "bar",
		ObjectJson:      `{"foo":"bar"}`,
		Tags: map[string]string{
			"test": "tag",
		},
	}
}

func (ts *MarshalerTestSuite) TestProtobuf() {
	uplinkEvent := ts.GetUplinkEvent()

	assert := require.New(ts.T())

	b, err := Marshal(Protobuf, &uplinkEvent)
	assert.NoError(err)

	var msg integration.UplinkEvent
	assert.NoError(proto.Unmarshal(b, &msg))

	assert.True(proto.Equal(&uplinkEvent, &msg))
}

func (ts *MarshalerTestSuite) TestProtobufJSON() {
	uplinkEvent := ts.GetUplinkEvent()

	assert := require.New(ts.T())

	b, err := Marshal(ProtobufJSON, &uplinkEvent)
	assert.NoError(err)

	var msg integration.UplinkEvent
	assert.NoError(unmarshalProtobufJSON(b, &msg))

	assert.True(proto.Equal(&uplinkEvent, &msg))
}

func (ts *MarshalerTestSuite) TestJSONV3() {
	ts.T().Run("UplinkEvent", func(t *testing.T) {
		uplinkEvent := ts.GetUplinkEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &uplinkEvent)
		assert.NoError(err)

		uplinkTime, err := ptypes.Timestamp(uplinkEvent.RxInfo[0].Time)
		assert.NoError(err)
		uplinkTime = uplinkTime.UTC()

		var pl models.DataUpPayload
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.DataUpPayload{
			ApplicationID:     123,
			ApplicationName:   "test-application",
			DeviceName:        "test-device",
			DeviceProfileName: "test-profile",
			DeviceProfileID:   "f293e453-6d9c-4a22-8c4d-99b2dbe4e94f",
			DevEUI:            lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			RXInfo: []models.RXInfo{
				{
					GatewayID: lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
					Time:      &uplinkTime,
					RSSI:      110,
					LoRaSNR:   5.6,
					Location: &models.Location{
						Latitude:  1.1234,
						Longitude: 2.1234,
						Altitude:  3.1,
					},
				},
			},
			TXInfo: models.TXInfo{
				Frequency: 868100000,
				DR:        3,
			},
			ADR:   true,
			FCnt:  123,
			FPort: 101,
			Data:  []byte{0x01, 0x02, 0x03, 0x04},
			Tags: map[string]string{
				"test": "tag",
			},
			Object: map[string]interface{}{
				"foo": "bar",
			},
		}, pl)
	})

	ts.T().Run("JoinNotification", func(t *testing.T) {
		joinEvent := ts.GetJoinEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &joinEvent)
		assert.NoError(err)

		joinTime, err := ptypes.Timestamp(joinEvent.RxInfo[0].Time)
		assert.NoError(err)
		joinTime = joinTime.UTC()

		var pl models.JoinNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.JoinNotification{
			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			DevAddr:         lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
			RXInfo: []models.RXInfo{
				{
					GatewayID: lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
					Time:      &joinTime,
					RSSI:      110,
					LoRaSNR:   5.6,
					Location: &models.Location{
						Latitude:  1.1234,
						Longitude: 2.1234,
						Altitude:  3.1,
					},
				},
			},
			TXInfo: models.TXInfo{
				Frequency: 868100000,
				DR:        3,
			},
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("ACKNotification", func(t *testing.T) {
		ackEvent := ts.GetAckEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &ackEvent)
		assert.NoError(err)

		var pl models.ACKNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.ACKNotification{
			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			Acknowledged:    true,
			FCnt:            123,
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("ErrorNotification", func(t *testing.T) {
		errorEvent := ts.GetErrorEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &errorEvent)
		assert.NoError(err)

		var pl models.ErrorNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.ErrorNotification{
			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			Type:            "UPLINK_CODEC",
			Error:           "function error",
			FCnt:            110,
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("StatusNotification", func(t *testing.T) {
		statusEvent := ts.GetStatusEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &statusEvent)
		assert.NoError(err)

		var pl models.StatusNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.StatusNotification{
			ApplicationID:           123,
			ApplicationName:         "test-application",
			DeviceName:              "test-device",
			DevEUI:                  lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			Margin:                  10,
			ExternalPowerSource:     true,
			BatteryLevel:            55,
			BatteryLevelUnavailable: true,
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("LocationNotification", func(t *testing.T) {
		locationEvent := ts.GetLocationEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &locationEvent)
		assert.NoError(err)

		var pl models.LocationNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.LocationNotification{
			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			Location: models.Location{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  100,
			},
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("TxAckNotification", func(t *testing.T) {
		txAckEvent := ts.GetTxAckEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &txAckEvent)
		assert.NoError(err)

		var pl models.TxAckNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.TxAckNotification{
			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			FCnt:            123,
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})

	ts.T().Run("IntegrationNotification", func(t *testing.T) {
		event := ts.GetIntegrationEvent()

		assert := require.New(t)
		b, err := Marshal(JSONV3, &event)
		assert.NoError(err)

		var pl models.IntegrationNotification
		assert.NoError(json.Unmarshal(b, &pl))

		assert.Equal(models.IntegrationNotification{

			ApplicationID:   123,
			ApplicationName: "test-application",
			DeviceName:      "test-device",
			DevEUI:          lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			Object: map[string]interface{}{
				"foo": "bar",
			},
			Tags: map[string]string{
				"test": "tag",
			},
		}, pl)
	})
}

func TestMarshaler(t *testing.T) {
	suite.Run(t, new(MarshalerTestSuite))
}
