package influxdb

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

type testHTTPHandler struct {
	requests chan *http.Request
}

func (h *testHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	h.requests <- r
	w.WriteHeader(http.StatusOK)
}

type IntegrationV1TestSuite struct {
	suite.Suite

	Handler  models.IntegrationHandler
	Requests chan *http.Request
	Server   *httptest.Server
}

func (ts *IntegrationV1TestSuite) SetupSuite() {
	assert := require.New(ts.T())
	ts.Requests = make(chan *http.Request, 100)

	httpHandler := testHTTPHandler{
		requests: ts.Requests,
	}
	ts.Server = httptest.NewServer(&httpHandler)

	conf := Config{
		Endpoint:            ts.Server.URL + "/write",
		DB:                  "chirpstack",
		Username:            "user",
		Password:            "password",
		RetentionPolicyName: "DEFAULT",
		Precision:           "s",
	}
	var err error
	ts.Handler, err = New(conf)
	assert.NoError(err)
}

func (ts *IntegrationV1TestSuite) TearDownSuite() {
	ts.Server.Close()
}

func (ts *IntegrationV1TestSuite) TestStatus() {
	tests := []struct {
		Name         string
		Payload      pb.StatusEvent
		ExpectedBody string
	}{
		{
			Name: "margin and battery status",
			Payload: pb.StatusEvent{
				ApplicationName: "test-app",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DeviceName:      "test-device",
				BatteryLevel:    48.43,
				Margin:          10,
				Tags: map[string]string{
					"foo": "bar",
				},
			},
			ExpectedBody: `device_status_battery_level,application_name=test-app,dev_eui=0102030405060708,device_name=test-device,foo=bar value=48.430000
device_status_margin,application_name=test-app,dev_eui=0102030405060708,device_name=test-device,foo=bar value=10i`,
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(ts.Handler.HandleStatusEvent(context.Background(), nil, nil, tst.Payload))
			req := <-ts.Requests
			assert.Equal("/write", req.URL.Path)
			assert.Equal(url.Values{
				"db":        []string{"chirpstack"},
				"precision": []string{"s"},
				"rp":        []string{"DEFAULT"},
			}, req.URL.Query())

			b, err := ioutil.ReadAll(req.Body)
			assert.NoError(err)
			assert.Equal(tst.ExpectedBody, string(b))

			user, pw, ok := req.BasicAuth()
			assert.Equal("user", user)
			assert.Equal("password", pw)
			assert.True(ok)

			assert.Equal("text/plain", req.Header.Get("Content-Type"))
		})
	}
}

func (ts *IntegrationV1TestSuite) TestUplink() {
	tests := []struct {
		Name         string
		Payload      pb.UplinkEvent
		ExpectedBody string
	}{
		{
			Name: "One level depth",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				ObjectJson: `{
					"temperature": 25.4,
					"humidity":    20,
					"active":      true,
					"status":      "on"
				}`,
				Tags: map[string]string{
					"fo o": "ba,r",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\ o=ba\,r value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\ o=ba\,r value=20.000000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\ o=ba\,r value="on"
device_frmpayload_data_temperature,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\ o=ba\,r value=25.400000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,fo\ o=ba\,r,frequency=868100000 f_cnt=10i,value=1i`,
		},
		{
			Name: "One level depth with nil value",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				ObjectJson: `{
					"temperature": null,
					"humidity":    20,
					"active":      true,
					"status":      "on"
				}`,
				Tags: map[string]string{
					"fo=o": "bar",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\=o=bar value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\=o=bar value=20.000000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,fo\=o=bar value="on"
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,fo\=o=bar,frequency=868100000 f_cnt=10i,value=1i`,
		},
		{
			Name: "One level depth + RXInfo",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				RxInfo: []*gw.UplinkRXInfo{
					{
						LoraSnr: 1,
						Rssi:    -60,
					},
					{
						LoraSnr: 2.5,
						Rssi:    -55,
					},
					{
						LoraSnr: 1,
						Rssi:    -70,
					},
				},
				ObjectJson: `{
					"temperature": 25.4,
					"humidity":    20,
					"active":      true,
					"status":      "on"
				}`,
				Tags: map[string]string{
					"foo": "bar",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=20.000000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value="on"
device_frmpayload_data_temperature,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=25.400000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,foo=bar,frequency=868100000 f_cnt=10i,rssi=-55i,snr=2.500000,value=1i`,
		},
		{
			Name: "Mixed level depth",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				ObjectJson: `{
					"temperature": {
						"a": 20.5,
						"b": 33.3
					},
					"humidity": 20,
					"active":   true,
					"status":   "on"
				}`,
				Tags: map[string]string{
					"foo": "bar",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=20.000000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value="on"
device_frmpayload_data_temperature_a,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=20.500000
device_frmpayload_data_temperature_b,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=33.300000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,foo=bar,frequency=868100000 f_cnt=10i,value=1i`,
		},
		{
			Name: "One level depth + device status fields",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				ObjectJson: `{
					"temperature": 25.4,
					"humidity":    20,
					"active":      true,
					"status":      "on"
				}`,
				Tags: map[string]string{
					"foo": "bar",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=20.000000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value="on"
device_frmpayload_data_temperature,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=25.400000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,foo=bar,frequency=868100000 f_cnt=10i,value=1i`,
		},
		{
			Name: "Latitude and longitude",
			Payload: pb.UplinkEvent{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEui:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				Dr:              2,
				TxInfo: &gw.UplinkTXInfo{
					Frequency: 868100000,
				},
				ObjectJson: `{
					"latitude":  1.123,
					"longitude": 2.123,
					"active":    true,
					"status":    "on"
				}`,
				Tags: map[string]string{
					"foo": "bar",
				},
			},
			ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value=true
device_frmpayload_data_location,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar geohash="s01w2k3vvqre",latitude=1.123000,longitude=2.123000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20,foo=bar value="on"
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,foo=bar,frequency=868100000 f_cnt=10i,value=1i`,
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(ts.Handler.HandleUplinkEvent(context.Background(), nil, nil, tst.Payload))
			req := <-ts.Requests
			assert.Equal("/write", req.URL.Path)
			assert.Equal(url.Values{
				"db":        []string{"chirpstack"},
				"precision": []string{"s"},
				"rp":        []string{"DEFAULT"},
			}, req.URL.Query())

			b, err := ioutil.ReadAll(req.Body)
			assert.NoError(err)
			assert.Equal(tst.ExpectedBody, string(b))

			user, pw, ok := req.BasicAuth()
			assert.Equal("user", user)
			assert.Equal("password", pw)
			assert.True(ok)

			assert.Equal("text/plain", req.Header.Get("Content-Type"))
		})
	}
}

type IntegrationV2TestSuite struct {
	suite.Suite

	Handler  *Integration
	Requests chan *http.Request
	Server   *httptest.Server
}

func (ts *IntegrationV2TestSuite) SetupSuite() {
	assert := require.New(ts.T())
	ts.Requests = make(chan *http.Request, 100)

	httpHandler := testHTTPHandler{
		requests: ts.Requests,
	}
	ts.Server = httptest.NewServer(&httpHandler)

	conf := Config{
		Endpoint:     ts.Server.URL + "/write",
		Version:      2,
		Token:        "test-token",
		Organization: "test-org",
		Bucket:       "test-bucket",
	}
	var err error
	ts.Handler, err = New(conf)
	assert.NoError(err)
}

func (ts *IntegrationV2TestSuite) TearDownSuite() {
	ts.Server.Close()
}

// TestSend tests the send method with the InfluxDB v2 parameters.
func (ts *IntegrationV2TestSuite) TestSend() {
	assert := require.New(ts.T())
	assert.NoError(ts.Handler.send([]measurement{
		{
			Name:   "test_measurement",
			Tags:   map[string]string{"foo": "bar"},
			Values: map[string]interface{}{"temperature": 22.5},
		},
	}))

	req := <-ts.Requests
	assert.Equal("/write", req.URL.Path)
	assert.Equal(url.Values{
		"org":    []string{"test-org"},
		"bucket": []string{"test-bucket"},
	}, req.URL.Query())

	assert.Equal("Token test-token", req.Header.Get("Authorization"))
}

func TestV1Integration(t *testing.T) {
	suite.Run(t, new(IntegrationV1TestSuite))
}

func TestV2Integration(t *testing.T) {
	suite.Run(t, new(IntegrationV2TestSuite))
}
