package mqtthandler

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lorawan"
)

type MQTTHandlerTestSuite struct {
	suite.Suite

	mqttClient paho.Client
	handler    handler.Handler
	redisPool  *redis.Pool
}

func (ts *MQTTHandlerTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	log.SetLevel(log.ErrorLevel)

	mqttServer := "tcp://127.0.0.1:1883"
	redisServer := "redis://localhost:6379/1"
	var username string
	var password string

	if v := os.Getenv("TEST_MQTT_SERVER"); v != "" {
		mqttServer = v
	}

	if v := os.Getenv("TEST_MQTT_USERNAME"); v != "" {
		username = v
	}
	if v := os.Getenv("TEST_MQTT_PASSWORD"); v != "" {
		password = v
	}

	if v := os.Getenv("TEST_REDIS_URL"); v != "" {
		redisServer = v
	}

	opts := paho.NewClientOptions().AddBroker(mqttServer).SetUsername(username).SetPassword(password)
	ts.mqttClient = paho.NewClient(opts)
	token := ts.mqttClient.Connect()
	token.Wait()
	assert.NoError(token.Error())

	ts.redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisServer)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
	}

	var err error
	ts.handler, err = NewHandler(
		ts.redisPool,
		Config{
			Server:                mqttServer,
			Username:              username,
			Password:              password,
			CleanSession:          true,
			UplinkTopicTemplate:   "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/rx",
			DownlinkTopicTemplate: "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/tx",
			JoinTopicTemplate:     "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/join",
			AckTopicTemplate:      "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/ack",
			ErrorTopicTemplate:    "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/error",
			StatusTopicTemplate:   "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/status",
		},
	)
	assert.NoError(err)
	time.Sleep(time.Millisecond * 100) // give the backend some time to connect
}

func (ts *MQTTHandlerTestSuite) TearDownSuite() {
	ts.mqttClient.Disconnect(0)
	ts.handler.Close()
}

func (ts *MQTTHandlerTestSuite) SetupTest() {
	assert := require.New(ts.T())

	c := ts.redisPool.Get()
	defer c.Close()

	_, err := c.Do("FLUSHALL")
	assert.NoError(err)
}

func (ts *MQTTHandlerTestSuite) TestUplink() {
	assert := require.New(ts.T())

	uplinkChan := make(chan handler.DataUpPayload, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/rx", 0, func(c paho.Client, msg paho.Message) {
		var pl handler.DataUpPayload
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		uplinkChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := handler.DataUpPayload{
		ApplicationID: 123,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.handler.SendDataUp(pl))
	assert.Equal(pl, <-uplinkChan)
}

func (ts *MQTTHandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	joinChan := make(chan handler.JoinNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/join", 0, func(c paho.Client, msg paho.Message) {
		var pl handler.JoinNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		joinChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := handler.JoinNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-node",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DevAddr:         [4]byte{1, 2, 3, 4},
	}
	assert.NoError(ts.handler.SendJoinNotification(pl))
	assert.Equal(pl, <-joinChan)
}

func (ts *MQTTHandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	ackChan := make(chan handler.ACKNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/ack", 0, func(c paho.Client, msg paho.Message) {
		var pl handler.ACKNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		ackChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := handler.ACKNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DeviceName:      "test-node",
	}
	assert.NoError(ts.handler.SendACKNotification(pl))
	assert.Equal(pl, <-ackChan)
}

func (ts *MQTTHandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	errChan := make(chan handler.ErrorNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/error", 0, func(c paho.Client, msg paho.Message) {
		var pl handler.ErrorNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		errChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := handler.ErrorNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-node",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Type:            "BOOM",
		Error:           "boom boom boom",
	}
	assert.NoError(ts.handler.SendErrorNotification(pl))
	assert.Equal(pl, <-errChan)
}

func (ts *MQTTHandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	statusChan := make(chan handler.StatusNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/status", 0, func(c paho.Client, msg paho.Message) {
		var pl handler.StatusNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		statusChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := handler.StatusNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Margin:          123,
		Battery:         234,
	}
	assert.NoError(ts.handler.SendStatusNotification(pl))
	assert.Equal(pl, <-statusChan)
}

func (ts *MQTTHandlerTestSuite) TestDownlink() {
	assert := require.New(ts.T())

	pl := handler.DataDownPayload{
		Confirmed: false,
		FPort:     1,
		Data:      []byte("hello"),
	}

	b, err := json.Marshal(pl)
	assert.NoError(err)

	token := ts.mqttClient.Publish("application/123/device/0102030405060708/tx", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())
	assert.Equal(handler.DataDownPayload{
		ApplicationID: 123,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Confirmed:     false,
		FPort:         1,
		Data:          []byte("hello"),
		Object:        json.RawMessage("null"),
	}, <-ts.handler.DataDownChan())

	ts.T().Run("invalid fport", func(t *testing.T) {
		assert := require.New(t)

		for _, fPort := range []uint8{0, 225} {
			pl.FPort = fPort

			b, err := json.Marshal(pl)
			assert.NoError(err)
			token := ts.mqttClient.Publish("application/123/device/0102030405060708/tx", 0, false, b)
			token.Wait()
			assert.NoError(token.Error())
			assert.Len(ts.handler.DataDownChan(), 0)
		}
	})
}

func TestMQTTHandler(t *testing.T) {
	suite.Run(t, new(MQTTHandlerTestSuite))
}
