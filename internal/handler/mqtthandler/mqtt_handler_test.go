package mqtthandler

import (
	"testing"

	"encoding/json"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/gusseleet/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMQTTHandler(t *testing.T) {
	conf := test.GetConfig()

	Convey("Given a MQTT client and Redis database", t, func() {
		opts := mqtt.NewClientOptions().AddBroker(conf.MQTTServer).SetUsername(conf.MQTTUsername).SetPassword(conf.MQTTPassword)
		c := mqtt.NewClient(opts)
		token := c.Connect()
		token.Wait()
		So(token.Error(), ShouldBeNil)

		config.C.Redis.Pool = storage.NewRedisPool(conf.RedisURL)
		test.MustFlushRedis(config.C.Redis.Pool)

		Convey("Given a new MQTTHandler", func() {
			h, err := NewHandler(conf.MQTTServer, conf.MQTTUsername, conf.MQTTPassword, "", "", "")
			So(err, ShouldBeNil)
			defer h.Close()
			time.Sleep(time.Millisecond * 100) // give the backend some time to connect

			Convey("Given the MQTT client is subscribed to application/123/node/0102030405060708/rx", func() {
				dataUpChan := make(chan handler.DataUpPayload, 1)
				token := c.Subscribe("application/123/node/0102030405060708/rx", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl handler.DataUpPayload
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					dataUpChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending a DataUpPayload (from the handler)", func() {
					pl := handler.DataUpPayload{
						ApplicationID: 123,
						DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					}
					So(h.SendDataUp(pl), ShouldBeNil)

					Convey("Then the same payload is consumed by the MQTT client", func() {
						So(<-dataUpChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/123/node/0102030405060708/join", func() {
				joinChan := make(chan handler.JoinNotification)
				token := c.Subscribe("application/123/node/0102030405060708/join", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl handler.JoinNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					joinChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending a join notification (from the handler)", func() {
					pl := handler.JoinNotification{
						ApplicationID:   123,
						ApplicationName: "test-app",
						DeviceName:      "test-node",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						DevAddr:         [4]byte{1, 2, 3, 4},
					}
					So(h.SendJoinNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-joinChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/123/node/0102030405060708/ack", func() {
				ackChan := make(chan handler.ACKNotification)
				token := c.Subscribe("application/123/node/0102030405060708/ack", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl handler.ACKNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					ackChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending an ack notification (from the handler)", func() {
					pl := handler.ACKNotification{
						ApplicationID:   123,
						ApplicationName: "test-app",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						DeviceName:      "test-node",
						Reference:       "1234",
					}
					So(h.SendACKNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-ackChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/123/node/0102030405060708/error", func() {
				errChan := make(chan handler.ErrorNotification)
				token := c.Subscribe("application/123/node/0102030405060708/error", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl handler.ErrorNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					errChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending an error notification (from the handler)", func() {
					pl := handler.ErrorNotification{
						ApplicationID:   123,
						ApplicationName: "test-app",
						DeviceName:      "test-node",
						DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
						Type:            "BOOM",
						Error:           "boom boom boom",
					}
					So(h.SendErrorNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-errChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given a DataDownPayload", func() {
				pl := handler.DataDownPayload{
					Confirmed: false,
					FPort:     1,
					Data:      []byte("hello"),
				}

				Convey("When published with a valid (1-244) port", func() {
					b, err := json.Marshal(pl)
					So(err, ShouldBeNil)
					token := c.Publish("application/123/node/0102030405060708/tx", 0, false, b)
					token.Wait()
					So(token.Error(), ShouldBeNil)

					Convey("Then the same payload is received by the handler", func() {
						So(<-h.DataDownChan(), ShouldResemble, handler.DataDownPayload{
							ApplicationID: 123,
							DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
							Confirmed:     false,
							FPort:         1,
							Data:          []byte("hello"),
							Object:        json.RawMessage("null"),
						})
					})
				})

				Convey("When published with FPort 0, the handler drops the payload", func() {
					pl.FPort = 0
					b, err := json.Marshal(pl)
					So(err, ShouldBeNil)
					token := c.Publish("application/123/node/0102030405060708/tx", 0, false, b)
					token.Wait()
					So(token.Error(), ShouldBeNil)

					So(h.DataDownChan(), ShouldHaveLength, 0)
				})

				Convey("When published with FPort > 224, the handler drops the payload", func() {
					pl.FPort = 225
					b, err := json.Marshal(pl)
					So(err, ShouldBeNil)
					token := c.Publish("application/123/node/0102030405060708/tx", 0, false, b)
					token.Wait()
					So(token.Error(), ShouldBeNil)

					So(h.DataDownChan(), ShouldHaveLength, 0)
				})
			})
		})
	})
}
