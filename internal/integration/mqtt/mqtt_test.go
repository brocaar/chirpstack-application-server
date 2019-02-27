package mqtt

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lorawan"
)

type MQTTHandlerTestSuite struct {
	suite.Suite

	mqttClient  paho.Client
	integration integration.Integrator
	redisPool   *redis.Pool
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
	ts.integration, err = New(
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
			LocationTopicTemplate: "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/location",
		},
	)
	assert.NoError(err)
	time.Sleep(time.Millisecond * 100) // give the backend some time to connect
}

func (ts *MQTTHandlerTestSuite) TearDownSuite() {
	ts.mqttClient.Disconnect(0)
	ts.integration.Close()
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

	uplinkChan := make(chan integration.DataUpPayload, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/rx", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.DataUpPayload
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		uplinkChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.DataUpPayload{
		ApplicationID: 123,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.SendDataUp(pl))
	assert.Equal(pl, <-uplinkChan)
}

func (ts *MQTTHandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	joinChan := make(chan integration.JoinNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/join", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.JoinNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		joinChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.JoinNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-node",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DevAddr:         [4]byte{1, 2, 3, 4},
	}
	assert.NoError(ts.integration.SendJoinNotification(pl))
	assert.Equal(pl, <-joinChan)
}

func (ts *MQTTHandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	ackChan := make(chan integration.ACKNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/ack", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.ACKNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		ackChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.ACKNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		DeviceName:      "test-node",
	}
	assert.NoError(ts.integration.SendACKNotification(pl))
	assert.Equal(pl, <-ackChan)
}

func (ts *MQTTHandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	errChan := make(chan integration.ErrorNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/error", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.ErrorNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		errChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.ErrorNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-node",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Type:            "BOOM",
		Error:           "boom boom boom",
	}
	assert.NoError(ts.integration.SendErrorNotification(pl))
	assert.Equal(pl, <-errChan)
}

func (ts *MQTTHandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	statusChan := make(chan integration.StatusNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/status", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.StatusNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		statusChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.StatusNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Margin:          123,
		Battery:         234,
	}
	assert.NoError(ts.integration.SendStatusNotification(pl))
	assert.Equal(pl, <-statusChan)
}

func (ts *MQTTHandlerTestSuite) TestLocation() {
	assert := require.New(ts.T())

	locationChan := make(chan integration.LocationNotification, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/location", 0, func(c paho.Client, msg paho.Message) {
		var pl integration.LocationNotification
		assert.NoError(json.Unmarshal(msg.Payload(), &pl))
		locationChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := integration.LocationNotification{
		ApplicationID:   123,
		ApplicationName: "test-app",
		DeviceName:      "test-device",
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Location: integration.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
		},
	}
	assert.NoError(ts.integration.SendLocationNotification(pl))
	assert.Equal(pl, <-locationChan)
}

func (ts *MQTTHandlerTestSuite) TestDownlink() {
	assert := require.New(ts.T())

	pl := integration.DataDownPayload{
		Confirmed: false,
		FPort:     1,
		Data:      []byte("hello"),
	}

	b, err := json.Marshal(pl)
	assert.NoError(err)

	token := ts.mqttClient.Publish("application/123/device/0102030405060708/tx", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())
	assert.Equal(integration.DataDownPayload{
		ApplicationID: 123,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		Confirmed:     false,
		FPort:         1,
		Data:          []byte("hello"),
		Object:        json.RawMessage("null"),
	}, <-ts.integration.DataDownChan())

	ts.T().Run("invalid fport", func(t *testing.T) {
		assert := require.New(t)

		for _, fPort := range []uint8{0, 225} {
			pl.FPort = fPort

			b, err := json.Marshal(pl)
			assert.NoError(err)
			token := ts.mqttClient.Publish("application/123/device/0102030405060708/tx", 0, false, b)
			token.Wait()
			assert.NoError(token.Error())
			assert.Len(ts.integration.DataDownChan(), 0)
		}
	})
}

func TestMQTTHandler(t *testing.T) {
	suite.Run(t, new(MQTTHandlerTestSuite))
}
