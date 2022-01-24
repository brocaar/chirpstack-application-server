package azureservicebus

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestParseConnectionString(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		expectedKV       map[string]string
		expectedError    error
	}{
		{
			name:             "valid string",
			connectionString: "Endpoint=sb://chirpstack-tst.servicebus.windows.net/;SharedAccessKeyName=TestKeyName;SharedAccessKey=TestKey",
			expectedKV: map[string]string{
				"Endpoint":            "sb://chirpstack-tst.servicebus.windows.net/",
				"SharedAccessKeyName": "TestKeyName",
				"SharedAccessKey":     "TestKey",
			},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			assert := require.New(t)
			kv, err := parseConnectionString(tst.connectionString)
			assert.Equal(tst.expectedError, err)
			if err != nil {
				return
			}

			assert.EqualValues(tst.expectedKV, kv)
		})
	}
}

func TestCreateSASToken(t *testing.T) {
	assert := require.New(t)
	var exp time.Time
	token, err := createSASToken("https://chirpstack-tst.servicebus.windows.net/", "MyKey", "AQID", exp)
	assert.NoError(err)
	assert.Equal("SharedAccessSignature sig=%2BzYoiYqfIOWmNoHwvAPChBukHMHPFxNT1nhYGvCnKLg%3D&se=-62135596800&skn=MyKey&sr=https%3A%2F%2Fchirpstack-tst.servicebus.windows.net%2F", token)
}

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

	integration models.IntegrationHandler
	httpHandler *testHTTPHandler
	server      *httptest.Server
}

func (ts *IntegrationTestSuite) SetupSuite() {
	ts.httpHandler = &testHTTPHandler{
		requests: make(chan *http.Request, 100),
	}

	ts.server = httptest.NewServer(ts.httpHandler)

	ts.integration = &Integration{
		marshaler:   marshaler.Protobuf,
		publishName: "my-queue",
		uri:         ts.server.URL,
	}
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *IntegrationTestSuite) TestUplink() {
	assert := require.New(ts.T())

	reqPL := pb.UplinkEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.UplinkEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"up"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestJoin() {
	assert := require.New(ts.T())

	reqPL := pb.JoinEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.JoinEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"join"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestAck() {
	assert := require.New(ts.T())

	reqPL := pb.AckEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.AckEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"ack"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestError() {
	assert := require.New(ts.T())

	reqPL := pb.ErrorEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.ErrorEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"error"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestStatus() {
	assert := require.New(ts.T())

	reqPL := pb.StatusEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.StatusEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"status"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestLocation() {
	assert := require.New(ts.T())

	reqPL := pb.LocationEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.LocationEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"location"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestTxAck() {
	assert := require.New(ts.T())

	reqPL := pb.TxAckEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.TxAckEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"txack"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func (ts *IntegrationTestSuite) TestIntegration() {
	assert := require.New(ts.T())

	reqPL := pb.IntegrationEvent{
		ApplicationId: 123,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleIntegrationEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.IntegrationEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))

	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.NotEqual("", req.Header.Get("Authorization"))
	assert.Equal(`"integration"`, req.Header.Get("event"))
	assert.Equal("123", req.Header.Get("application_id"))
	assert.Equal(`"0102030405060708"`, req.Header.Get("dev_eui"))
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
