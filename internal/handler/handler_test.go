package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMultiHandler(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given an MQTT client and handler, Redis and PostgreSQL databases and test http handler", t, func() {
		opts := mqtt.NewClientOptions().AddBroker(conf.MQTTServer).SetUsername(conf.MQTTUsername).SetPassword(conf.MQTTPassword)
		c := mqtt.NewClient(opts)
		token := c.Connect()
		token.Wait()
		So(token.Error(), ShouldBeNil)

		p := storage.NewRedisPool(conf.RedisURL)
		test.MustFlushRedis(p)

		db, err := storage.OpenDatabase(conf.PostgresDSN)
		So(err, ShouldBeNil)
		test.MustResetDB(db)

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

		mqttHandler, err := NewMQTTHandler(p, conf.MQTTServer, conf.MQTTUsername, conf.MQTTPassword, "")
		So(err, ShouldBeNil)

		Convey("Given an organization, application with http integration and node", func() {
			org := storage.Organization{
				Name: "test-org",
			}
			So(storage.CreateOrganization(db, &org), ShouldBeNil)

			app := storage.Application{
				OrganizationID: org.ID,
				Name:           "test-app",
			}
			So(storage.CreateApplication(db, &app), ShouldBeNil)

			config := HTTPHandlerConfig{
				DataUpURL:            server.URL + "/rx",
				JoinNotificationURL:  server.URL + "/join",
				ACKNotificationURL:   server.URL + "/ack",
				ErrorNotificationURL: server.URL + "/error",
			}
			configJSON, err := json.Marshal(config)
			So(err, ShouldBeNil)

			So(storage.CreateIntegration(db, &storage.Integration{
				ApplicationID: app.ID,
				Kind:          HTTPHandlerKind,
				Settings:      configJSON,
			}), ShouldBeNil)

			node := storage.Node{
				ApplicationID: app.ID,
				Name:          "test-node",
				DevEUI:        lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
				AppEUI:        lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2},
				AppKey:        lorawan.AES128Key{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
			}
			So(storage.CreateNode(db, node), ShouldBeNil)

			Convey("Getting the multi-handler for the created application", func() {
				multiHandler := NewMultiHandler(db, mqttHandler)
				defer multiHandler.Close()

				Convey("Calling SendDataUp", func() {
					So(multiHandler.SendDataUp(DataUpPayload{
						ApplicationID: app.ID,
						DevEUI:        node.DevEUI,
					}), ShouldBeNil)

					Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
						So(mqttMessages, ShouldHaveLength, 1)
						msg := <-mqttMessages
						So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/rx")

						So(h.requests, ShouldHaveLength, 1)
						req := <-h.requests
						So(req.URL.Path, ShouldEqual, "/rx")
					})
				})

				Convey("Calling SendJoinNotification", func() {
					So(multiHandler.SendJoinNotification(JoinNotification{
						ApplicationID: app.ID,
						DevEUI:        node.DevEUI,
					}), ShouldBeNil)

					Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
						So(mqttMessages, ShouldHaveLength, 1)
						msg := <-mqttMessages
						So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/join")

						So(h.requests, ShouldHaveLength, 1)
						req := <-h.requests
						So(req.URL.Path, ShouldEqual, "/join")
					})
				})

				Convey("Calling SendACKNotification", func() {
					So(multiHandler.SendACKNotification(ACKNotification{
						ApplicationID: app.ID,
						DevEUI:        node.DevEUI,
					}), ShouldBeNil)

					Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
						So(mqttMessages, ShouldHaveLength, 1)
						msg := <-mqttMessages
						So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/ack")

						So(h.requests, ShouldHaveLength, 1)
						req := <-h.requests
						So(req.URL.Path, ShouldEqual, "/ack")
					})
				})

				Convey("Calling SendErrorNotification", func() {
					So(multiHandler.SendErrorNotification(ErrorNotification{
						ApplicationID: app.ID,
						DevEUI:        node.DevEUI,
					}), ShouldBeNil)

					Convey("Then the payload was sent to both the MQTT and HTTP handler", func() {
						So(mqttMessages, ShouldHaveLength, 1)
						msg := <-mqttMessages
						So(msg.Topic(), ShouldEqual, "application/1/node/0101010101010101/error")

						So(h.requests, ShouldHaveLength, 1)
						req := <-h.requests
						So(req.URL.Path, ShouldEqual, "/error")
					})
				})
			})
		})
	})
}
