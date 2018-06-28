package test

import (
	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/ptypes/empty"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/lora-app-server/internal/nsclient"
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
	Client      ns.NetworkServerServiceClient
	GetHostname string
}

// Get returns the Client.
func (p *NetworkServerPool) Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerServiceClient, error) {
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
	UpdateServiceProfileResponse empty.Empty

	DeleteServiceProfileChan     chan ns.DeleteServiceProfileRequest
	DeleteServiceProfileResponse empty.Empty

	CreateRoutingProfileChan     chan ns.CreateRoutingProfileRequest
	CreateRoutingProfileResponse ns.CreateRoutingProfileResponse

	GetRoutingProfileChan     chan ns.GetRoutingProfileRequest
	GetRoutingProfileResponse ns.GetRoutingProfileResponse

	UpdateRoutingProfileChan     chan ns.UpdateRoutingProfileRequest
	UpdateRoutingProfileResponse empty.Empty

	DeleteRoutingProfileChan     chan ns.DeleteRoutingProfileRequest
	DeleteRoutingProfileResponse empty.Empty

	CreateDeviceProfileChan     chan ns.CreateDeviceProfileRequest
	CreateDeviceProfileResponse ns.CreateDeviceProfileResponse

	GetDeviceProfileChan     chan ns.GetDeviceProfileRequest
	GetDeviceProfileResponse ns.GetDeviceProfileResponse

	UpdateDeviceProfileChan     chan ns.UpdateDeviceProfileRequest
	UpdateDeviceProfileResponse empty.Empty

	DeleteDeviceProfileChan     chan ns.DeleteDeviceProfileRequest
	DeleteDeviceProfileResponse empty.Empty

	CreateDeviceChan     chan ns.CreateDeviceRequest
	CreateDeviceResponse empty.Empty

	GetDeviceChan     chan ns.GetDeviceRequest
	GetDeviceResponse ns.GetDeviceResponse

	UpdateDeviceChan     chan ns.UpdateDeviceRequest
	UpdateDeviceResponse empty.Empty

	DeleteDeviceChan     chan ns.DeleteDeviceRequest
	DeleteDeviceResponse empty.Empty

	ActivateDeviceChan     chan ns.ActivateDeviceRequest
	ActivateDeviceResponse empty.Empty

	DeactivateDeviceChan     chan ns.DeactivateDeviceRequest
	DeactivateDeviceResponse empty.Empty

	GetDeviceActivationChan     chan ns.GetDeviceActivationRequest
	GetDeviceActivationResponse ns.GetDeviceActivationResponse

	GetRandomDevAddrChan     chan empty.Empty
	GetRandomDevAddrResponse ns.GetRandomDevAddrResponse

	CreateMACCommandQueueItemChan     chan ns.CreateMACCommandQueueItemRequest
	CreateMACCommandQueueItemResponse empty.Empty

	SendProprietaryPayloadChan     chan ns.SendProprietaryPayloadRequest
	SendProprietaryPayloadResponse empty.Empty

	CreateGatewayChan     chan ns.CreateGatewayRequest
	CreateGatewayResponse empty.Empty

	GetGatewayChan     chan ns.GetGatewayRequest
	GetGatewayResponse ns.GetGatewayResponse

	UpdateGatewayChan     chan ns.UpdateGatewayRequest
	UpdateGatewayResponse empty.Empty

	DeleteGatewayChan     chan ns.DeleteGatewayRequest
	DeleteGatewayResponse empty.Empty

	GetGatewayStatsChan     chan ns.GetGatewayStatsRequest
	GetGatewayStatsResponse ns.GetGatewayStatsResponse

	CreateGatewayProfileChan     chan ns.CreateGatewayProfileRequest
	CreateGatewayProfileResponse ns.CreateGatewayProfileResponse

	GetGatewayProfileChan     chan ns.GetGatewayProfileRequest
	GetGatewayProfileResponse ns.GetGatewayProfileResponse

	UpdateGatewayProfileChan     chan ns.UpdateGatewayProfileRequest
	UpdateGatewayProfileResponse empty.Empty

	DeleteGatewayProfileChan     chan ns.DeleteGatewayProfileRequest
	DeleteGatewayProfileResponse empty.Empty

	CreateDeviceQueueItemChan     chan ns.CreateDeviceQueueItemRequest
	CreateDeviceQueueItemResponse empty.Empty

	FlushDeviceQueueForDevEUIChan     chan ns.FlushDeviceQueueForDevEUIRequest
	FlushDeviceQueueForDevEUIResponse empty.Empty

	GetDeviceQueueItemsForDevEUIChan     chan ns.GetDeviceQueueItemsForDevEUIRequest
	GetDeviceQueueItemsForDevEUIResponse ns.GetDeviceQueueItemsForDevEUIResponse

	GetNextDownlinkFCntForDevEUIChan     chan ns.GetNextDownlinkFCntForDevEUIRequest
	GetNextDownlinkFCntForDevEUIResponse ns.GetNextDownlinkFCntForDevEUIResponse

	GetVersionResponse ns.GetVersionResponse
}

// NewNetworkServerClient creates a new NetworkServerClient.
func NewNetworkServerClient() *NetworkServerClient {
	return &NetworkServerClient{
		CreateServiceProfileChan:         make(chan ns.CreateServiceProfileRequest, 100),
		GetServiceProfileChan:            make(chan ns.GetServiceProfileRequest, 100),
		UpdateServiceProfileChan:         make(chan ns.UpdateServiceProfileRequest, 100),
		DeleteServiceProfileChan:         make(chan ns.DeleteServiceProfileRequest, 100),
		CreateRoutingProfileChan:         make(chan ns.CreateRoutingProfileRequest, 100),
		GetRoutingProfileChan:            make(chan ns.GetRoutingProfileRequest, 100),
		UpdateRoutingProfileChan:         make(chan ns.UpdateRoutingProfileRequest, 100),
		DeleteRoutingProfileChan:         make(chan ns.DeleteRoutingProfileRequest, 100),
		CreateDeviceProfileChan:          make(chan ns.CreateDeviceProfileRequest, 100),
		GetDeviceProfileChan:             make(chan ns.GetDeviceProfileRequest, 100),
		UpdateDeviceProfileChan:          make(chan ns.UpdateDeviceProfileRequest, 100),
		DeleteDeviceProfileChan:          make(chan ns.DeleteDeviceProfileRequest, 100),
		CreateDeviceChan:                 make(chan ns.CreateDeviceRequest, 100),
		GetDeviceChan:                    make(chan ns.GetDeviceRequest, 100),
		UpdateDeviceChan:                 make(chan ns.UpdateDeviceRequest, 100),
		DeleteDeviceChan:                 make(chan ns.DeleteDeviceRequest, 100),
		ActivateDeviceChan:               make(chan ns.ActivateDeviceRequest, 100),
		DeactivateDeviceChan:             make(chan ns.DeactivateDeviceRequest, 100),
		GetDeviceActivationChan:          make(chan ns.GetDeviceActivationRequest, 100),
		GetRandomDevAddrChan:             make(chan empty.Empty, 100),
		CreateMACCommandQueueItemChan:    make(chan ns.CreateMACCommandQueueItemRequest, 100),
		SendProprietaryPayloadChan:       make(chan ns.SendProprietaryPayloadRequest, 100),
		CreateGatewayChan:                make(chan ns.CreateGatewayRequest, 100),
		GetGatewayChan:                   make(chan ns.GetGatewayRequest, 100),
		UpdateGatewayChan:                make(chan ns.UpdateGatewayRequest, 100),
		DeleteGatewayChan:                make(chan ns.DeleteGatewayRequest, 100),
		GetGatewayStatsChan:              make(chan ns.GetGatewayStatsRequest, 100),
		CreateGatewayProfileChan:         make(chan ns.CreateGatewayProfileRequest, 100),
		GetGatewayProfileChan:            make(chan ns.GetGatewayProfileRequest, 100),
		UpdateGatewayProfileChan:         make(chan ns.UpdateGatewayProfileRequest, 100),
		DeleteGatewayProfileChan:         make(chan ns.DeleteGatewayProfileRequest, 100),
		GetNextDownlinkFCntForDevEUIChan: make(chan ns.GetNextDownlinkFCntForDevEUIRequest, 100),
		CreateDeviceQueueItemChan:        make(chan ns.CreateDeviceQueueItemRequest, 100),
		FlushDeviceQueueForDevEUIChan:    make(chan ns.FlushDeviceQueueForDevEUIRequest, 100),
		GetDeviceQueueItemsForDevEUIChan: make(chan ns.GetDeviceQueueItemsForDevEUIRequest, 100),
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
func (n *NetworkServerClient) UpdateServiceProfile(ctx context.Context, in *ns.UpdateServiceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateServiceProfileChan <- *in
	return &n.UpdateServiceProfileResponse, nil
}

// DeleteServiceProfile method.
func (n *NetworkServerClient) DeleteServiceProfile(ctx context.Context, in *ns.DeleteServiceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
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
func (n *NetworkServerClient) UpdateRoutingProfile(ctx context.Context, in *ns.UpdateRoutingProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateRoutingProfileChan <- *in
	return &n.UpdateRoutingProfileResponse, nil
}

// DeleteRoutingProfile method.
func (n *NetworkServerClient) DeleteRoutingProfile(ctx context.Context, in *ns.DeleteRoutingProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
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
func (n *NetworkServerClient) UpdateDeviceProfile(ctx context.Context, in *ns.UpdateDeviceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateDeviceProfileChan <- *in
	return &n.UpdateDeviceProfileResponse, nil
}

// DeleteDeviceProfile method.
func (n *NetworkServerClient) DeleteDeviceProfile(ctx context.Context, in *ns.DeleteDeviceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteDeviceProfileChan <- *in
	return &n.DeleteDeviceProfileResponse, nil
}

// CreateDevice method.
func (n *NetworkServerClient) CreateDevice(ctx context.Context, in *ns.CreateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateDeviceChan <- *in
	return &n.CreateDeviceResponse, nil
}

// GetDevice method.
func (n *NetworkServerClient) GetDevice(ctx context.Context, in *ns.GetDeviceRequest, opts ...grpc.CallOption) (*ns.GetDeviceResponse, error) {
	n.GetDeviceChan <- *in
	return &n.GetDeviceResponse, nil
}

// UpdateDevice method.
func (n *NetworkServerClient) UpdateDevice(ctx context.Context, in *ns.UpdateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateDeviceChan <- *in
	return &n.UpdateDeviceResponse, nil
}

// DeleteDevice method.
func (n *NetworkServerClient) DeleteDevice(ctx context.Context, in *ns.DeleteDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteDeviceChan <- *in
	return &n.DeleteDeviceResponse, nil
}

// ActivateDevice method.
func (n *NetworkServerClient) ActivateDevice(ctx context.Context, in *ns.ActivateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.ActivateDeviceChan <- *in
	return &n.ActivateDeviceResponse, nil
}

// GetDeviceActivation method.
func (n *NetworkServerClient) GetDeviceActivation(ctx context.Context, in *ns.GetDeviceActivationRequest, opts ...grpc.CallOption) (*ns.GetDeviceActivationResponse, error) {
	n.GetDeviceActivationChan <- *in
	return &n.GetDeviceActivationResponse, nil
}

// DeactivateDevice method.
func (n *NetworkServerClient) DeactivateDevice(ctx context.Context, in *ns.DeactivateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeactivateDeviceChan <- *in
	return &n.DeactivateDeviceResponse, nil
}

// CreateGateway method.
func (n *NetworkServerClient) CreateGateway(ctx context.Context, in *ns.CreateGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateGatewayChan <- *in
	return &n.CreateGatewayResponse, nil
}

// GetGateway method.
func (n *NetworkServerClient) GetGateway(ctx context.Context, in *ns.GetGatewayRequest, opts ...grpc.CallOption) (*ns.GetGatewayResponse, error) {
	n.GetGatewayChan <- *in
	return &n.GetGatewayResponse, nil
}

// UpdateGateway method.
func (n *NetworkServerClient) UpdateGateway(ctx context.Context, in *ns.UpdateGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateGatewayChan <- *in
	return &n.UpdateGatewayResponse, nil
}

// DeleteGateway method.
func (n *NetworkServerClient) DeleteGateway(ctx context.Context, in *ns.DeleteGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteGatewayChan <- *in
	return &n.DeleteGatewayResponse, nil
}

// GetGatewayStats method.
func (n *NetworkServerClient) GetGatewayStats(ctx context.Context, in *ns.GetGatewayStatsRequest, opts ...grpc.CallOption) (*ns.GetGatewayStatsResponse, error) {
	n.GetGatewayStatsChan <- *in
	return &n.GetGatewayStatsResponse, nil
}

// GetRandomDevAddr method.
func (n *NetworkServerClient) GetRandomDevAddr(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ns.GetRandomDevAddrResponse, error) {
	n.GetRandomDevAddrChan <- *in
	return &n.GetRandomDevAddrResponse, nil
}

// CreateMACCommandQueueItem method.
func (n *NetworkServerClient) CreateMACCommandQueueItem(ctx context.Context, in *ns.CreateMACCommandQueueItemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateMACCommandQueueItemChan <- *in
	return &n.CreateMACCommandQueueItemResponse, nil
}

// SendProprietaryPayload method.
func (n *NetworkServerClient) SendProprietaryPayload(ctx context.Context, in *ns.SendProprietaryPayloadRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.SendProprietaryPayloadChan <- *in
	return &n.SendProprietaryPayloadResponse, nil
}

// CreateGatewayProfile method.
func (n *NetworkServerClient) CreateGatewayProfile(ctx context.Context, in *ns.CreateGatewayProfileRequest, opts ...grpc.CallOption) (*ns.CreateGatewayProfileResponse, error) {
	n.CreateGatewayProfileChan <- *in
	return &n.CreateGatewayProfileResponse, nil
}

// GetGatewayProfile method.
func (n *NetworkServerClient) GetGatewayProfile(ctx context.Context, in *ns.GetGatewayProfileRequest, opts ...grpc.CallOption) (*ns.GetGatewayProfileResponse, error) {
	n.GetGatewayProfileChan <- *in
	return &n.GetGatewayProfileResponse, nil
}

// UpdateGatewayProfile method.
func (n *NetworkServerClient) UpdateGatewayProfile(ctx context.Context, in *ns.UpdateGatewayProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateGatewayProfileChan <- *in
	return &n.UpdateGatewayProfileResponse, nil
}

// DeleteGatewayProfile method.
func (n *NetworkServerClient) DeleteGatewayProfile(ctx context.Context, in *ns.DeleteGatewayProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteGatewayProfileChan <- *in
	return &n.DeleteGatewayProfileResponse, nil
}

// CreateDeviceQueueItem method.
func (n NetworkServerClient) CreateDeviceQueueItem(ctx context.Context, in *ns.CreateDeviceQueueItemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateDeviceQueueItemChan <- *in
	return &n.CreateDeviceQueueItemResponse, nil
}

// FlushDeviceQueueForDevEUI method.
func (n NetworkServerClient) FlushDeviceQueueForDevEUI(ctx context.Context, in *ns.FlushDeviceQueueForDevEUIRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
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

// GetVersion method.
func (n NetworkServerClient) GetVersion(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ns.GetVersionResponse, error) {
	return &n.GetVersionResponse, nil
}

// StreamFrameLogsForGateway method.
func (n NetworkServerClient) StreamFrameLogsForGateway(ctx context.Context, in *ns.StreamFrameLogsForGatewayRequest, opts ...grpc.CallOption) (ns.NetworkServerService_StreamFrameLogsForGatewayClient, error) {
	panic("not implemented")
}

// StreamFrameLogsForDevice method.
func (n NetworkServerClient) StreamFrameLogsForDevice(ctx context.Context, in *ns.StreamFrameLogsForDeviceRequest, opts ...grpc.CallOption) (ns.NetworkServerService_StreamFrameLogsForDeviceClient, error) {
	panic("not implemented")
}
