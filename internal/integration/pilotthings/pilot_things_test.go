package pilotthings

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
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

	integration models.IntegrationHandler
	httpHandler *testHTTPHandler
	server      *httptest.Server
}

func (ts *IntegrationTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpHandler = &testHTTPHandler{
		requests: make(chan *http.Request, 100),
	}

	ts.server = httptest.NewServer(ts.httpHandler)

	conf := Config{
		Server: ts.server.URL,
		Token:  "very secure token",
	}
	var err error
	ts.integration, err = New(conf)
	assert.NoError(err)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *IntegrationTestSuite) TestUplink() {
	assert := require.New(ts.T())

	pl := pb.UplinkEvent{
		DeviceName: "mock device",
		DevEui:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		FPort:      11,
		Data:       []byte{0x01, 0x02, 0x03, 0x04},
		DevAddr:    []byte{0x05, 0x06, 0x07, 0x08},
		FCnt:       10,
		RxInfo: []*gw.UplinkRXInfo{
			{
				Rssi:    1,
				LoraSnr: 1.1,
				RfChain: 1,
				Antenna: 1,
				Board:   1,
			},
			{
				Rssi:    2,
				LoraSnr: 2.1,
				RfChain: 2,
				Antenna: 2,
				Board:   3,
			},
		},
	}

	ctxID, err := uuid.NewV4()
	assert.NoError(err)
	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

	assert.NoError(ts.integration.HandleUplinkEvent(ctx, nil, nil, pl))

	req := <-ts.httpHandler.requests
	assert.Equal("/om2m/ipe-loraserver/up-link?token=very+secure+token", req.RequestURI)
	assert.Equal("application/json", req.Header.Get("Content-Type"))
	assert.Equal("POST", req.Method)

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl2 uplinkPayload
	assert.NoError(json.Unmarshal(b, &pl2))

	assert.Equal(uplinkPayload{
		DeviceName: "mock device",
		Data:       "01020304",
		DevEUI:     "0102030405060708",
		FPort:      11,
		DevAddr:    "05060708",
		FCnt:       10,
		Metadata: []uplinkMetadata{
			{
				Rssi:    1,
				LoraSnr: 1.1,
				RfChain: 1,
				Antenna: 1,
				Board:   1,
			},
			{
				Rssi:    2,
				LoraSnr: 2.1,
				RfChain: 2,
				Antenna: 2,
				Board:   3,
			},
		},
	}, pl2)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
