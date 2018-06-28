package influxdbhandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/brocaar/lora-app-server/internal/codec"
	"github.com/brocaar/lora-app-server/internal/handler"
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

func TestHandler(t *testing.T) {
	Convey("Given a test HTTP Server and a Handler instance", t, func() {
		httpHandler := testHTTPHandler{
			requests: make(chan *http.Request, 100),
		}
		server := httptest.NewServer(&httpHandler)
		defer server.Close()

		conf := HandlerConfig{
			Endpoint:            server.URL + "/write",
			DB:                  "loraserver",
			Username:            "user",
			Password:            "password",
			RetentionPolicyName: "DEFAULT",
			Precision:           "s",
		}
		h, err := NewHandler(conf)
		So(err, ShouldBeNil)

		Convey("Status testcases", func() {
			tests := []struct {
				Name         string
				Payload      handler.StatusNotification
				ExpectedBody string
			}{
				{
					Name: "margin and battery status",
					Payload: handler.StatusNotification{
						ApplicationName: "test-app",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						DeviceName:      "test-device",
						Battery:         123,
						Margin:          10,
					},
					ExpectedBody: `device_status_battery,application_name=test-app,dev_eui=0102030405060708,device_name=test-device value=123i
device_status_margin,application_name=test-app,dev_eui=0102030405060708,device_name=test-device value=10i`,
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					So(h.SendStatusNotification(test.Payload), ShouldBeNil)
					req := <-httpHandler.requests
					So(req.URL.Path, ShouldEqual, "/write")
					So(req.URL.Query(), ShouldResemble, url.Values{
						"db":        []string{"loraserver"},
						"precision": []string{"s"},
						"rp":        []string{"DEFAULT"},
					})

					b, err := ioutil.ReadAll(req.Body)
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, test.ExpectedBody)

					user, pw, ok := req.BasicAuth()
					So(user, ShouldEqual, conf.Username)
					So(pw, ShouldEqual, conf.Password)
					So(ok, ShouldBeTrue)

					So(req.Header.Get("Content-Type"), ShouldEqual, "text/plain")
				})
			}
		})

		Convey("Uplink data testcases", func() {

			tests := []struct {
				Name         string
				Payload      handler.DataUpPayload
				ExpectedBody string
			}{
				{
					Name: "One level depth",
					Payload: handler.DataUpPayload{
						ApplicationName: "test-app",
						DeviceName:      "test-dev",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						FCnt:            10,
						FPort:           20,
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DR:        2,
						},
						Object: map[string]interface{}{
							"temperature": 25.4,
							"humidity":    20,
							"active":      true,
							"status":      "on",
						},
					},
					ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=20i
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value="on"
device_frmpayload_data_temperature,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=25.400000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,frequency=868100000 value=1i`,
				},
				{
					Name: "Mixed level depth",
					Payload: handler.DataUpPayload{
						ApplicationName: "test-app",
						DeviceName:      "test-dev",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						FCnt:            10,
						FPort:           20,
						TXInfo: handler.TXInfo{
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
					},
					ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=20i
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value="on"
device_frmpayload_data_temperature_a,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=20.500000
device_frmpayload_data_temperature_b,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=33.300000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,frequency=868100000 value=1i`,
				},
				{
					Name: "One level depth + device status fields",
					Payload: handler.DataUpPayload{
						ApplicationName: "test-app",
						DeviceName:      "test-dev",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						FCnt:            10,
						FPort:           20,
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DR:        2,
						},
						Object: map[string]interface{}{
							"temperature": 25.4,
							"humidity":    20,
							"active":      true,
							"status":      "on",
						},
					},
					ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=true
device_frmpayload_data_humidity,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=20i
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value="on"
device_frmpayload_data_temperature,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=25.400000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,frequency=868100000 value=1i`,
				},
				{
					Name: "Latitude and longitude",
					Payload: handler.DataUpPayload{
						ApplicationName: "test-app",
						DeviceName:      "test-dev",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						FCnt:            10,
						FPort:           20,
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DR:        2,
						},
						Object: map[string]interface{}{
							"latitude":  1.123,
							"longitude": 2.123,
							"active":    true,
							"status":    "on",
						},
					},
					ExpectedBody: `device_frmpayload_data_active,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=true
device_frmpayload_data_location,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 geohash="s01w2k3vvqre",latitude=1.123000,longitude=2.123000
device_frmpayload_data_status,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value="on"
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,frequency=868100000 value=1i`,
				},
				{
					Name: "Cayenne LPP with latitude and longitude",
					Payload: handler.DataUpPayload{
						ApplicationName: "test-app",
						DeviceName:      "test-dev",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						FCnt:            10,
						FPort:           20,
						TXInfo: handler.TXInfo{
							Frequency: 868100000,
							DR:        2,
						},
						Object: &codec.CayenneLPP{
							GPSLocation: map[byte]codec.GPSLocation{
								10: codec.GPSLocation{
									Latitude:  1.123,
									Longitude: 2.123,
									Altitude:  3.123,
								},
							},
						},
					},
					ExpectedBody: `device_frmpayload_data_gps_location_10_altitude,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 value=3.123000
device_frmpayload_data_gps_location_10_location,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,f_port=20 geohash="s01w2k3vvqre",latitude=1.123000,longitude=2.123000
device_uplink,application_name=test-app,dev_eui=0102030405060708,device_name=test-dev,dr=2,frequency=868100000 value=1i`,
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
					So(h.SendDataUp(test.Payload), ShouldBeNil)
					req := <-httpHandler.requests
					So(req.URL.Path, ShouldEqual, "/write")
					So(req.URL.Query(), ShouldResemble, url.Values{
						"db":        []string{"loraserver"},
						"precision": []string{"s"},
						"rp":        []string{"DEFAULT"},
					})

					b, err := ioutil.ReadAll(req.Body)
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, test.ExpectedBody)

					user, pw, ok := req.BasicAuth()
					So(user, ShouldEqual, conf.Username)
					So(pw, ShouldEqual, conf.Password)
					So(ok, ShouldBeTrue)

					So(req.Header.Get("Content-Type"), ShouldEqual, "text/plain")
				})
			}
		})
	})
}
