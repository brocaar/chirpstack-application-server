package geolocation

import (
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
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

func TestStructs(t *testing.T) {
	suite.Run(t, new(StructsTestSuite))
}
