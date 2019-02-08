package httphandler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/handler"
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

func TestHandlerConfig(t *testing.T) {
	testTable := []struct {
		Name          string
		HandlerConfig HandlerConfig
		Valid         bool
	}{
		{
			Name: "Valid headers",
			HandlerConfig: HandlerConfig{
				Headers: map[string]string{
					"Foo":     "Bar",
					"Foo-Bar": "Test",
				},
			},
			Valid: true,
		},
		{
			Name: "Invalid space in header name",
			HandlerConfig: HandlerConfig{
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

	handler     *Handler
	httpHandler *testHTTPHandler
	server      *httptest.Server
}

func (ts *HandlerTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpHandler = &testHTTPHandler{
		requests: make(chan *http.Request, 100),
	}

	ts.server = httptest.NewServer(ts.httpHandler)

	conf := HandlerConfig{
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
	ts.handler, err = NewHandler(conf)
	assert.NoError(err)
}

func (ts *HandlerTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *HandlerTestSuite) TestUplink() {
	assert := require.New(ts.T())

	reqPL := handler.DataUpPayload{
		Data: []byte{1, 2, 3, 4},
	}
	assert.NoError(ts.handler.SendDataUp(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/dataup", req.URL.Path)

	var pl handler.DataUpPayload
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestJoin() {
	assert := require.New(ts.T())

	reqPL := handler.JoinNotification{
		DevAddr: lorawan.DevAddr{1, 2, 3, 4},
	}
	assert.NoError(ts.handler.SendJoinNotification(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/join", req.URL.Path)

	var pl handler.JoinNotification
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestAck() {
	assert := require.New(ts.T())

	reqPL := handler.ACKNotification{
		DevEUI: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}
	assert.NoError(ts.handler.SendACKNotification(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/ack", req.URL.Path)

	var pl handler.ACKNotification
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestError() {
	assert := require.New(ts.T())

	reqPL := handler.ErrorNotification{
		Error: "boom!",
	}
	assert.NoError(ts.handler.SendErrorNotification(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/error", req.URL.Path)

	var pl handler.ErrorNotification
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestStatus() {
	assert := require.New(ts.T())

	reqPL := handler.StatusNotification{
		Battery: 123,
	}
	assert.NoError(ts.handler.SendStatusNotification(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/status", req.URL.Path)

	var pl handler.StatusNotification
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func (ts *HandlerTestSuite) TestLocation() {
	assert := require.New(ts.T())

	reqPL := handler.LocationNotification{
		Location: handler.Location{
			Latitude:  1.123,
			Longitude: 2.123,
			Altitude:  3.123,
		},
	}
	assert.NoError(ts.handler.SendLocationNotification(reqPL))

	req := <-ts.httpHandler.requests
	assert.Equal("/location", req.URL.Path)

	var pl handler.LocationNotification
	assert.NoError(json.NewDecoder(req.Body).Decode(&pl))
	assert.Equal(reqPL, pl)
	assert.Equal(reqPL, pl)
	assert.Equal("Bar", req.Header.Get("Foo"))
	assert.Equal("application/json", req.Header.Get("Content-Type"))
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
