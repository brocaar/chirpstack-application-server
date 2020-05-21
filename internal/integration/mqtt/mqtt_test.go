package mqtt

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type MQTTHandlerTestSuite struct {
	suite.Suite

	mqttClient  paho.Client
	integration models.IntegrationHandler
}

func (ts *MQTTHandlerTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	mqttServer := conf.ApplicationServer.Integration.MQTT.Server
	username := conf.ApplicationServer.Integration.MQTT.Username
	password := conf.ApplicationServer.Integration.MQTT.Password

	opts := paho.NewClientOptions().AddBroker(mqttServer).SetUsername(username).SetPassword(password)
	ts.mqttClient = paho.NewClient(opts)
	token := ts.mqttClient.Connect()
	token.Wait()
	assert.NoError(token.Error())

	var err error
	ts.integration, err = New(
		marshaler.Protobuf,
		config.IntegrationMQTTConfig{
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
			TxAckTopicTemplate:    "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/txack",
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
	storage.RedisClient().FlushAll()
}

func (ts *MQTTHandlerTestSuite) TestUplink() {
	assert := require.New(ts.T())

	uplinkChan := make(chan pb.UplinkEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/rx", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.UplinkEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		uplinkChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.UplinkEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-uplinkChan)
}

func (ts *MQTTHandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	joinChan := make(chan pb.JoinEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/join", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.JoinEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		joinChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.JoinEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
		DevAddr:       []byte{1, 2, 3, 4},
	}
	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-joinChan)
}

func (ts *MQTTHandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	ackChan := make(chan pb.AckEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/ack", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.AckEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		ackChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.AckEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-ackChan)
}

func (ts *MQTTHandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	errChan := make(chan pb.ErrorEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/error", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.ErrorEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		errChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.ErrorEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-errChan)
}

func (ts *MQTTHandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	statusChan := make(chan pb.StatusEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/status", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.StatusEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		statusChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.StatusEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}

	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-statusChan)
}

func (ts *MQTTHandlerTestSuite) TestLocation() {
	assert := require.New(ts.T())

	locationChan := make(chan pb.LocationEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/location", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.LocationEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		locationChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.LocationEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-locationChan)
}

func (ts *MQTTHandlerTestSuite) TestTxAck() {
	assert := require.New(ts.T())

	txAckChan := make(chan pb.TxAckEvent, 1)
	token := ts.mqttClient.Subscribe("application/123/device/0102030405060708/txack", 0, func(c paho.Client, msg paho.Message) {
		var pl pb.TxAckEvent
		assert.NoError(proto.Unmarshal(msg.Payload(), &pl))
		txAckChan <- pl
	})
	token.Wait()
	assert.NoError(token.Error())

	pl := pb.TxAckEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, nil, pl))
	assert.Equal(pl, <-txAckChan)
}

func (ts *MQTTHandlerTestSuite) TestDownlink() {
	assert := require.New(ts.T())

	pl := models.DataDownPayload{
		Confirmed: false,
		FPort:     1,
		Data:      []byte("hello"),
	}

	b, err := json.Marshal(pl)
	assert.NoError(err)

	token := ts.mqttClient.Publish("application/123/device/0102030405060708/tx", 0, false, b)
	token.Wait()
	assert.NoError(token.Error())
	assert.Equal(models.DataDownPayload{
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
