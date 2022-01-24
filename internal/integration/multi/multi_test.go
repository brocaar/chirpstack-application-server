package multi

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	httpint "github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	mqttint "github.com/brocaar/chirpstack-application-server/internal/integration/mqtt"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
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

type MultiTestSuite struct {
	suite.Suite

	mqttClient mqtt.Client
	httpServer *httptest.Server

	mqttMessages chan mqtt.Message
	httpRequests chan *http.Request

	integration models.Integration
}

func (ts *MultiTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	// setup storage
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	// setup channels
	ts.httpRequests = make(chan *http.Request, 100)
	ts.mqttMessages = make(chan mqtt.Message, 100)

	// setup mqtt client
	opts := mqtt.NewClientOptions().AddBroker(conf.ApplicationServer.Integration.MQTT.Server).SetUsername(conf.ApplicationServer.Integration.MQTT.Username).SetPassword(conf.ApplicationServer.Integration.MQTT.Password)
	ts.mqttClient = mqtt.NewClient(opts)
	token := ts.mqttClient.Connect()
	token.Wait()
	assert.NoError(token.Error())

	token = ts.mqttClient.Subscribe("#", 0, func(c mqtt.Client, msg mqtt.Message) {
		ts.mqttMessages <- msg
	})
	token.Wait()
	assert.NoError(token.Error())

	// setup http handler
	ts.httpServer = httptest.NewServer(&testHTTPHandler{
		requests: ts.httpRequests,
	})

	// setup integrations
	mi, err := mqttint.New(marshaler.Protobuf, config.IntegrationMQTTConfig{
		Server:                conf.ApplicationServer.Integration.MQTT.Server,
		Username:              conf.ApplicationServer.Integration.MQTT.Username,
		Password:              conf.ApplicationServer.Integration.MQTT.Password,
		CleanSession:          true,
		UplinkTopicTemplate:   "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/rx",
		DownlinkTopicTemplate: "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/tx",
		JoinTopicTemplate:     "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/join",
		AckTopicTemplate:      "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/ack",
		ErrorTopicTemplate:    "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/error",
		StatusTopicTemplate:   "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/status",
		LocationTopicTemplate: "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/location",
		TxAckTopicTemplate:    "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/txack",
	})
	assert.NoError(err)
	globalIntegrations := []models.IntegrationHandler{mi}

	hi, err := httpint.New(marshaler.Protobuf, httpint.Config{
		DataUpURL:               ts.httpServer.URL + "/rx",
		JoinNotificationURL:     ts.httpServer.URL + "/join",
		ACKNotificationURL:      ts.httpServer.URL + "/ack",
		ErrorNotificationURL:    ts.httpServer.URL + "/error",
		StatusNotificationURL:   ts.httpServer.URL + "/status",
		LocationNotificationURL: ts.httpServer.URL + "/location",
		TxAckNotificationURL:    ts.httpServer.URL + "/txack",
	})
	assert.NoError(err)
	appIntegrations := []models.IntegrationHandler{hi}

	ts.integration = New(globalIntegrations, appIntegrations)
}

func (ts *MultiTestSuite) TearDownSuite() {
	ts.mqttClient.Disconnect(0)
	ts.httpServer.Close()
}

func (ts *MultiTestSuite) TestHandleUplinkEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, pb.UplinkEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/rx", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/rx", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleJoinEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, pb.JoinEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/join", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/join", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleAckEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, pb.AckEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/ack", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/ack", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleErrorEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, pb.ErrorEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/error", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/error", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleStatusEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, pb.StatusEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/status", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/status", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleLocationEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, pb.LocationEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/location", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/location", req.URL.Path)
}

func (ts *MultiTestSuite) TestHandleTxAckEvent() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, pb.TxAckEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/txack", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/txack", req.URL.Path)
}

func TestMulti(t *testing.T) {
	suite.Run(t, new(MultiTestSuite))
}
