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
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
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

	integration integration.Integrator
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
		DataUpURL:               ts.server.URL + "/dataup",
		JoinNotificationURL:     ts.server.URL + "/join",
		ACKNotificationURL:      ts.server.URL + "/ack",
		ErrorNotificationURL:    ts.server.URL + "/error",
		StatusNotificationURL:   ts.server.URL + "/status",
		LocationNotificationURL: ts.server.URL + "/location",
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
	assert.NoError(ts.integration.SendDataUp(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/dataup", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.UplinkEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	reqPL := pb.JoinEvent{
		DevAddr: []byte{1, 2, 3, 4},
	}
	assert.NoError(ts.integration.SendJoinNotification(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/join", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.JoinEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	reqPL := pb.AckEvent{
		DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.integration.SendACKNotification(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/ack", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.AckEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	reqPL := pb.ErrorEvent{
		Error: "boom!",
	}
	assert.NoError(ts.integration.SendErrorNotification(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/error", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.ErrorEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	reqPL := pb.StatusEvent{
		BatteryLevel: 55,
	}
	assert.NoError(ts.integration.SendStatusNotification(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/status", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.StatusEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
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
	assert.NoError(ts.integration.SendLocationNotification(context.Background(), nil, reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/location", req.URL.Path)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl pb.LocationEvent
	assert.NoError(proto.Unmarshal(b, &pl))
	assert.True(proto.Equal(&reqPL, &pl))
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/octet-stream", req.Header.Get("Content-Type"))
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
