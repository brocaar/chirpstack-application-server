package das

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/lorawan"
)

type ClientTestSuite struct {
	suite.Suite

	apiResponse string
	apiRequest  string
	server      *httptest.Server
	client      *Client
}

func (ts *ClientTestSuite) SetupSuite() {
	ts.server = httptest.NewServer(http.HandlerFunc(ts.apiHandler))
	ts.client = New(ts.server.URL, "foobar")
}

func (ts *ClientTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *ClientTestSuite) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	ts.apiRequest = string(b)
	w.Write([]byte(ts.apiResponse))
}

func (ts *ClientTestSuite) TestUplinkSend() {
	assert := require.New(ts.T())

	devEUI := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	req := UplinkRequest{
		helpers.EUI64(devEUI): &UplinkMsgJoining{
			MsgType:   "joining",
			Timestamp: 1234,
			DR:        3,
			Freq:      868100000,
		},
	}

	result := UplinkResponse{
		Result: UplinkDeviceMapResponse{
			helpers.EUI64(devEUI): UplinkResponseItem{
				Result: UplinkResponseResult{
					Downlink: &LoRaDownlink{
						Port:    10,
						Payload: helpers.HEXBytes{1, 2, 3},
					},
				},
			},
		},
	}
	resultB, err := json.Marshal(result)
	assert.NoError(err)
	ts.apiResponse = string(resultB)

	resp, err := ts.client.UplinkSend(context.Background(), req)
	assert.NoError(err)
	assert.Equal(result, resp)

	var requested UplinkRequest
	assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &requested))
	assert.EqualValues(UplinkRequest{
		helpers.EUI64(devEUI): map[string]interface{}{
			"msgtype":   "joining",
			"dr":        float64(3), // everything is casted to float64
			"freq":      float64(868100000),
			"timestamp": float64(1234),
		},
	}, requested)
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
