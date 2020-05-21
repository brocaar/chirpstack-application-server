package mydevices

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
	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
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
		Endpoint: ts.server.URL,
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
		DevEui: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		FCnt:   10,
		FPort:  11,
		Data:   []byte{0x01, 0x02, 0x03, 0x04},
		RxInfo: []*gw.UplinkRXInfo{
			{
				GatewayId: []byte{0x08, 0x07, 0x6, 0x5, 0x04, 0x03, 0x02, 0x01},
				Rssi:      20,
				LoraSnr:   5,
				Location: &common.Location{
					Latitude:  1.12345,
					Longitude: 2.12345,
				},
			},
		},
	}

	ctxID, err := uuid.NewV4()
	assert.NoError(err)
	ctx := context.Background()
	ctx = context.WithValue(ctx, logging.ContextIDKey, ctxID)

	assert.NoError(ts.integration.HandleUplinkEvent(ctx, nil, nil, pl))

	req := <-ts.httpHandler.requests
	assert.Equal("application/json", req.Header.Get("Content-Type"))

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(err)

	var pl2 uplinkPayload
	assert.NoError(json.Unmarshal(b, &pl2))

	assert.Equal(uplinkPayload{
		CorrelationID: ctxID.String(),
		DevEUI:        lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		FCnt:          10,
		FPort:         11,
		Data:          []byte{0x01, 0x02, 0x03, 0x04},
		RXInfo: []rxInfo{
			{
				GatewayID: lorawan.EUI64{0x08, 0x07, 0x6, 0x5, 0x04, 0x03, 0x02, 0x01},
				RSSI:      20,
				LoRaSNR:   5,
				Location: location{
					Latitude:  1.12345,
					Longitude: 2.12345,
				},
			},
		},
	}, pl2)
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
