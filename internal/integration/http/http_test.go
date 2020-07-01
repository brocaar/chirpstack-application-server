package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
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

func TestHandlerConfig(t *testing.T) {
	testTable := []struct {
		Name          string
		HandlerConfig Config
		Valid         bool
	}{
		{
			Name: "Valid headers",
			HandlerConfig: Config{
				Headers: map[string]string{
					"Foo":     "Bar",
					"Foo-Bar": "Test",
				},
			},
			Valid: true,
		},
		{
			Name: "Invalid space in header name",
			HandlerConfig: Config{
				Headers: map[string]string{
					"Invalid Header": "Test",
				},
			},
			Valid: false,
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			assert := require.New(t)
			err := test.HandlerConfig.Validate()
			if test.Valid {
				assert.NoError(err)
			} else {
				assert.NotNil(err)
			}
		})
	}
}

type HandlerTestSuite struct {
	suite.Suite

	integration models.IntegrationHandler
	httpHandler *testHTTPHandler
	server      *httptest.Server
}

func (ts *HandlerTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpHandler = &testHTTPHandler{
		requests: make(chan *http.Request, 100),
	}

	ts.server = httptest.NewServer(ts.httpHandler)

	conf := Config{
		Headers: map[string]string{
			"Foo": "Bar",
		},
		EventEndpointURL: ts.server.URL + "/event?myToken=abc123",
	}

	var err error
	ts.integration, err = New(marshaler.Protobuf, conf)
	assert.NoError(err)
}

func (ts *HandlerTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *HandlerTestSuite) TestUplink() {
	assert := require.New(ts.T())

	reqPL := pb.UplinkEvent{
		Data: []byte{1, 2, 3, 4},
	}
	assert.NoError(ts.integration.HandleUplinkEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.UplinkEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("up", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	reqPL := pb.JoinEvent{
		DevAddr: []byte{1, 2, 3, 4},
	}
	assert.NoError(ts.integration.HandleJoinEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.JoinEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("join", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	reqPL := pb.AckEvent{
		DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.HandleAckEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.AckEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("ack", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	reqPL := pb.ErrorEvent{
		Error: "boom!",
	}
	assert.NoError(ts.integration.HandleErrorEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.ErrorEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("error", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	reqPL := pb.StatusEvent{
		BatteryLevel: 55,
	}
	assert.NoError(ts.integration.HandleStatusEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.StatusEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("status", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestLocation() {
	assert := require.New(ts.T())

	reqPL := pb.LocationEvent{
		Location: &common.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
		},
	}
	assert.NoError(ts.integration.HandleLocationEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.LocationEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("location", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestTxAck() {
	assert := require.New(ts.T())

	reqPL := pb.TxAckEvent{
		FCnt: 123,
	}
	assert.NoError(ts.integration.HandleTxAckEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.TxAckEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("txack", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func (ts *HandlerTestSuite) TestIntegration() {
	assert := require.New(ts.T())

	reqPL := pb.IntegrationEvent{
		IntegrationName: "foo",
	}
	assert.NoError(ts.integration.HandleIntegrationEvent(context.Background(), nil, nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/event", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.IntegrationEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
	assert.Equal("integration", req.URL.Query().Get("event"))
	assert.Equal("abc123", req.URL.Query().Get("myToken"))
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func TestGetURLs(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		In  string
		Out []string
	}{
		{
			In:  "",
			Out: nil,
		},
		{
			In:  ",",
			Out: nil,
		},
		{
			In:  "http://example.com",
			Out: []string{"http://example.com"},
		},
		{
			In:  "http://example.com,",
			Out: []string{"http://example.com"},
		},
		{
			In:  " http://example.com , ",
			Out: []string{"http://example.com"},
		},
		{
			In:  "http://example.com,http://example.nl",
			Out: []string{"http://example.com", "http://example.nl"},
		},
		{
			In:  "http://example.com, http://example.nl",
			Out: []string{"http://example.com", "http://example.nl"},
		},
		{
			In:  "http://example.com , http://example.nl",
			Out: []string{"http://example.com", "http://example.nl"},
		},
		{
			In:  "http://example.com , http://example.nl,",
			Out: []string{"http://example.com", "http://example.nl"},
		},
	}

	for _, tst := range tests {
		assert.Equal(tst.Out, getURLs(tst.In))
	}
}
