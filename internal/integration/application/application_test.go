package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	httpint "github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type ApplicationTestSuite struct {
	suite.Suite

	httpServer   *httptest.Server
	httpRequests chan *http.Request
	integration  integration.Integrator
}

func (ts *ApplicationTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.httpRequests = make(chan *http.Request, 100)
	ts.integration = New(marshaler.Protobuf)

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	networkserver.SetPool(mock.NewPool(mock.NewClient()))

	test.MustFlushRedis(storage.RedisPool())
	test.MustResetDB(storage.DB().DB)

	ts.httpServer = httptest.NewServer(&testHTTPHandler{
		requests: ts.httpRequests,
	})

	// setup application with http integration
	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &org))

	ns := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &ns))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: ns.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		OrganizationID:   org.ID,
		Name:             "test-app",
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &app))

	httpConfig := httpint.Config{
		DataUpURL:               ts.httpServer.URL + "/rx",
		JoinNotificationURL:     ts.httpServer.URL + "/join",
		ACKNotificationURL:      ts.httpServer.URL + "/ack",
		ErrorNotificationURL:    ts.httpServer.URL + "/error",
		StatusNotificationURL:   ts.httpServer.URL + "/status",
		LocationNotificationURL: ts.httpServer.URL + "/location",
	}
	configJSON, err := json.Marshal(httpConfig)
	assert.NoError(err)

	assert.NoError(storage.CreateIntegration(context.Background(), storage.DB(), &storage.Integration{
		ApplicationID: app.ID,
		Kind:          integration.HTTP,
		Settings:      configJSON,
	}))
}

func (ts *ApplicationTestSuite) TearDownSuite() {
	ts.httpServer.Close()
	ts.integration.Close()
}

func (ts *ApplicationTestSuite) TestSendDataUp() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendDataUp(context.Background(), nil, pb.UplinkEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/rx", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestSendJoinNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendJoinNotification(context.Background(), nil, pb.JoinEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/join", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestSendACKNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendACKNotification(context.Background(), nil, pb.AckEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/ack", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestErrorNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendErrorNotification(context.Background(), nil, pb.ErrorEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/error", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestStatusNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendStatusNotification(context.Background(), nil, pb.StatusEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/status", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestLocationNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendLocationNotification(context.Background(), nil, pb.LocationEvent{
		ApplicationId: 1,
		DevEui:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/location", req.URL.Path)
}

func TestApplication(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}
