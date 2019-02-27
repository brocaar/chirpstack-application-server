package application

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/integration"
	httpint "github.com/brocaar/lora-app-server/internal/integration/http"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
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
	ts.integration = New()

	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	assert.NoError(err)
	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = storage.NewRedisPool(conf.RedisURL, 10, 0)
	config.C.NetworkServer.Pool = test.NewNetworkServerPool(test.NewNetworkServerClient())

	test.MustFlushRedis(config.C.Redis.Pool)
	test.MustResetDB(config.C.PostgreSQL.DB)

	ts.httpServer = httptest.NewServer(&testHTTPHandler{
		requests: ts.httpRequests,
	})

	// setup application with http integration
	org := storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(config.C.PostgreSQL.DB, &org))

	ns := storage.NetworkServer{
		Name:   "test-ns",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(config.C.PostgreSQL.DB, &ns))

	sp := storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: ns.ID,
	}
	assert.NoError(storage.CreateServiceProfile(config.C.PostgreSQL.DB, &sp))
	spID, err := uuid.FromBytes(sp.ServiceProfile.Id)
	assert.NoError(err)

	app := storage.Application{
		OrganizationID:   org.ID,
		Name:             "test-app",
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(config.C.PostgreSQL.DB, &app))

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

	assert.NoError(storage.CreateIntegration(config.C.PostgreSQL.DB, &storage.Integration{
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
	assert.NoError(ts.integration.SendDataUp(integration.DataUpPayload{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/rx", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestSendJoinNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendJoinNotification(integration.JoinNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/join", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestSendACKNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendACKNotification(integration.ACKNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/ack", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestErrorNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendErrorNotification(integration.ErrorNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/error", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestStatusNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendStatusNotification(integration.StatusNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/status", req.URL.Path)
}

func (ts *ApplicationTestSuite) TestLocationNotification() {
	assert := require.New(ts.T())
	assert.NoError(ts.integration.SendLocationNotification(integration.LocationNotification{
		ApplicationID: 1,
		DevEUI:        lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}))

	req := <-ts.httpRequests
	assert.Equal("/location", req.URL.Path)
}

func TestApplication(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}
