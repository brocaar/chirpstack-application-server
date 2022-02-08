package geolocation

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/lorawan"
)

type ClientTestSuite struct {
	suite.Suite

	apiResponse string
	apiRequest  string
	server      *httptest.Server
	client      *Client
}

func (ts *ClientTestSuite) SetupSuite() {
	ts.server = httptest.NewServer(http.HandlerFunc(ts.apiHandler))
	ts.client = New(true, ts.server.URL, "foobar")
}

func (ts *ClientTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *ClientTestSuite) TestTDOASingleFrame() {
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

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Result: &LocationResult{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Accuracy:  10,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.TDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  10,
			Source:    common.LocationSource_GEO_RESOLVER_TDOA,
		}, resp)

		var req TDOASingleFrameRequest
		assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
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
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.TDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.Equal("api returned error(s): boom, boom!", err.Error())
	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.TDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) TestTDOAMultiFrame() {
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

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Result: &LocationResult{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Accuracy:  10,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.TDOAMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  10,
			Source:    common.LocationSource_GEO_RESOLVER_TDOA,
		}, resp)

		var req TDOAMultiFrameRequest
		assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
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
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.TDOAMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.Equal("api returned error(s): boom, boom!", err.Error())
	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.TDOAMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) TestRSSISingleFrame() {
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

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Result: &LocationResult{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Accuracy:  10,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.RSSISingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  10,
			Source:    common.LocationSource_GEO_RESOLVER_RSSI,
		}, resp)

		var req RSSISingleFrameRequest
		assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
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
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.RSSISingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.Equal("api returned error(s): boom, boom!", err.Error())
	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.RSSISingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) TestRSSIMultiFrame() {
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

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Result: &LocationResult{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Accuracy:  10,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.RSSIMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  10,
			Source:    common.LocationSource_GEO_RESOLVER_RSSI,
		}, resp)

		var req RSSIMultiFrameRequest
		assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
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
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.RSSIMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.Equal("api returned error(s): boom, boom!", err.Error())
	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.RSSIMultiFrame(context.Background(), [][]*gw.UplinkRXInfo{{&rxInfo}})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) TestGNSSLR1110SingleFrame() {
	rxInfo := gw.UplinkRXInfo{}

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := V3Response{
			Result: &LocationSolverResult{
				LLH:      []float64{1.123, 2.123, 3.123},
				Accuracy: 4.123,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.GNSSLR1110SingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, false, []byte{1, 2, 3})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  4,
			Source:    common.LocationSource_GEO_RESOLVER_GNSS,
		}, resp)

		var req GNSSLR1110SingleFrameRequest
		assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
		assert.Equal(GNSSLR1110SingleFrameRequest{
			Payload: helpers.HEXBytes([]byte{1, 2, 3}),
		}, req)
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := V3Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.GNSSLR1110SingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, false, []byte{1, 2, 3})
		assert.Equal("api returned error(s): boom, boom!", err.Error())
	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := V3Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.GNSSLR1110SingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, false, []byte{1, 2, 3})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) TestWifiTDOASingleFrame() {
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

	ts.T().Run("Result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Result: &LocationResult{
				Latitude:  1.123,
				Longitude: 2.123,
				Altitude:  3.123,
				Accuracy:  10,
			},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		resp, err := ts.client.WifiTDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, []WifiAccessPoint{
			{MacAddress: BSSID{1, 1, 1, 1, 1, 1}, SignalStrength: -10},
			{MacAddress: BSSID{2, 2, 2, 2, 2, 2}, SignalStrength: -20},
			{MacAddress: BSSID{3, 3, 3, 3, 3, 3}, SignalStrength: -30},
		})
		assert.NoError(err)
		assert.Equal(common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
			Accuracy:  10,
			Source:    common.LocationSource_GEO_RESOLVER_WIFI,
		}, resp)
	})

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		result := Response{
			Errors: []string{"boom", "boom!"},
		}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.WifiTDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, []WifiAccessPoint{})
		assert.Equal("api returned error(s): boom, boom!", err.Error())

	})

	ts.T().Run("No result", func(t *testing.T) {
		assert := require.New(t)

		result := Response{}
		resultB, err := json.Marshal(result)
		assert.NoError(err)
		ts.apiResponse = string(resultB)

		_, err = ts.client.WifiTDOASingleFrame(context.Background(), []*gw.UplinkRXInfo{&rxInfo}, []WifiAccessPoint{})
		assert.Equal(ErrNoLocation, err)
	})
}

func (ts *ClientTestSuite) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	ts.apiRequest = string(b)
	w.Write([]byte(ts.apiResponse))
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
