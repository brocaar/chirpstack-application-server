package test

import (
	"os"

	"github.com/garyburd/redigo/redis"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/gusseleet/lora-app-server/internal/common"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/migrations"
	"github.com/gusseleet/lora-app-server/internal/nsclient"
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

func init() {
	config.C.ApplicationServer.ID = "6d5db27e-4ce2-4b2b-b5d7-91f069397978"
	config.C.ApplicationServer.API.PublicHost = "localhost:8001"
}

// GetConfig returns the test configuration.
func GetConfig() *Config {
	log.SetLevel(log.ErrorLevel)

	c := &Config{
		PostgresDSN: "postgres://localhost/loraserver_as_test?sslmode=disable",
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
func MustResetDB(db *common.DBLogger) {
	m := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "",
	}
	if _, err := migrate.Exec(db.DB.DB, "postgres", m, migrate.Down); err != nil {
		log.Fatal(err)
	}
	if _, err := migrate.Exec(db.DB.DB, "postgres", m, migrate.Up); err != nil {
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

// NetworkServerPool is a network-server pool for testing.
type NetworkServerPool struct {
	Client      ns.NetworkServerClient
	GetHostname string
}

// Get returns the Client.
func (p *NetworkServerPool) Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerClient, error) {
	p.GetHostname = hostname
	return p.Client, nil
}

// NewNetworkServerPool creates a network-server client pool which always
// returns the given client on Get.
func NewNetworkServerPool(client *NetworkServerClient) nsclient.Pool {
	return &NetworkServerPool{
		Client: client,
	}
}

// NetworkServerClient is a test network-server client.
type NetworkServerClient struct {
	CreateServiceProfileChan     chan ns.CreateServiceProfileRequest
	CreateServiceProfileResponse ns.CreateServiceProfileResponse

	GetServiceProfileChan     chan ns.GetServiceProfileRequest
	GetServiceProfileResponse ns.GetServiceProfileResponse

	UpdateServiceProfileChan     chan ns.UpdateServiceProfileRequest
	UpdateServiceProfileResponse ns.UpdateServiceProfileResponse

	DeleteServiceProfileChan     chan ns.DeleteServiceProfileRequest
	DeleteServiceProfileResponse ns.DeleteServiceProfileResponse

	CreateRoutingProfileChan     chan ns.CreateRoutingProfileRequest
	CreateRoutingProfileResponse ns.CreateRoutingProfileResponse

	GetRoutingProfileChan     chan ns.GetRoutingProfileRequest
	GetRoutingProfileResponse ns.GetRoutingProfileResponse

	UpdateRoutingProfileChan     chan ns.UpdateRoutingProfileRequest
	UpdateRoutingProfileResponse ns.UpdateRoutingProfileResponse

	DeleteRoutingProfileChan     chan ns.DeleteRoutingProfileRequest
	DeleteRoutingProfileResponse ns.DeleteRoutingProfileResponse

	CreateDeviceProfileChan     chan ns.CreateDeviceProfileRequest
	CreateDeviceProfileResponse ns.CreateDeviceProfileResponse

	GetDeviceProfileChan     chan ns.GetDeviceProfileRequest
	GetDeviceProfileResponse ns.GetDeviceProfileResponse

	UpdateDeviceProfileChan     chan ns.UpdateDeviceProfileRequest
	UpdateDeviceProfileResponse ns.UpdateDeviceProfileResponse

	DeleteDeviceProfileChan     chan ns.DeleteDeviceProfileRequest
	DeleteDeviceProfileResponse ns.DeleteDeviceProfileResponse

	CreateDeviceChan     chan ns.CreateDeviceRequest
	CreateDeviceResponse ns.CreateDeviceResponse

	GetDeviceChan     chan ns.GetDeviceRequest
	GetDeviceResponse ns.GetDeviceResponse

	UpdateDeviceChan     chan ns.UpdateDeviceRequest
	UpdateDeviceResponse ns.UpdateDeviceResponse

	DeleteDeviceChan     chan ns.DeleteDeviceRequest
	DeleteDeviceResponse ns.DeleteDeviceResponse

	ActivateDeviceChan     chan ns.ActivateDeviceRequest
	ActivateDeviceResponse ns.ActivateDeviceResponse

	DeactivateDeviceChan     chan ns.DeactivateDeviceRequest
	DeactivateDeviceResponse ns.DeactivateDeviceResponse

	GetDeviceActivationChan     chan ns.GetDeviceActivationRequest
	GetDeviceActivationResponse ns.GetDeviceActivationResponse

	GetRandomDevAddrChan     chan ns.GetRandomDevAddrRequest
	GetRandomDevAddrResponse ns.GetRandomDevAddrResponse

	EnqueueDownlinkMACCommandChan     chan ns.EnqueueDownlinkMACCommandRequest
	EnqueueDownlinkMACCommandResponse ns.EnqueueDownlinkMACCommandResponse

	SendProprietaryPayloadChan     chan ns.SendProprietaryPayloadRequest
	SendProprietaryPayloadResponse ns.SendProprietaryPayloadResponse

	CreateGatewayChan     chan ns.CreateGatewayRequest
	CreateGatewayResponse ns.CreateGatewayResponse

	GetGatewayChan     chan ns.GetGatewayRequest
	GetGatewayResponse ns.GetGatewayResponse

	UpdateGatewayChan     chan ns.UpdateGatewayRequest
	UpdateGatewayResponse ns.UpdateGatewayResponse

	ListGatewayChan     chan ns.ListGatewayRequest
	ListGatewayResponse ns.ListGatewayResponse

	DeleteGatewayChan     chan ns.DeleteGatewayRequest
	DeleteGatewayResponse ns.DeleteGatewayResponse

	GenerateGatewayTokenChan     chan ns.GenerateGatewayTokenRequest
	GenerateGatewayTokenResponse ns.GenerateGatewayTokenResponse

	GetGatewayStatsChan     chan ns.GetGatewayStatsRequest
	GetGatewayStatsResponse ns.GetGatewayStatsResponse

	CreateChannelConfigurationChan     chan ns.CreateChannelConfigurationRequest
	CreateChannelConfigurationResponse ns.CreateChannelConfigurationResponse

	GetChannelConfigurationChan     chan ns.GetChannelConfigurationRequest
	GetChannelConfigurationResponse ns.GetChannelConfigurationResponse

	UpdateChannelConfigurationChan     chan ns.UpdateChannelConfigurationRequest
	UpdateChannelConfigurationResponse ns.UpdateChannelConfigurationResponse

	DeleteChannelConfigurationChan     chan ns.DeleteChannelConfigurationRequest
	DeleteChannelConfigurationResponse ns.DeleteChannelConfigurationResponse

	ListChannelConfigurationsChan     chan ns.ListChannelConfigurationsRequest
	ListChannelConfigurationsResponse ns.ListChannelConfigurationsResponse

	CreateExtraChannelChan     chan ns.CreateExtraChannelRequest
	CreateExtraChannelResponse ns.CreateExtraChannelResponse

	UpdateExtraChannelChan     chan ns.UpdateExtraChannelRequest
	UpdateExtraChannelResponse ns.UpdateExtraChannelResponse

	DeleteExtraChannelChan     chan ns.DeleteExtraChannelRequest
	DeleteExtraChannelResponse ns.DeleteExtraChannelResponse

	GetExtraChannelsForChannelConfigurationIDChan     chan ns.GetExtraChannelsForChannelConfigurationIDRequest
	GetExtraChannelsForChannelConfigurationIDResponse ns.GetExtraChannelsForChannelConfigurationIDResponse

	CreateDeviceQueueItemChan     chan ns.CreateDeviceQueueItemRequest
	CreateDeviceQueueItemResponse ns.CreateDeviceQueueItemResponse

	FlushDeviceQueueForDevEUIChan     chan ns.FlushDeviceQueueForDevEUIRequest
	FlushDeviceQueueForDevEUIResponse ns.FlushDeviceQueueForDevEUIResponse

	GetDeviceQueueItemsForDevEUIChan     chan ns.GetDeviceQueueItemsForDevEUIRequest
	GetDeviceQueueItemsForDevEUIResponse ns.GetDeviceQueueItemsForDevEUIResponse

	GetNextDownlinkFCntForDevEUIChan     chan ns.GetNextDownlinkFCntForDevEUIRequest
	GetNextDownlinkFCntForDevEUIResponse ns.GetNextDownlinkFCntForDevEUIResponse
}

// NewNetworkServerClient creates a new NetworkServerClient.
func NewNetworkServerClient() *NetworkServerClient {
	return &NetworkServerClient{
		CreateServiceProfileChan:                      make(chan ns.CreateServiceProfileRequest, 100),
		GetServiceProfileChan:                         make(chan ns.GetServiceProfileRequest, 100),
		UpdateServiceProfileChan:                      make(chan ns.UpdateServiceProfileRequest, 100),
		DeleteServiceProfileChan:                      make(chan ns.DeleteServiceProfileRequest, 100),
		CreateRoutingProfileChan:                      make(chan ns.CreateRoutingProfileRequest, 100),
		GetRoutingProfileChan:                         make(chan ns.GetRoutingProfileRequest, 100),
		UpdateRoutingProfileChan:                      make(chan ns.UpdateRoutingProfileRequest, 100),
		DeleteRoutingProfileChan:                      make(chan ns.DeleteRoutingProfileRequest, 100),
		CreateDeviceProfileChan:                       make(chan ns.CreateDeviceProfileRequest, 100),
		GetDeviceProfileChan:                          make(chan ns.GetDeviceProfileRequest, 100),
		UpdateDeviceProfileChan:                       make(chan ns.UpdateDeviceProfileRequest, 100),
		DeleteDeviceProfileChan:                       make(chan ns.DeleteDeviceProfileRequest, 100),
		CreateDeviceChan:                              make(chan ns.CreateDeviceRequest, 100),
		GetDeviceChan:                                 make(chan ns.GetDeviceRequest, 100),
		UpdateDeviceChan:                              make(chan ns.UpdateDeviceRequest, 100),
		DeleteDeviceChan:                              make(chan ns.DeleteDeviceRequest, 100),
		ActivateDeviceChan:                            make(chan ns.ActivateDeviceRequest, 100),
		DeactivateDeviceChan:                          make(chan ns.DeactivateDeviceRequest, 100),
		GetDeviceActivationChan:                       make(chan ns.GetDeviceActivationRequest, 100),
		GetRandomDevAddrChan:                          make(chan ns.GetRandomDevAddrRequest, 100),
		EnqueueDownlinkMACCommandChan:                 make(chan ns.EnqueueDownlinkMACCommandRequest, 100),
		SendProprietaryPayloadChan:                    make(chan ns.SendProprietaryPayloadRequest, 100),
		CreateGatewayChan:                             make(chan ns.CreateGatewayRequest, 100),
		GetGatewayChan:                                make(chan ns.GetGatewayRequest, 100),
		UpdateGatewayChan:                             make(chan ns.UpdateGatewayRequest, 100),
		ListGatewayChan:                               make(chan ns.ListGatewayRequest, 100),
		DeleteGatewayChan:                             make(chan ns.DeleteGatewayRequest, 100),
		GenerateGatewayTokenChan:                      make(chan ns.GenerateGatewayTokenRequest, 100),
		GetGatewayStatsChan:                           make(chan ns.GetGatewayStatsRequest, 100),
		CreateChannelConfigurationChan:                make(chan ns.CreateChannelConfigurationRequest, 100),
		GetChannelConfigurationChan:                   make(chan ns.GetChannelConfigurationRequest, 100),
		UpdateChannelConfigurationChan:                make(chan ns.UpdateChannelConfigurationRequest, 100),
		DeleteChannelConfigurationChan:                make(chan ns.DeleteChannelConfigurationRequest, 100),
		ListChannelConfigurationsChan:                 make(chan ns.ListChannelConfigurationsRequest, 100),
		CreateExtraChannelChan:                        make(chan ns.CreateExtraChannelRequest, 100),
		UpdateExtraChannelChan:                        make(chan ns.UpdateExtraChannelRequest, 100),
		DeleteExtraChannelChan:                        make(chan ns.DeleteExtraChannelRequest, 100),
		GetExtraChannelsForChannelConfigurationIDChan: make(chan ns.GetExtraChannelsForChannelConfigurationIDRequest, 100),
		GetNextDownlinkFCntForDevEUIChan:              make(chan ns.GetNextDownlinkFCntForDevEUIRequest, 100),
		CreateDeviceQueueItemChan:                     make(chan ns.CreateDeviceQueueItemRequest, 100),
		FlushDeviceQueueForDevEUIChan:                 make(chan ns.FlushDeviceQueueForDevEUIRequest, 100),
		GetDeviceQueueItemsForDevEUIChan:              make(chan ns.GetDeviceQueueItemsForDevEUIRequest, 100),
	}
}

// CreateServiceProfile method.
func (n *NetworkServerClient) CreateServiceProfile(ctx context.Context, in *ns.CreateServiceProfileRequest, opts ...grpc.CallOption) (*ns.CreateServiceProfileResponse, error) {
	n.CreateServiceProfileChan <- *in
	return &n.CreateServiceProfileResponse, nil
}

// GetServiceProfile method.
func (n *NetworkServerClient) GetServiceProfile(ctx context.Context, in *ns.GetServiceProfileRequest, opts ...grpc.CallOption) (*ns.GetServiceProfileResponse, error) {
	n.GetServiceProfileChan <- *in
	return &n.GetServiceProfileResponse, nil
}

// UpdateServiceProfile method.
func (n *NetworkServerClient) UpdateServiceProfile(ctx context.Context, in *ns.UpdateServiceProfileRequest, opts ...grpc.CallOption) (*ns.UpdateServiceProfileResponse, error) {
	n.UpdateServiceProfileChan <- *in
	return &n.UpdateServiceProfileResponse, nil
}

// DeleteServiceProfile method.
func (n *NetworkServerClient) DeleteServiceProfile(ctx context.Context, in *ns.DeleteServiceProfileRequest, opts ...grpc.CallOption) (*ns.DeleteServiceProfileResponse, error) {
	n.DeleteServiceProfileChan <- *in
	return &n.DeleteServiceProfileResponse, nil
}

// CreateRoutingProfile method.
func (n *NetworkServerClient) CreateRoutingProfile(ctx context.Context, in *ns.CreateRoutingProfileRequest, opts ...grpc.CallOption) (*ns.CreateRoutingProfileResponse, error) {
	n.CreateRoutingProfileChan <- *in
	return &n.CreateRoutingProfileResponse, nil
}

// GetRoutingProfile method.
func (n *NetworkServerClient) GetRoutingProfile(ctx context.Context, in *ns.GetRoutingProfileRequest, opts ...grpc.CallOption) (*ns.GetRoutingProfileResponse, error) {
	n.GetRoutingProfileChan <- *in
	return &n.GetRoutingProfileResponse, nil
}

// UpdateRoutingProfile method.
func (n *NetworkServerClient) UpdateRoutingProfile(ctx context.Context, in *ns.UpdateRoutingProfileRequest, opts ...grpc.CallOption) (*ns.UpdateRoutingProfileResponse, error) {
	n.UpdateRoutingProfileChan <- *in
	return &n.UpdateRoutingProfileResponse, nil
}

// DeleteRoutingProfile method.
func (n *NetworkServerClient) DeleteRoutingProfile(ctx context.Context, in *ns.DeleteRoutingProfileRequest, opts ...grpc.CallOption) (*ns.DeleteRoutingProfileResponse, error) {
	n.DeleteRoutingProfileChan <- *in
	return &n.DeleteRoutingProfileResponse, nil
}

// CreateDeviceProfile method.
func (n *NetworkServerClient) CreateDeviceProfile(ctx context.Context, in *ns.CreateDeviceProfileRequest, opts ...grpc.CallOption) (*ns.CreateDeviceProfileResponse, error) {
	n.CreateDeviceProfileChan <- *in
	return &n.CreateDeviceProfileResponse, nil
}

// GetDeviceProfile method.
func (n *NetworkServerClient) GetDeviceProfile(ctx context.Context, in *ns.GetDeviceProfileRequest, opts ...grpc.CallOption) (*ns.GetDeviceProfileResponse, error) {
	n.GetDeviceProfileChan <- *in
	return &n.GetDeviceProfileResponse, nil
}

// UpdateDeviceProfile method.
func (n *NetworkServerClient) UpdateDeviceProfile(ctx context.Context, in *ns.UpdateDeviceProfileRequest, opts ...grpc.CallOption) (*ns.UpdateDeviceProfileResponse, error) {
	n.UpdateDeviceProfileChan <- *in
	return &n.UpdateDeviceProfileResponse, nil
}

// DeleteDeviceProfile method.
func (n *NetworkServerClient) DeleteDeviceProfile(ctx context.Context, in *ns.DeleteDeviceProfileRequest, opts ...grpc.CallOption) (*ns.DeleteDeviceProfileResponse, error) {
	n.DeleteDeviceProfileChan <- *in
	return &n.DeleteDeviceProfileResponse, nil
}

// CreateDevice method.
func (n *NetworkServerClient) CreateDevice(ctx context.Context, in *ns.CreateDeviceRequest, opts ...grpc.CallOption) (*ns.CreateDeviceResponse, error) {
	n.CreateDeviceChan <- *in
	return &n.CreateDeviceResponse, nil
}

// GetDevice method.
func (n *NetworkServerClient) GetDevice(ctx context.Context, in *ns.GetDeviceRequest, opts ...grpc.CallOption) (*ns.GetDeviceResponse, error) {
	n.GetDeviceChan <- *in
	return &n.GetDeviceResponse, nil
}

// UpdateDevice method.
func (n *NetworkServerClient) UpdateDevice(ctx context.Context, in *ns.UpdateDeviceRequest, opts ...grpc.CallOption) (*ns.UpdateDeviceResponse, error) {
	n.UpdateDeviceChan <- *in
	return &n.UpdateDeviceResponse, nil
}

// DeleteDevice method.
func (n *NetworkServerClient) DeleteDevice(ctx context.Context, in *ns.DeleteDeviceRequest, opts ...grpc.CallOption) (*ns.DeleteDeviceResponse, error) {
	n.DeleteDeviceChan <- *in
	return &n.DeleteDeviceResponse, nil
}

// ActivateDevice method.
func (n *NetworkServerClient) ActivateDevice(ctx context.Context, in *ns.ActivateDeviceRequest, opts ...grpc.CallOption) (*ns.ActivateDeviceResponse, error) {
	n.ActivateDeviceChan <- *in
	return &n.ActivateDeviceResponse, nil
}

// GetDeviceActivation method.
func (n *NetworkServerClient) GetDeviceActivation(ctx context.Context, in *ns.GetDeviceActivationRequest, opts ...grpc.CallOption) (*ns.GetDeviceActivationResponse, error) {
	n.GetDeviceActivationChan <- *in
	return &n.GetDeviceActivationResponse, nil
}

// DeactivateDevice method.
func (n *NetworkServerClient) DeactivateDevice(ctx context.Context, in *ns.DeactivateDeviceRequest, opts ...grpc.CallOption) (*ns.DeactivateDeviceResponse, error) {
	n.DeactivateDeviceChan <- *in
	return &n.DeactivateDeviceResponse, nil
}

// CreateGateway method.
func (n *NetworkServerClient) CreateGateway(ctx context.Context, in *ns.CreateGatewayRequest, opts ...grpc.CallOption) (*ns.CreateGatewayResponse, error) {
	n.CreateGatewayChan <- *in
	return &n.CreateGatewayResponse, nil
}

// GetGateway method.
func (n *NetworkServerClient) GetGateway(ctx context.Context, in *ns.GetGatewayRequest, opts ...grpc.CallOption) (*ns.GetGatewayResponse, error) {
	n.GetGatewayChan <- *in
	return &n.GetGatewayResponse, nil
}

// ListGateways method,
func (n *NetworkServerClient) ListGateways(ctx context.Context, in *ns.ListGatewayRequest, opts ...grpc.CallOption) (*ns.ListGatewayResponse, error) {
	n.ListGatewayChan <- *in
	return &n.ListGatewayResponse, nil
}

// UpdateGateway method.
func (n *NetworkServerClient) UpdateGateway(ctx context.Context, in *ns.UpdateGatewayRequest, opts ...grpc.CallOption) (*ns.UpdateGatewayResponse, error) {
	n.UpdateGatewayChan <- *in
	return &n.UpdateGatewayResponse, nil
}

// DeleteGateway method.
func (n *NetworkServerClient) DeleteGateway(ctx context.Context, in *ns.DeleteGatewayRequest, opts ...grpc.CallOption) (*ns.DeleteGatewayResponse, error) {
	n.DeleteGatewayChan <- *in
	return &n.DeleteGatewayResponse, nil
}

// GenerateGatewayToken method.
func (n *NetworkServerClient) GenerateGatewayToken(ctx context.Context, in *ns.GenerateGatewayTokenRequest, opts ...grpc.CallOption) (*ns.GenerateGatewayTokenResponse, error) {
	n.GenerateGatewayTokenChan <- *in
	return &n.GenerateGatewayTokenResponse, nil
}

// GetGatewayStats method.
func (n *NetworkServerClient) GetGatewayStats(ctx context.Context, in *ns.GetGatewayStatsRequest, opts ...grpc.CallOption) (*ns.GetGatewayStatsResponse, error) {
	n.GetGatewayStatsChan <- *in
	return &n.GetGatewayStatsResponse, nil
}

// GetRandomDevAddr method.
func (n *NetworkServerClient) GetRandomDevAddr(ctx context.Context, in *ns.GetRandomDevAddrRequest, opts ...grpc.CallOption) (*ns.GetRandomDevAddrResponse, error) {
	n.GetRandomDevAddrChan <- *in
	return &n.GetRandomDevAddrResponse, nil
}

// EnqueueDownlinkMACCommand method.
func (n *NetworkServerClient) EnqueueDownlinkMACCommand(ctx context.Context, in *ns.EnqueueDownlinkMACCommandRequest, opts ...grpc.CallOption) (*ns.EnqueueDownlinkMACCommandResponse, error) {
	n.EnqueueDownlinkMACCommandChan <- *in
	return &n.EnqueueDownlinkMACCommandResponse, nil
}

// SendProprietaryPayload method.
func (n *NetworkServerClient) SendProprietaryPayload(ctx context.Context, in *ns.SendProprietaryPayloadRequest, opts ...grpc.CallOption) (*ns.SendProprietaryPayloadResponse, error) {
	n.SendProprietaryPayloadChan <- *in
	return &n.SendProprietaryPayloadResponse, nil
}

// CreateChannelConfiguration method.
func (n *NetworkServerClient) CreateChannelConfiguration(ctx context.Context, in *ns.CreateChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.CreateChannelConfigurationResponse, error) {
	n.CreateChannelConfigurationChan <- *in
	return &n.CreateChannelConfigurationResponse, nil
}

// GetChannelConfiguration method.
func (n *NetworkServerClient) GetChannelConfiguration(ctx context.Context, in *ns.GetChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.GetChannelConfigurationResponse, error) {
	n.GetChannelConfigurationChan <- *in
	return &n.GetChannelConfigurationResponse, nil
}

// UpdateChannelConfiguration method.
func (n *NetworkServerClient) UpdateChannelConfiguration(ctx context.Context, in *ns.UpdateChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.UpdateChannelConfigurationResponse, error) {
	n.UpdateChannelConfigurationChan <- *in
	return &n.UpdateChannelConfigurationResponse, nil
}

// DeleteChannelConfiguration method.
func (n *NetworkServerClient) DeleteChannelConfiguration(ctx context.Context, in *ns.DeleteChannelConfigurationRequest, opts ...grpc.CallOption) (*ns.DeleteChannelConfigurationResponse, error) {
	n.DeleteChannelConfigurationChan <- *in
	return &n.DeleteChannelConfigurationResponse, nil
}

// ListChannelConfigurations method.
func (n *NetworkServerClient) ListChannelConfigurations(ctx context.Context, in *ns.ListChannelConfigurationsRequest, opts ...grpc.CallOption) (*ns.ListChannelConfigurationsResponse, error) {
	n.ListChannelConfigurationsChan <- *in
	return &n.ListChannelConfigurationsResponse, nil
}

// CreateExtraChannel method.
func (n *NetworkServerClient) CreateExtraChannel(ctx context.Context, in *ns.CreateExtraChannelRequest, opts ...grpc.CallOption) (*ns.CreateExtraChannelResponse, error) {
	n.CreateExtraChannelChan <- *in
	return &n.CreateExtraChannelResponse, nil
}

// UpdateExtraChannel method.
func (n *NetworkServerClient) UpdateExtraChannel(ctx context.Context, in *ns.UpdateExtraChannelRequest, opts ...grpc.CallOption) (*ns.UpdateExtraChannelResponse, error) {
	n.UpdateExtraChannelChan <- *in
	return &n.UpdateExtraChannelResponse, nil
}

// DeleteExtraChannel method.
func (n *NetworkServerClient) DeleteExtraChannel(ctx context.Context, in *ns.DeleteExtraChannelRequest, opts ...grpc.CallOption) (*ns.DeleteExtraChannelResponse, error) {
	n.DeleteExtraChannelChan <- *in
	return &n.DeleteExtraChannelResponse, nil
}

// GetExtraChannelsForChannelConfigurationID method.
func (n *NetworkServerClient) GetExtraChannelsForChannelConfigurationID(ctx context.Context, in *ns.GetExtraChannelsForChannelConfigurationIDRequest, opts ...grpc.CallOption) (*ns.GetExtraChannelsForChannelConfigurationIDResponse, error) {
	n.GetExtraChannelsForChannelConfigurationIDChan <- *in
	return &n.GetExtraChannelsForChannelConfigurationIDResponse, nil
}

// MigrateNodeToDeviceSession is not implemented.
func (n *NetworkServerClient) MigrateNodeToDeviceSession(ctx context.Context, in *ns.MigrateNodeToDeviceSessionRequest, opts ...grpc.CallOption) (*ns.MigrateNodeToDeviceSessionResponse, error) {
	panic("not implemented")
}

// CreateDeviceQueueItem method.
func (n NetworkServerClient) CreateDeviceQueueItem(ctx context.Context, in *ns.CreateDeviceQueueItemRequest, opts ...grpc.CallOption) (*ns.CreateDeviceQueueItemResponse, error) {
	n.CreateDeviceQueueItemChan <- *in
	return &n.CreateDeviceQueueItemResponse, nil
}

// FlushDeviceQueueForDevEUI method.
func (n NetworkServerClient) FlushDeviceQueueForDevEUI(ctx context.Context, in *ns.FlushDeviceQueueForDevEUIRequest, opts ...grpc.CallOption) (*ns.FlushDeviceQueueForDevEUIResponse, error) {
	n.FlushDeviceQueueForDevEUIChan <- *in
	return &n.FlushDeviceQueueForDevEUIResponse, nil
}

// GetDeviceQueueItemsForDevEUI method.
func (n NetworkServerClient) GetDeviceQueueItemsForDevEUI(ctx context.Context, in *ns.GetDeviceQueueItemsForDevEUIRequest, opts ...grpc.CallOption) (*ns.GetDeviceQueueItemsForDevEUIResponse, error) {
	n.GetDeviceQueueItemsForDevEUIChan <- *in
	return &n.GetDeviceQueueItemsForDevEUIResponse, nil
}

// GetNextDownlinkFCntForDevEUI method.
func (n NetworkServerClient) GetNextDownlinkFCntForDevEUI(ctx context.Context, in *ns.GetNextDownlinkFCntForDevEUIRequest, opts ...grpc.CallOption) (*ns.GetNextDownlinkFCntForDevEUIResponse, error) {
	n.GetNextDownlinkFCntForDevEUIChan <- *in
	return &n.GetNextDownlinkFCntForDevEUIResponse, nil
}

// StreamFrameLogsForGateway method.
func (n NetworkServerClient) StreamFrameLogsForGateway(ctx context.Context, in *ns.StreamFrameLogsForGatewayRequest, opts ...grpc.CallOption) (ns.NetworkServer_StreamFrameLogsForGatewayClient, error) {
	panic("not implemented")
}

// StreamFrameLogsForDevice method.
func (n NetworkServerClient) StreamFrameLogsForDevice(ctx context.Context, in *ns.StreamFrameLogsForDeviceRequest, opts ...grpc.CallOption) (ns.NetworkServer_StreamFrameLogsForDeviceClient, error) {
	panic("not implemented")
}
