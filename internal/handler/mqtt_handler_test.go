package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
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

		p := storage.NewRedisPool(conf.RedisURL)
		test.MustFlushRedis(p)

		Convey("Given a new MQTTHandler", func() {
			handler, err := NewMQTTHandler(p, conf.MQTTServer, conf.MQTTUsername, conf.MQTTPassword)
			So(err, ShouldBeNil)
			defer handler.Close()
			time.Sleep(time.Millisecond * 100) // give the backend some time to connect

			Convey("Given the MQTT client is subscribed to application/test-app/node/test-node/rx", func() {
				dataUpChan := make(chan DataUpPayload)
				token := c.Subscribe("application/test-app/node/test-node/rx", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl DataUpPayload
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					dataUpChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending a DataUpPayload (from the handler)", func() {
					pl := DataUpPayload{
						ApplicationName: "test-app",
						NodeName:        "test-node",
					}
					So(handler.SendDataUp(pl), ShouldBeNil)

					Convey("Then the same payload is consumed by the MQTT client", func() {
						So(<-dataUpChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/test-app/node/test-node/join", func() {
				joinChan := make(chan JoinNotification)
				token := c.Subscribe("application/test-app/node/test-node/join", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl JoinNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					joinChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending a join notification (from the handler)", func() {
					pl := JoinNotification{
						ApplicationName: "test-app",
						NodeName:        "test-node",
						DevAddr:         [4]byte{1, 2, 3, 4},
					}
					So(handler.SendJoinNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-joinChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/test-app/node/test-node/ack", func() {
				ackChan := make(chan ACKNotification)
				token := c.Subscribe("application/test-app/node/test-node/ack", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl ACKNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					ackChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending an ack notification (from the handler)", func() {
					pl := ACKNotification{
						ApplicationName: "test-app",
						NodeName:        "test-node",
						Reference:       "1234",
					}
					So(handler.SendACKNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-ackChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given the MQTT client is subscribed to application/test-app/node/test-node/error", func() {
				errChan := make(chan ErrorNotification)
				token := c.Subscribe("application/test-app/node/test-node/error", 0, func(c mqtt.Client, msg mqtt.Message) {
					var pl ErrorNotification
					if err := json.Unmarshal(msg.Payload(), &pl); err != nil {
						t.Fatal(err)
					}
					errChan <- pl
				})
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("When sending an error notification (from the handler)", func() {
					pl := ErrorNotification{
						ApplicationName: "test-app",
						NodeName:        "test-node",
						Type:            "BOOM",
						Error:           "boom boom boom",
					}
					So(handler.SendErrorNotification(pl), ShouldBeNil)

					Convey("Then the same notification is received by the MQTT client", func() {
						So(<-errChan, ShouldResemble, pl)
					})
				})
			})

			Convey("Given a DataDownPayload is published by the MQTT client", func() {
				pl := DataDownPayload{
					Confirmed: false,
					FPort:     1,
					Data:      []byte("hello"),
				}
				b, err := json.Marshal(pl)
				So(err, ShouldBeNil)
				token := c.Publish("application/test-app/node/test-node/tx", 0, false, b)
				token.Wait()
				So(token.Error(), ShouldBeNil)

				Convey("Then the same payload is received by the handler", func() {
					So(<-handler.DataDownChan(), ShouldResemble, DataDownPayload{
						ApplicationName: "test-app",
						NodeName:        "test-node",
						Confirmed:       false,
						FPort:           1,
						Data:            []byte("hello"),
					})
				})
			})
		})
	})
}
