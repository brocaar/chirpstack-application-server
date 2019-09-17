package thingsboard

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lorawan"
)

type testHTTPHandler struct {
	requests chan *http.Request
}

func (h *testHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	h.requests <- r
	w.WriteHeader(http.StatusOK)
}

type IntegrationTestSuite struct {
	suite.Suite

	integration integration.Integrator
	httpHandler *testHTTPHandler
	server      *httptest.Server
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpHandler = &testHTTPHandler{
		requests: make(chan *http.Request, 100),
	}
	ts.server = httptest.NewServer(ts.httpHandler)

	conf := Config{
		Server: ts.server.URL,
	}

	var err error
	ts.integration, err = New(conf)
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *IntegrationTestSuite) TestUplink() {
	tests := []struct {
		Name           string
		Payload        integration.DataUpPayload
		ExpectedBodies map[string]string
	}{
		{
			Name: "One level depth",
			Payload: integration.DataUpPayload{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				TXInfo: integration.TXInfo{
					Frequency: 868100000,
					DR:        2,
				},
				Object: map[string]interface{}{
					"temperature": 25.4,
					"humidity":    20,
					"active":      true,
					"status":      "on",
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Variables: map[string]string{
					"ThingsBoardAccessToken": "verysecret",
				},
			},
			ExpectedBodies: map[string]string{
				"/api/v1/verysecret/attributes": `{"application_id":"0","application_name":"test-app","dev_eui":"0102030405060708","device_name":"test-dev","foo":"bar"}`,
				"/api/v1/verysecret/telemetry":  `{"data_active":true,"data_humidity":20,"data_status":"on","data_temperature":25.4}`,
			},
		},
		{
			Name: "One level depth with nil value",
			Payload: integration.DataUpPayload{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				TXInfo: integration.TXInfo{
					Frequency: 868100000,
					DR:        2,
				},
				Object: map[string]interface{}{
					"temperature": nil,
					"humidity":    20,
					"active":      true,
					"status":      "on",
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Variables: map[string]string{
					"ThingsBoardAccessToken": "verysecret",
				},
			},
			ExpectedBodies: map[string]string{
				"/api/v1/verysecret/attributes": `{"application_id":"0","application_name":"test-app","dev_eui":"0102030405060708","device_name":"test-dev","foo":"bar"}`,
				"/api/v1/verysecret/telemetry":  `{"data_active":true,"data_humidity":20,"data_status":"on"}`,
			},
		},
		{
			Name: "Mixed level depth",
			Payload: integration.DataUpPayload{
				ApplicationName: "test-app",
				DeviceName:      "test-dev",
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				FCnt:            10,
				FPort:           20,
				TXInfo: integration.TXInfo{
					Frequency: 868100000,
					DR:        2,
				},
				Object: map[string]interface{}{
					"temperature": map[string]interface{}{
						"a": 20.5,
						"b": 33.3,
					},
					"humidity": 20,
					"active":   true,
					"status":   "on",
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Variables: map[string]string{
					"ThingsBoardAccessToken": "verysecret",
				},
			},
			ExpectedBodies: map[string]string{
				"/api/v1/verysecret/attributes": `{"application_id":"0","application_name":"test-app","dev_eui":"0102030405060708","device_name":"test-dev","foo":"bar"}`,
				"/api/v1/verysecret/telemetry":  `{"data_active":true,"data_humidity":20,"data_status":"on","data_temperature_a":20.5,"data_temperature_b":33.3}`,
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(ts.integration.SendDataUp(context.Background(), tst.Payload))

			for _, _ = range tst.ExpectedBodies {
				req := <-ts.httpHandler.requests
				assert.Equal("application/json", req.Header.Get("Content-Type"))

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(err)
				assert.Equal(tst.ExpectedBodies[req.URL.Path], string(b))
			}
		})
	}
}

func (ts *IntegrationTestSuite) TestDeviceStatus() {
	tests := []struct {
		Name           string
		Payload        integration.StatusNotification
		ExpectedBodies map[string]string
	}{
		{
			Name: "margin and battery status",
			Payload: integration.StatusNotification{
				ApplicationName: "test-app",
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				DeviceName:      "test-dev",
				Battery:         123,
				BatteryLevel:    48.43,
				Margin:          10,
				Tags: map[string]string{
					"foo": "bar",
				},
				Variables: map[string]string{
					"ThingsBoardAccessToken": "verysecret",
				},
			},
			ExpectedBodies: map[string]string{
				"/api/v1/verysecret/attributes": `{"application_id":"0","application_name":"test-app","dev_eui":"0102030405060708","device_name":"test-dev","foo":"bar"}`,
				"/api/v1/verysecret/telemetry":  `{"status_battery":123,"status_battery_level":48.43,"status_battery_level_unavailable":false,"status_external_power_source":false,"status_margin":10}`,
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(ts.integration.SendStatusNotification(context.Background(), tst.Payload))

			for _, _ = range tst.ExpectedBodies {
				req := <-ts.httpHandler.requests
				assert.Equal("application/json", req.Header.Get("Content-Type"))

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(err)
				assert.Equal(tst.ExpectedBodies[req.URL.Path], string(b))
			}
		})
	}
}

func (ts *IntegrationTestSuite) TestLocation() {
	tests := []struct {
		Name           string
		Payload        integration.LocationNotification
		ExpectedBodies map[string]string
	}{
		{
			Name: "location",
			Payload: integration.LocationNotification{
				ApplicationName: "test-app",
				DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				DeviceName:      "test-dev",
				Location: integration.Location{
					Latitude:  1.123,
					Longitude: 2.123,
					Altitude:  3.123,
				},
				Tags: map[string]string{
					"foo": "bar",
				},
				Variables: map[string]string{
					"ThingsBoardAccessToken": "verysecret",
				},
			},
			ExpectedBodies: map[string]string{
				"/api/v1/verysecret/attributes": `{"application_id":"0","application_name":"test-app","dev_eui":"0102030405060708","device_name":"test-dev","foo":"bar"}`,
				"/api/v1/verysecret/telemetry":  `{"location_altitude":3.123,"location_latitude":1.123,"location_longitude":2.123}`,
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)
			assert.NoError(ts.integration.SendLocationNotification(context.Background(), tst.Payload))

			for _, _ = range tst.ExpectedBodies {
				req := <-ts.httpHandler.requests
				assert.Equal("application/json", req.Header.Get("Content-Type"))

				b, err := ioutil.ReadAll(req.Body)
				assert.NoError(err)
				assert.Equal(tst.ExpectedBodies[req.URL.Path], string(b))
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
