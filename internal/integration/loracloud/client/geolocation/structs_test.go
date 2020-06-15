package geolocation

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/lorawan"
)

type StructsTestSuite struct {
	suite.Suite
}

func (ts *StructsTestSuite) TestTDOASingleFrameRequest() {
	assert := require.New(ts.T())

	rxInfo := gw.UplinkRXInfo{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Rssi:      1,
		LoraSnr:   2,
		Antenna:   3,
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 1.223,
			Altitude:  1.323,
		},
		FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
			PlainFineTimestamp: &gw.PlainFineTimestamp{
				Time: &timestamp.Timestamp{
					Nanos: 12345,
				},
			},
		},
	}

	req := NewTDOASingleFrameRequest([]*gw.UplinkRXInfo{&rxInfo})
	assert.Equal(TDOASingleFrameRequest{
		LoRaWAN: []UplinkTDOA{
			{
				GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				RSSI:      1,
				SNR:       2,
				TOA:       12345,
				AntennaID: 3,
				AntennaLocation: AntennaLocation{
					Latitude:  1.123,
					Longitude: 1.223,
					Altitude:  1.323,
				},
			},
		},
	}, req)
}

func (ts *StructsTestSuite) TestTDOAMultiFrameRequest() {
	assert := require.New(ts.T())

	rxInfo := gw.UplinkRXInfo{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Rssi:      1,
		LoraSnr:   2,
		Antenna:   3,
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 1.223,
			Altitude:  1.323,
		},
		FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
			PlainFineTimestamp: &gw.PlainFineTimestamp{
				Time: &timestamp.Timestamp{
					Nanos: 12345,
				},
			},
		},
	}

	req := NewTDOAMultiFrameRequest([][]*gw.UplinkRXInfo{{&rxInfo}})
	assert.Equal(TDOAMultiFrameRequest{
		LoRaWAN: [][]UplinkTDOA{
			{
				{
					GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					RSSI:      1,
					SNR:       2,
					TOA:       12345,
					AntennaID: 3,
					AntennaLocation: AntennaLocation{
						Latitude:  1.123,
						Longitude: 1.223,
						Altitude:  1.323,
					},
				},
			},
		},
	}, req)
}

func (ts *StructsTestSuite) TestRSSISingleFrameRequest() {
	assert := require.New(ts.T())

	rxInfo := gw.UplinkRXInfo{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Rssi:      1,
		LoraSnr:   2,
		Antenna:   3,
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 1.223,
			Altitude:  1.323,
		},
		FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
			PlainFineTimestamp: &gw.PlainFineTimestamp{
				Time: &timestamp.Timestamp{
					Nanos: 12345,
				},
			},
		},
	}

	req := NewRSSISingleFrameRequest([]*gw.UplinkRXInfo{&rxInfo})
	assert.Equal(RSSISingleFrameRequest{
		LoRaWAN: []UplinkRSSI{
			{
				GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				RSSI:      1,
				SNR:       2,
				AntennaID: 3,
				AntennaLocation: AntennaLocation{
					Latitude:  1.123,
					Longitude: 1.223,
					Altitude:  1.323,
				},
			},
		},
	}, req)
}

func (ts *StructsTestSuite) TestRSSIMultiFrameRequest() {
	assert := require.New(ts.T())

	rxInfo := gw.UplinkRXInfo{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Rssi:      1,
		LoraSnr:   2,
		Antenna:   3,
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 1.223,
			Altitude:  1.323,
		},
		FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
			PlainFineTimestamp: &gw.PlainFineTimestamp{
				Time: &timestamp.Timestamp{
					Nanos: 12345,
				},
			},
		},
	}

	req := NewRSSIMultiFrameRequest([][]*gw.UplinkRXInfo{{&rxInfo}})
	assert.Equal(RSSIMultiFrameRequest{
		LoRaWAN: [][]UplinkRSSI{
			{
				{
					GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					RSSI:      1,
					SNR:       2,
					AntennaID: 3,
					AntennaLocation: AntennaLocation{
						Latitude:  1.123,
						Longitude: 1.223,
						Altitude:  1.323,
					},
				},
			},
		},
	}, req)
}

func (ts *StructsTestSuite) TestGNSSLR1110SingleFrameRequest() {
	ts.T().Run("Only payload", func(t *testing.T) {
		assert := require.New(t)

		rxInfo := &gw.UplinkRXInfo{}
		pl := []byte{1, 2, 3}
		req := NewGNSSLR1110SingleFrameRequest([]*gw.UplinkRXInfo{rxInfo}, false, pl)

		assert.Equal(GNSSLR1110SingleFrameRequest{
			Payload: helpers.HEXBytes(pl),
		}, req)
	})

	ts.T().Run("With GPS time", func(t *testing.T) {
		assert := require.New(t)
		dur := ptypes.DurationProto(time.Second * 60)
		captTime := float64(60 - 6)
		acc := float64(15)

		rxInfo := &gw.UplinkRXInfo{
			TimeSinceGpsEpoch: dur,
		}
		pl := []byte{1, 2, 3}
		req := NewGNSSLR1110SingleFrameRequest([]*gw.UplinkRXInfo{rxInfo}, true, pl)

		assert.Equal(GNSSLR1110SingleFrameRequest{
			Payload:                 helpers.HEXBytes(pl),
			GNSSCaptureTime:         &captTime,
			GNSSCaptureTimeAccuracy: &acc,
		}, req)
	})

	ts.T().Run("With location", func(t *testing.T) {
		assert := require.New(t)

		rxInfo := &gw.UplinkRXInfo{
			Location: &common.Location{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
			},
		}
		pl := []byte{1, 2, 3}
		req := NewGNSSLR1110SingleFrameRequest([]*gw.UplinkRXInfo{rxInfo}, false, pl)

		assert.Equal(GNSSLR1110SingleFrameRequest{
			Payload:            helpers.HEXBytes(pl),
			GNSSAssistPosition: []float64{1.123, 2.123},
			GNSSAssistAltitude: &rxInfo.Location.Altitude,
		}, req)
	})
}

func (ts *StructsTestSuite) TestWifiTDOASingleFrameRequest() {
	assert := require.New(ts.T())

	rxInfo := &gw.UplinkRXInfo{
		GatewayId: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Rssi:      1,
		LoraSnr:   2,
		Antenna:   3,
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 1.223,
			Altitude:  1.323,
		},
		FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
			PlainFineTimestamp: &gw.PlainFineTimestamp{
				Time: &timestamp.Timestamp{
					Nanos: 12345,
				},
			},
		},
	}

	wifiAP := WifiAccessPoint{
		MacAddress:     BSSID{1, 2, 3, 4, 5, 6},
		SignalStrength: -10,
	}

	req := NewWifiTDOASingleFrameRequest([]*gw.UplinkRXInfo{rxInfo}, []WifiAccessPoint{wifiAP})

	assert.Equal(WifiTDOASingleFrameRequest{
		LoRaWAN: []UplinkTDOA{
			{
				GatewayID: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				RSSI:      1,
				SNR:       2,
				TOA:       12345,
				AntennaID: 3,
				AntennaLocation: AntennaLocation{
					Latitude:  1.123,
					Longitude: 1.223,
					Altitude:  1.323,
				},
			},
		},
		WifiAccessPoints: []WifiAccessPoint{wifiAP},
	}, req)

}

func TestStructs(t *testing.T) {
	suite.Run(t, new(StructsTestSuite))
}
