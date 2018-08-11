package multihandler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/brocaar/lora-app-server/internal/handler/influxdbhandler"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/handler/httphandler"
	"github.com/brocaar/lora-app-server/internal/handler/mqtthandler"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/smartystreets/goconvey/convey"
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
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = storage.NewRedisPool(conf.RedisURL)

	Convey("Given an MQTT client and handler, Redis and PostgreSQL databases and test http handler", t, func() {
		opts := mqtt.NewClientOptions().AddBroker(conf.MQTTServer).SetUsername(conf.MQTTUsername).SetPassword(conf.MQTTPassword)
		c := mqtt.NewClient(opts)
		token := c.Connect()
		token.Wait()
		So(token.Error(), ShouldBeNil)

		nsClient := test.NewNetworkServerClient()
		config.C.NetworkServer.Pool = test.NewNetworkServerPool(nsClient)

		test.MustFlushRedis(config.C.Redis.Pool)
		test.MustResetDB(config.C.PostgreSQL.DB)

		h := testHTTPHandler{
			requests: make(chan *http.Request, 100),
		}
		server := httptest.NewServer(&h)
		defer server.Close()

		mqttMessages := make(chan mqtt.Message, 100)
		token = c.Subscribe("#", 0, func(c mqtt.Client, msg mqtt.Message) {
			mqttMessages <- msg
		})
		token.Wait()
		So(token.Error(), ShouldBeNil)

		mqttHandler, err := mqtthandler.NewHandler(config.C.Redis.Pool, mqtthandler.Config{
			Server:                conf.MQTTServer,
			Username:              conf.MQTTUsername,
			Password:              conf.MQTTPassword,
			CleanSession:          true,
			UplinkTopicTemplate:   "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/rx",
			DownlinkTopicTemplate: "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/tx",
			JoinTopicTemplate:     "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/join",
			AckTopicTemplate:      "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/ack",
			ErrorTopicTemplate:    "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/error",
			StatusTopicTemplate:   "application/{{ .ApplicationID }}/node/{{ .DevEUI }}/status",
		})
		So(err, ShouldBeNil)

		Convey("Given an organization, application with http integration and node", func() {
			org := storage.Organization{
				Name: "test-org",
			}
			So(storage.CreateOrganization(db, &org), ShouldBeNil)

			n := storage.NetworkServer{
				Name:   "test-ns",
				Server: "test-ns:1234",
			}
			So(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &n), ShouldBeNil)

			sp := storage.ServiceProfile{
				Name:            "test-sp",
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
			}
			So(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp), ShouldBeNil)
			spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
			So(err, ShouldBeNil)

			app := storage.Application{
				OrganizationID:   org.ID,
				Name:             "test-app",
				ServiceProfileID: spID,
			}
			So(storage.CreateApplication(db, &app), ShouldBeNil)

			dp := storage.DeviceProfile{
				Name:            "test-dp",
				OrganizationID:  org.ID,
				NetworkServerID: n.ID,
			}
			So(storage.CreateDeviceProfile(config.C.PostgreSQL.DB, &dp), ShouldBeNil)
			dpID, err := uuid.FromBytes(dp.DeviceProfile.Id)
			So(err, ShouldBeNil)

			device := storage.Device{
				ApplicationID:   app.ID,
				Name:            "test-node",
				DevEUI:          lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
				DeviceProfileID: dpID,
			}
			So(storage.CreateDevice(db, &device), ShouldBeNil)

			Convey("Given an InfluxDB integration", func() {
				handlerConfig := influxdbhandler.HandlerConfig{
					Endpoint: server.URL + "/write",
					DB:       "loraserver",
				}
				configJSON, err := json.Marshal(handlerConfig)
				So(err, ShouldBeNil)

				So(storage.CreateIntegration(db, &storage.Integration{
					ApplicationID: app.ID,
					Kind:          InfluxDBHandlerKind,
					Settings:      configJSON,
				}), ShouldBeNil)

				Convey("Getting the multi-handler for the created application", func() {
					multiHandler := NewHandler(mqttHandler)
					defer multiHandler.Close()

					Convey("Calling SendDataUp", func() {
						So(multiHandler.SendDataUp(handler.DataUpPayload{
							ApplicationID:   app.ID,
							DevEUI:          device.DevEUI,
							DeviceName:      "test-device",
							ApplicationName: "test-app",
							Object: map[string]interface{}{
								"temp":     10,
								"humidity": 20,
							},
						}), ShouldBeNil)

						msg := <-mqttMessages
						So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/rx")

						req := <-h.requests
						So(req.URL.Path, ShouldEqual, "/write")
					})
				})
			})

			Convey("Given a HTTP integration", func() {
				handlerConfig := httphandler.HandlerConfig{
					DataUpURL:             server.URL + "/rx",
					JoinNotificationURL:   server.URL + "/join",
					ACKNotificationURL:    server.URL + "/ack",
					ErrorNotificationURL:  server.URL + "/error",
					StatusNotificationURL: server.URL + "/status",
				}
				configJSON, err := json.Marshal(handlerConfig)
				So(err, ShouldBeNil)

				So(storage.CreateIntegration(db, &storage.Integration{
					ApplicationID: app.ID,
					Kind:          HTTPHandlerKind,
					Settings:      configJSON,
				}), ShouldBeNil)

				Convey("Getting the multi-handler for the created application", func() {
					multiHandler := NewHandler(mqttHandler)
					defer multiHandler.Close()

					Convey("Calling SendDataUp", func() {
						So(multiHandler.SendDataUp(handler.DataUpPayload{
							ApplicationID: app.ID,
							DevEUI:        device.DevEUI,
						}), ShouldBeNil)

						Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
							msg := <-mqttMessages
							So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/rx")

							req := <-h.requests
							So(req.URL.Path, ShouldEqual, "/rx")
						})
					})

					Convey("Calling SendJoinNotification", func() {
						So(multiHandler.SendJoinNotification(handler.JoinNotification{
							ApplicationID: app.ID,
							DevEUI:        device.DevEUI,
						}), ShouldBeNil)

						Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
							msg := <-mqttMessages
							So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/join")

							req := <-h.requests
							So(req.URL.Path, ShouldEqual, "/join")
						})
					})

					Convey("Calling SendACKNotification", func() {
						So(multiHandler.SendACKNotification(handler.ACKNotification{
							ApplicationID: app.ID,
							DevEUI:        device.DevEUI,
						}), ShouldBeNil)

						Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
							msg := <-mqttMessages
							So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/ack")

							req := <-h.requests
							So(req.URL.Path, ShouldEqual, "/ack")
						})
					})

					Convey("Calling SendErrorNotification", func() {
						So(multiHandler.SendErrorNotification(handler.ErrorNotification{
							ApplicationID: app.ID,
							DevEUI:        device.DevEUI,
						}), ShouldBeNil)

						Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
							msg := <-mqttMessages
							So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/error")

							req := <-h.requests
							So(req.URL.Path, ShouldEqual, "/error")
						})
					})

					Convey("Calling SendStatusNotification", func() {
						So(multiHandler.SendStatusNotification(handler.StatusNotification{
							ApplicationID: app.ID,
							DevEUI:        device.DevEUI,
						}), ShouldBeNil)

						Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
							msg := <-mqttMessages
							So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/status")

							req := <-h.requests
							So(req.URL.Path, ShouldEqual, "/status")
						})
					})
				})
			})
		})
	})
}
