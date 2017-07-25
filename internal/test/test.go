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
	CreateGatewayChan        chan ns.CreateGatewayRequest
	GetGatewayChan           chan ns.GetGatewayRequest
	UpdateGatewayChan        chan ns.UpdateGatewayRequest
	DeleteGatewayChan        chan ns.DeleteGatewayRequest
	GetGatewayStatsChan      chan ns.GetGatewayStatsRequest
	ListGatewayChan          chan ns.ListGatewayRequest
	GenerateGatewayTokenChan chan ns.GenerateGatewayTokenRequest

	CreateNodeSessionChan     chan ns.CreateNodeSessionRequest
	GetNodeSessionChan        chan ns.GetNodeSessionRequest
	UpdateNodeSessionChan     chan ns.UpdateNodeSessionRequest
	DeleteNodeSessionChan     chan ns.DeleteNodeSessionRequest
	GetRandomDevAddrChan      chan ns.GetRandomDevAddrRequest
	PushDataDownChan          chan ns.PushDataDownRequest
	GetFrameLogsForDevEUIChan chan ns.GetFrameLogsForDevEUIRequest

	CreateChannelConfigurationChan                chan ns.CreateChannelConfigurationRequest
	GetChannelConfigurationChan                   chan ns.GetChannelConfigurationRequest
	UpdateChannelConfigurationChan                chan ns.UpdateChannelConfigurationRequest
	DeleteChannelConfigurationChan                chan ns.DeleteChannelConfigurationRequest
	ListChannelConfigurationsChan                 chan ns.ListChannelConfigurationsRequest
	CreateExtraChannelChan                        chan ns.CreateExtraChannelRequest
	UpdateExtraChannelChan                        chan ns.UpdateExtraChannelRequest
	DeleteExtraChannelChan                        chan ns.DeleteExtraChannelRequest
	GetExtraChannelsForChannelConfigurationIDChan chan ns.GetExtraChannelsForChannelConfigurationIDRequest

	CreateGatewayResponse         ns.CreateGatewayResponse
	GetGatewayResponse            ns.GetGatewayResponse
	UpdateGatewayResponse         ns.UpdateGatewayResponse
	DeleteGatewayResponse         ns.DeleteGatewayResponse
	GetGatewayStatsResponse       ns.GetGatewayStatsResponse
	ListGatewayResponse           ns.ListGatewayResponse
	GetFrameLogsForDevEUIResponse ns.GetFrameLogsResponse
	GenerateGatewayTokenResponse  ns.GenerateGatewayTokenResponse

	CreateChannelConfigurationResponse                ns.CreateChannelConfigurationResponse
	GetChannelConfigurationResponse                   ns.GetChannelConfigurationResponse
	UpdateChannelConfigurationResponse                ns.UpdateChannelConfigurationResponse
	DeleteChannelConfigurationResponse                ns.DeleteChannelConfigurationResponse
	ListChannelConfigurationsResponse                 ns.ListChannelConfigurationsResponse
	CreateExtraChannelResponse                        ns.CreateExtraChannelResponse
	UpdateExtraChannelResponse                        ns.UpdateExtraChannelResponse
	DeleteExtraChannelResponse                        ns.DeleteExtraChannelResponse
	GetExtraChannelsForChannelConfigurationIDResponse ns.GetExtraChannelsForChannelConfigurationIDResponse

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
		CreateNodeSessionChan:                         make(chan ns.CreateNodeSessionRequest, 100),
		GetNodeSessionChan:                            make(chan ns.GetNodeSessionRequest, 100),
		UpdateNodeSessionChan:                         make(chan ns.UpdateNodeSessionRequest, 100),
		DeleteNodeSessionChan:                         make(chan ns.DeleteNodeSessionRequest, 100),
		GetRandomDevAddrChan:                          make(chan ns.GetRandomDevAddrRequest, 100),
		PushDataDownChan:                              make(chan ns.PushDataDownRequest, 100),
		CreateGatewayChan:                             make(chan ns.CreateGatewayRequest, 100),
		GetGatewayChan:                                make(chan ns.GetGatewayRequest, 100),
		UpdateGatewayChan:                             make(chan ns.UpdateGatewayRequest, 100),
		DeleteGatewayChan:                             make(chan ns.DeleteGatewayRequest, 100),
		GetGatewayStatsChan:                           make(chan ns.GetGatewayStatsRequest, 100),
		ListGatewayChan:                               make(chan ns.ListGatewayRequest, 100),
		GenerateGatewayTokenChan:                      make(chan ns.GenerateGatewayTokenRequest, 100),
		GetFrameLogsForDevEUIChan:                     make(chan ns.GetFrameLogsForDevEUIRequest, 100),
		CreateChannelConfigurationChan:                make(chan ns.CreateChannelConfigurationRequest, 100),
		GetChannelConfigurationChan:                   make(chan ns.GetChannelConfigurationRequest, 100),
		UpdateChannelConfigurationChan:                make(chan ns.UpdateChannelConfigurationRequest, 100),
		DeleteChannelConfigurationChan:                make(chan ns.DeleteChannelConfigurationRequest, 100),
		ListChannelConfigurationsChan:                 make(chan ns.ListChannelConfigurationsRequest, 100),
		CreateExtraChannelChan:                        make(chan ns.CreateExtraChannelRequest, 100),
		UpdateExtraChannelChan:                        make(chan ns.UpdateExtraChannelRequest, 100),
		DeleteExtraChannelChan:                        make(chan ns.DeleteExtraChannelRequest, 100),
		GetExtraChannelsForChannelConfigurationIDChan: make(chan ns.GetExtraChannelsForChannelConfigurationIDRequest, 100),

		/*
			CreateGatewayResponse:   ns.CreateGatewayResponse{},
			GetGatewayResponse:      ns.GetGatewayResponse{},
			UpdateGatewayResponse:   ns.UpdateGatewayResponse{},
			DeleteGatewayResponse:   ns.DeleteGatewayResponse{},
			GetGatewayStatsResponse: ns.GetGatewayStatsResponse{},
			ListGatewayResponse:     ns.ListGatewayResponse{},
		*/
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

func (n *NetworkServerClient) GenerateGatewayToken(ctx context.Context, in *ns.GenerateGatewayTokenRequest, opts ...grpc.CallOption) (*ns.GenerateGatewayTokenResponse, error) {
	n.GenerateGatewayTokenChan <- *in
	return &n.GenerateGatewayTokenResponse, nil
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

func (n *NetworkServerClient) CreateChannelConfiguration(ctx context.Context, in *ns.CreateChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.CreateChannelConfigurationResponse, error) {
	n.CreateChannelConfigurationChan <- *in
	return &n.CreateChannelConfigurationResponse, nil
}

func (n *NetworkServerClient) GetChannelConfiguration(ctx context.Context, in *ns.GetChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.GetChannelConfigurationResponse, error) {
	n.GetChannelConfigurationChan <- *in
	return &n.GetChannelConfigurationResponse, nil
}

func (n *NetworkServerClient) UpdateChannelConfiguration(ctx context.Context, in *ns.UpdateChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.UpdateChannelConfigurationResponse, error) {
	n.UpdateChannelConfigurationChan <- *in
	return &n.UpdateChannelConfigurationResponse, nil
}

func (n *NetworkServerClient) DeleteChannelConfiguration(ctx context.Context, in *ns.DeleteChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.DeleteChannelConfigurationResponse, error) {
	n.DeleteChannelConfigurationChan <- *in
	return &n.DeleteChannelConfigurationResponse, nil
}

func (n *NetworkServerClient) ListChannelConfigurations(ctx context.Context, in *ns.ListChannelConfigurationsRequest, opts ...grpc.CallOption) (*ns.ListChannelConfigurationsResponse, error) {
	n.ListChannelConfigurationsChan <- *in
	return &n.ListChannelConfigurationsResponse, nil
}

func (n *NetworkServerClient) CreateExtraChannel(ctx context.Context, in *ns.CreateExtraChannelRequest, opts ...grpc.CallOption) (*ns.CreateExtraChannelResponse, error) {
	n.CreateExtraChannelChan <- *in
	return &n.CreateExtraChannelResponse, nil
}

func (n *NetworkServerClient) UpdateExtraChannel(ctx context.Context, in *ns.UpdateExtraChannelRequest, opts ...grpc.CallOption) (*ns.UpdateExtraChannelResponse, error) {
	n.UpdateExtraChannelChan <- *in
	return &n.UpdateExtraChannelResponse, nil
}

func (n *NetworkServerClient) DeleteExtraChannel(ctx context.Context, in *ns.DeleteExtraChannelRequest, opts ...grpc.CallOption) (*ns.DeleteExtraChannelResponse, error) {
	n.DeleteExtraChannelChan <- *in
	return &n.DeleteExtraChannelResponse, nil
}

func (n *NetworkServerClient) GetExtraChannelsForChannelConfigurationID(ctx context.Context, in *ns.GetExtraChannelsForChannelConfigurationIDRequest, opts ...grpc.CallOption) (*ns.GetExtraChannelsForChannelConfigurationIDResponse, error) {
	n.GetExtraChannelsForChannelConfigurationIDChan <- *in
	return &n.GetExtraChannelsForChannelConfigurationIDResponse, nil
}
