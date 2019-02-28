package multi

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/integration"
	httpint "github.com/brocaar/lora-app-server/internal/integration/http"
	mqttint "github.com/brocaar/lora-app-server/internal/integration/mqtt"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
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

	mqttClient mqtt.Client
	httpServer *httptest.Server

	mqttMessages chan mqtt.Message
	httpRequests chan *http.Request

	integration integration.Integrator
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpRequests = make(chan *http.Request, 100)
	ts.mqttMessages = make(chan mqtt.Message, 100)

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	opts := mqtt.NewClientOptions().AddBroker(conf.ApplicationServer.Integration.MQTT.Server).SetUsername(conf.ApplicationServer.Integration.MQTT.Username).SetPassword(conf.ApplicationServer.Integration.MQTT.Password)
	ts.mqttClient = mqtt.NewClient(opts)
	token := ts.mqttClient.Connect()
	token.Wait()
	assert.NoError(token.Error())

	ts.httpServer = httptest.NewServer(&testHTTPHandler{
		requests: ts.httpRequests,
	})

	token = ts.mqttClient.Subscribe("#", 0, func(c mqtt.Client, msg mqtt.Message) {
		ts.mqttMessages <- msg
	})
	token.Wait()
	assert.NoError(token.Error())

	var err error
	ts.integration, err = New([]interface{}{
		mqttint.Config{
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
		},
		httpint.Config{
			DataUpURL:               ts.httpServer.URL + "/rx",
			JoinNotificationURL:     ts.httpServer.URL + "/join",
			ACKNotificationURL:      ts.httpServer.URL + "/ack",
			ErrorNotificationURL:    ts.httpServer.URL + "/error",
			StatusNotificationURL:   ts.httpServer.URL + "/status",
			LocationNotificationURL: ts.httpServer.URL + "/location",
		},
	})
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.mqttClient.Disconnect(0)
	ts.httpServer.Close()
	ts.integration.Close()
}

func (ts *IntegrationTestSuite) TestSendDataUp() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendDataUp(integration.DataUpPayload{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/rx", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/rx", req.URL.Path)
}

func (ts *IntegrationTestSuite) TestSendJoinNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendJoinNotification(integration.JoinNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/join", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/join", req.URL.Path)
}

func (ts *IntegrationTestSuite) TestSendACKNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendACKNotification(integration.ACKNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/ack", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/ack", req.URL.Path)
}

func (ts *IntegrationTestSuite) TestErrorNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendErrorNotification(integration.ErrorNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/error", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/error", req.URL.Path)
}

func (ts *IntegrationTestSuite) TestStatusNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendStatusNotification(integration.StatusNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/status", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/status", req.URL.Path)
}

func (ts *IntegrationTestSuite) TestLocationNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendLocationNotification(integration.LocationNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	msg := <-ts.mqttMessages
	assert.Equal("application/1/device/0102030405060708/location", msg.Topic())

	req := <-ts.httpRequests
	assert.Equal("/location", req.URL.Path)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
