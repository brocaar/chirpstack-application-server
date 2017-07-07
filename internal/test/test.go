package test

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/loraserver/api/ns"
)

// Config contains the test configuration.
type Config struct {
	PostgresDSN  string
	RedisURL     string
	MQTTServer   string
	MQTTUsername string
	MQTTPassword string
}

// GetConfig returns the test configuration.
func GetConfig() *Config {
	log.SetLevel(log.ErrorLevel)

	c := &Config{
		PostgresDSN: "postgres://localhost/loraserver?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		MQTTServer:  "tcp://localhost:1883",
	}

	if v := os.Getenv("TEST_POSTGRES_DSN"); v != "" {
		c.PostgresDSN = v
	}

	if v := os.Getenv("TEST_REDIS_URL"); v != "" {
		c.RedisURL = v
	}

	if v := os.Getenv("TEST_MQTT_SERVER"); v != "" {
		c.MQTTServer = v
	}

	if v := os.Getenv("TEST_MQTT_USERNAME"); v != "" {
		c.MQTTUsername = v
	}

	if v := os.Getenv("TEST_MQTT_PASSWORD"); v != "" {
		c.MQTTPassword = v
	}

	return c
}

// MustResetDB re-applies all database migrations.
func MustResetDB(db *sqlx.DB) {
	m := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "",
	}
	if _, err := migrate.Exec(db.DB, "postgres", m, migrate.Down); err != nil {
		log.Fatal(err)
	}
	if _, err := migrate.Exec(db.DB, "postgres", m, migrate.Up); err != nil {
		log.Fatal(err)
	}
}

// MustFlushRedis flushes the Redis storage.
func MustFlushRedis(p *redis.Pool) {
	c := p.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHALL"); err != nil {
		log.Fatal(err)
	}
}

// NetworkServerClient is a test network-server client.
type NetworkServerClient struct {
	CreateGatewayChan   chan ns.CreateGatewayRequest
	GetGatewayChan      chan ns.GetGatewayRequest
	UpdateGatewayChan   chan ns.UpdateGatewayRequest
	DeleteGatewayChan   chan ns.DeleteGatewayRequest
	GetGatewayStatsChan chan ns.GetGatewayStatsRequest
	ListGatewayChan     chan ns.ListGatewayRequest

	CreateNodeSessionChan     chan ns.CreateNodeSessionRequest
	GetNodeSessionChan        chan ns.GetNodeSessionRequest
	UpdateNodeSessionChan     chan ns.UpdateNodeSessionRequest
	DeleteNodeSessionChan     chan ns.DeleteNodeSessionRequest
	GetRandomDevAddrChan      chan ns.GetRandomDevAddrRequest
	PushDataDownChan          chan ns.PushDataDownRequest
	GetFrameLogsForDevEUIChan chan ns.GetFrameLogsForDevEUIRequest

	CreateGatewayResponse         ns.CreateGatewayResponse
	GetGatewayResponse            ns.GetGatewayResponse
	UpdateGatewayResponse         ns.UpdateGatewayResponse
	DeleteGatewayResponse         ns.DeleteGatewayResponse
	GetGatewayStatsResponse       ns.GetGatewayStatsResponse
	ListGatewayResponse           ns.ListGatewayResponse
	GetFrameLogsForDevEUIResponse ns.GetFrameLogsResponse

	CreateNodeSessionResponse ns.CreateNodeSessionResponse
	GetNodeSessionResponse    ns.GetNodeSessionResponse
	UpdateNodeSessionResponse ns.UpdateNodeSessionResponse
	DeleteNodeSessionResponse ns.DeleteNodeSessionResponse
	GetRandomDevAddrResponse  ns.GetRandomDevAddrResponse
	PushDataDownResponse      ns.PushDataDownResponse
}

// NewNetworkServerClient creates a new NetworkServerClient.
func NewNetworkServerClient() *NetworkServerClient {
	return &NetworkServerClient{
		CreateNodeSessionChan:     make(chan ns.CreateNodeSessionRequest, 100),
		GetNodeSessionChan:        make(chan ns.GetNodeSessionRequest, 100),
		UpdateNodeSessionChan:     make(chan ns.UpdateNodeSessionRequest, 100),
		DeleteNodeSessionChan:     make(chan ns.DeleteNodeSessionRequest, 100),
		GetRandomDevAddrChan:      make(chan ns.GetRandomDevAddrRequest, 100),
		PushDataDownChan:          make(chan ns.PushDataDownRequest, 100),
		CreateGatewayChan:         make(chan ns.CreateGatewayRequest, 100),
		GetGatewayChan:            make(chan ns.GetGatewayRequest, 100),
		UpdateGatewayChan:         make(chan ns.UpdateGatewayRequest, 100),
		DeleteGatewayChan:         make(chan ns.DeleteGatewayRequest, 100),
		GetGatewayStatsChan:       make(chan ns.GetGatewayStatsRequest, 100),
		ListGatewayChan:           make(chan ns.ListGatewayRequest, 100),
		GetFrameLogsForDevEUIChan: make(chan ns.GetFrameLogsForDevEUIRequest, 100),

		CreateGatewayResponse:   ns.CreateGatewayResponse{},
		GetGatewayResponse:      ns.GetGatewayResponse{},
		UpdateGatewayResponse:   ns.UpdateGatewayResponse{},
		DeleteGatewayResponse:   ns.DeleteGatewayResponse{},
		GetGatewayStatsResponse: ns.GetGatewayStatsResponse{},
		ListGatewayResponse:     ns.ListGatewayResponse{},
	}
}

func (n *NetworkServerClient) CreateGateway(ctx context.Context, in *ns.CreateGatewayRequest, opts ...grpc.CallOption) (*ns.CreateGatewayResponse, error) {
	n.CreateGatewayChan <- *in
	return &n.CreateGatewayResponse, nil
}

func (n *NetworkServerClient) GetGateway(ctx context.Context, in *ns.GetGatewayRequest, opts ...grpc.CallOption) (*ns.GetGatewayResponse, error) {
	n.GetGatewayChan <- *in
	return &n.GetGatewayResponse, nil
}

func (n *NetworkServerClient) ListGateways(ctx context.Context, in *ns.ListGatewayRequest, opts ...grpc.CallOption) (*ns.ListGatewayResponse, error) {
	n.ListGatewayChan <- *in
	return &n.ListGatewayResponse, nil
}

func (n *NetworkServerClient) UpdateGateway(ctx context.Context, in *ns.UpdateGatewayRequest, opts ...grpc.CallOption) (*ns.UpdateGatewayResponse, error) {
	n.UpdateGatewayChan <- *in
	return &n.UpdateGatewayResponse, nil
}

func (n *NetworkServerClient) DeleteGateway(ctx context.Context, in *ns.DeleteGatewayRequest, opts ...grpc.CallOption) (*ns.DeleteGatewayResponse, error) {
	n.DeleteGatewayChan <- *in
	return &n.DeleteGatewayResponse, nil
}

func (n *NetworkServerClient) GetGatewayStats(ctx context.Context, in *ns.GetGatewayStatsRequest, opts ...grpc.CallOption) (*ns.GetGatewayStatsResponse, error) {
	n.GetGatewayStatsChan <- *in
	return &n.GetGatewayStatsResponse, nil
}

func (n *NetworkServerClient) CreateNodeSession(ctx context.Context, in *ns.CreateNodeSessionRequest, opts ...grpc.CallOption) (*ns.CreateNodeSessionResponse, error) {
	n.CreateNodeSessionChan <- *in
	return &n.CreateNodeSessionResponse, nil
}

func (n *NetworkServerClient) GetNodeSession(ctx context.Context, in *ns.GetNodeSessionRequest, opts ...grpc.CallOption) (*ns.GetNodeSessionResponse, error) {
	n.GetNodeSessionChan <- *in
	return &n.GetNodeSessionResponse, nil
}

func (n *NetworkServerClient) UpdateNodeSession(ctx context.Context, in *ns.UpdateNodeSessionRequest, opts ...grpc.CallOption) (*ns.UpdateNodeSessionResponse, error) {
	n.UpdateNodeSessionChan <- *in
	return &n.UpdateNodeSessionResponse, nil
}

func (n *NetworkServerClient) DeleteNodeSession(ctx context.Context, in *ns.DeleteNodeSessionRequest, opts ...grpc.CallOption) (*ns.DeleteNodeSessionResponse, error) {
	n.DeleteNodeSessionChan <- *in
	return &n.DeleteNodeSessionResponse, nil
}

func (n *NetworkServerClient) GetRandomDevAddr(ctx context.Context, in *ns.GetRandomDevAddrRequest, opts ...grpc.CallOption) (*ns.GetRandomDevAddrResponse, error) {
	n.GetRandomDevAddrChan <- *in
	return &n.GetRandomDevAddrResponse, nil
}

func (n *NetworkServerClient) PushDataDown(ctx context.Context, in *ns.PushDataDownRequest, opts ...grpc.CallOption) (*ns.PushDataDownResponse, error) {
	n.PushDataDownChan <- *in
	return &n.PushDataDownResponse, nil
}

func (n *NetworkServerClient) EnqueueDataDownMACCommand(ctx context.Context, in *ns.EnqueueDataDownMACCommandRequest, opts ...grpc.CallOption) (*ns.EnqueueDataDownMACCommandResponse, error) {
	panic("not implemented")
}

func (n *NetworkServerClient) GetFrameLogsForDevEUI(ctx context.Context, in *ns.GetFrameLogsForDevEUIRequest, opts ...grpc.CallOption) (*ns.GetFrameLogsResponse, error) {
	n.GetFrameLogsForDevEUIChan <- *in
	return &n.GetFrameLogsForDevEUIResponse, nil
}
