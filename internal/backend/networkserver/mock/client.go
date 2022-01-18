package mock

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
)

// Client is a test network-server client.
type Client struct {
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

	CreateMulticastGroupChan     chan ns.CreateMulticastGroupRequest
	CreateMulticastGroupResponse ns.CreateMulticastGroupResponse

	GetMulticastGroupChan     chan ns.GetMulticastGroupRequest
	GetMulticastGroupResponse ns.GetMulticastGroupResponse

	UpdateMulticastGroupChan     chan ns.UpdateMulticastGroupRequest
	UpdateMulticastGroupResponse empty.Empty

	DeleteMulticastGroupChan     chan ns.DeleteMulticastGroupRequest
	DeleteMulticastGroupResponse empty.Empty

	AddDeviceToMulticastGroupChan     chan ns.AddDeviceToMulticastGroupRequest
	AddDeviceToMulticastGroupResponse empty.Empty

	RemoveDeviceFromMulticastGroupChan     chan ns.RemoveDeviceFromMulticastGroupRequest
	RemoveDeviceFromMulticastGroupResponse empty.Empty

	EnqueueMulticastQueueItemChan     chan ns.EnqueueMulticastQueueItemRequest
	EnqueueMulticastQueueItemResponse empty.Empty

	FlushMulticastQueueForMulticastGroupChan     chan ns.FlushMulticastQueueForMulticastGroupRequest
	FlushMulticastQueueForMulticastGroupResponse empty.Empty

	GetMulticastQueueItemsForMulticastGroupChan     chan ns.GetMulticastQueueItemsForMulticastGroupRequest
	GetMulticastQueueItemsForMulticastGroupResponse ns.GetMulticastQueueItemsForMulticastGroupResponse

	GenerateGatewayClientCertificateChan     chan ns.GenerateGatewayClientCertificateRequest
	GenerateGatewayClientCertificateResponse ns.GenerateGatewayClientCertificateResponse

	GetVersionResponse ns.GetVersionResponse

	GetADRAlgorithmsResponse ns.GetADRAlgorithmsResponse

	ClearDeviceNoncesChan     chan ns.ClearDeviceNoncesRequest
	ClearDeviceNoncesResponse empty.Empty
}

// NewClient creates a new Client.
func NewClient() *Client {
	return &Client{
		CreateServiceProfileChan:                    make(chan ns.CreateServiceProfileRequest, 100),
		GetServiceProfileChan:                       make(chan ns.GetServiceProfileRequest, 100),
		UpdateServiceProfileChan:                    make(chan ns.UpdateServiceProfileRequest, 100),
		DeleteServiceProfileChan:                    make(chan ns.DeleteServiceProfileRequest, 100),
		CreateRoutingProfileChan:                    make(chan ns.CreateRoutingProfileRequest, 100),
		GetRoutingProfileChan:                       make(chan ns.GetRoutingProfileRequest, 100),
		UpdateRoutingProfileChan:                    make(chan ns.UpdateRoutingProfileRequest, 100),
		DeleteRoutingProfileChan:                    make(chan ns.DeleteRoutingProfileRequest, 100),
		CreateDeviceProfileChan:                     make(chan ns.CreateDeviceProfileRequest, 100),
		GetDeviceProfileChan:                        make(chan ns.GetDeviceProfileRequest, 100),
		UpdateDeviceProfileChan:                     make(chan ns.UpdateDeviceProfileRequest, 100),
		DeleteDeviceProfileChan:                     make(chan ns.DeleteDeviceProfileRequest, 100),
		CreateDeviceChan:                            make(chan ns.CreateDeviceRequest, 100),
		GetDeviceChan:                               make(chan ns.GetDeviceRequest, 100),
		UpdateDeviceChan:                            make(chan ns.UpdateDeviceRequest, 100),
		DeleteDeviceChan:                            make(chan ns.DeleteDeviceRequest, 100),
		ActivateDeviceChan:                          make(chan ns.ActivateDeviceRequest, 100),
		DeactivateDeviceChan:                        make(chan ns.DeactivateDeviceRequest, 100),
		GetDeviceActivationChan:                     make(chan ns.GetDeviceActivationRequest, 100),
		GetRandomDevAddrChan:                        make(chan empty.Empty, 100),
		CreateMACCommandQueueItemChan:               make(chan ns.CreateMACCommandQueueItemRequest, 100),
		SendProprietaryPayloadChan:                  make(chan ns.SendProprietaryPayloadRequest, 100),
		CreateGatewayChan:                           make(chan ns.CreateGatewayRequest, 100),
		GetGatewayChan:                              make(chan ns.GetGatewayRequest, 100),
		UpdateGatewayChan:                           make(chan ns.UpdateGatewayRequest, 100),
		DeleteGatewayChan:                           make(chan ns.DeleteGatewayRequest, 100),
		GetGatewayStatsChan:                         make(chan ns.GetGatewayStatsRequest, 100),
		CreateGatewayProfileChan:                    make(chan ns.CreateGatewayProfileRequest, 100),
		GetGatewayProfileChan:                       make(chan ns.GetGatewayProfileRequest, 100),
		UpdateGatewayProfileChan:                    make(chan ns.UpdateGatewayProfileRequest, 100),
		DeleteGatewayProfileChan:                    make(chan ns.DeleteGatewayProfileRequest, 100),
		GetNextDownlinkFCntForDevEUIChan:            make(chan ns.GetNextDownlinkFCntForDevEUIRequest, 100),
		CreateDeviceQueueItemChan:                   make(chan ns.CreateDeviceQueueItemRequest, 100),
		FlushDeviceQueueForDevEUIChan:               make(chan ns.FlushDeviceQueueForDevEUIRequest, 100),
		GetDeviceQueueItemsForDevEUIChan:            make(chan ns.GetDeviceQueueItemsForDevEUIRequest, 100),
		CreateMulticastGroupChan:                    make(chan ns.CreateMulticastGroupRequest, 100),
		GetMulticastGroupChan:                       make(chan ns.GetMulticastGroupRequest, 100),
		UpdateMulticastGroupChan:                    make(chan ns.UpdateMulticastGroupRequest, 100),
		DeleteMulticastGroupChan:                    make(chan ns.DeleteMulticastGroupRequest, 100),
		AddDeviceToMulticastGroupChan:               make(chan ns.AddDeviceToMulticastGroupRequest, 100),
		RemoveDeviceFromMulticastGroupChan:          make(chan ns.RemoveDeviceFromMulticastGroupRequest, 100),
		EnqueueMulticastQueueItemChan:               make(chan ns.EnqueueMulticastQueueItemRequest, 100),
		FlushMulticastQueueForMulticastGroupChan:    make(chan ns.FlushMulticastQueueForMulticastGroupRequest, 100),
		GetMulticastQueueItemsForMulticastGroupChan: make(chan ns.GetMulticastQueueItemsForMulticastGroupRequest, 100),
		GenerateGatewayClientCertificateChan:        make(chan ns.GenerateGatewayClientCertificateRequest, 100),
		ClearDeviceNoncesChan:                       make(chan ns.ClearDeviceNoncesRequest, 100),
	}
}

// CreateServiceProfile method.
func (n *Client) CreateServiceProfile(ctx context.Context, in *ns.CreateServiceProfileRequest, opts ...grpc.CallOption) (*ns.CreateServiceProfileResponse, error) {
	n.CreateServiceProfileChan <- *in
	return &n.CreateServiceProfileResponse, nil
}

// GetServiceProfile method.
func (n *Client) GetServiceProfile(ctx context.Context, in *ns.GetServiceProfileRequest, opts ...grpc.CallOption) (*ns.GetServiceProfileResponse, error) {
	n.GetServiceProfileChan <- *in
	return &n.GetServiceProfileResponse, nil
}

// UpdateServiceProfile method.
func (n *Client) UpdateServiceProfile(ctx context.Context, in *ns.UpdateServiceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateServiceProfileChan <- *in
	return &n.UpdateServiceProfileResponse, nil
}

// DeleteServiceProfile method.
func (n *Client) DeleteServiceProfile(ctx context.Context, in *ns.DeleteServiceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteServiceProfileChan <- *in
	return &n.DeleteServiceProfileResponse, nil
}

// CreateRoutingProfile method.
func (n *Client) CreateRoutingProfile(ctx context.Context, in *ns.CreateRoutingProfileRequest, opts ...grpc.CallOption) (*ns.CreateRoutingProfileResponse, error) {
	n.CreateRoutingProfileChan <- *in
	return &n.CreateRoutingProfileResponse, nil
}

// GetRoutingProfile method.
func (n *Client) GetRoutingProfile(ctx context.Context, in *ns.GetRoutingProfileRequest, opts ...grpc.CallOption) (*ns.GetRoutingProfileResponse, error) {
	n.GetRoutingProfileChan <- *in
	return &n.GetRoutingProfileResponse, nil
}

// UpdateRoutingProfile method.
func (n *Client) UpdateRoutingProfile(ctx context.Context, in *ns.UpdateRoutingProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateRoutingProfileChan <- *in
	return &n.UpdateRoutingProfileResponse, nil
}

// DeleteRoutingProfile method.
func (n *Client) DeleteRoutingProfile(ctx context.Context, in *ns.DeleteRoutingProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteRoutingProfileChan <- *in
	return &n.DeleteRoutingProfileResponse, nil
}

// CreateDeviceProfile method.
func (n *Client) CreateDeviceProfile(ctx context.Context, in *ns.CreateDeviceProfileRequest, opts ...grpc.CallOption) (*ns.CreateDeviceProfileResponse, error) {
	n.CreateDeviceProfileChan <- *in
	return &n.CreateDeviceProfileResponse, nil
}

// GetDeviceProfile method.
func (n *Client) GetDeviceProfile(ctx context.Context, in *ns.GetDeviceProfileRequest, opts ...grpc.CallOption) (*ns.GetDeviceProfileResponse, error) {
	n.GetDeviceProfileChan <- *in
	return &n.GetDeviceProfileResponse, nil
}

// UpdateDeviceProfile method.
func (n *Client) UpdateDeviceProfile(ctx context.Context, in *ns.UpdateDeviceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateDeviceProfileChan <- *in
	return &n.UpdateDeviceProfileResponse, nil
}

// DeleteDeviceProfile method.
func (n *Client) DeleteDeviceProfile(ctx context.Context, in *ns.DeleteDeviceProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteDeviceProfileChan <- *in
	return &n.DeleteDeviceProfileResponse, nil
}

// CreateDevice method.
func (n *Client) CreateDevice(ctx context.Context, in *ns.CreateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateDeviceChan <- *in
	return &n.CreateDeviceResponse, nil
}

// GetDevice method.
func (n *Client) GetDevice(ctx context.Context, in *ns.GetDeviceRequest, opts ...grpc.CallOption) (*ns.GetDeviceResponse, error) {
	n.GetDeviceChan <- *in
	return &n.GetDeviceResponse, nil
}

// UpdateDevice method.
func (n *Client) UpdateDevice(ctx context.Context, in *ns.UpdateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateDeviceChan <- *in
	return &n.UpdateDeviceResponse, nil
}

// DeleteDevice method.
func (n *Client) DeleteDevice(ctx context.Context, in *ns.DeleteDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteDeviceChan <- *in
	return &n.DeleteDeviceResponse, nil
}

// ActivateDevice method.
func (n *Client) ActivateDevice(ctx context.Context, in *ns.ActivateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.ActivateDeviceChan <- *in
	return &n.ActivateDeviceResponse, nil
}

// GetDeviceActivation method.
func (n *Client) GetDeviceActivation(ctx context.Context, in *ns.GetDeviceActivationRequest, opts ...grpc.CallOption) (*ns.GetDeviceActivationResponse, error) {
	n.GetDeviceActivationChan <- *in
	return &n.GetDeviceActivationResponse, nil
}

// DeactivateDevice method.
func (n *Client) DeactivateDevice(ctx context.Context, in *ns.DeactivateDeviceRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeactivateDeviceChan <- *in
	return &n.DeactivateDeviceResponse, nil
}

// CreateGateway method.
func (n *Client) CreateGateway(ctx context.Context, in *ns.CreateGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateGatewayChan <- *in
	return &n.CreateGatewayResponse, nil
}

// GetGateway method.
func (n *Client) GetGateway(ctx context.Context, in *ns.GetGatewayRequest, opts ...grpc.CallOption) (*ns.GetGatewayResponse, error) {
	n.GetGatewayChan <- *in
	return &n.GetGatewayResponse, nil
}

// UpdateGateway method.
func (n *Client) UpdateGateway(ctx context.Context, in *ns.UpdateGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateGatewayChan <- *in
	return &n.UpdateGatewayResponse, nil
}

// DeleteGateway method.
func (n *Client) DeleteGateway(ctx context.Context, in *ns.DeleteGatewayRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteGatewayChan <- *in
	return &n.DeleteGatewayResponse, nil
}

// GetGatewayStats method.
func (n *Client) GetGatewayStats(ctx context.Context, in *ns.GetGatewayStatsRequest, opts ...grpc.CallOption) (*ns.GetGatewayStatsResponse, error) {
	n.GetGatewayStatsChan <- *in
	return &n.GetGatewayStatsResponse, nil
}

// GetRandomDevAddr method.
func (n *Client) GetRandomDevAddr(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ns.GetRandomDevAddrResponse, error) {
	n.GetRandomDevAddrChan <- *in
	return &n.GetRandomDevAddrResponse, nil
}

// CreateMACCommandQueueItem method.
func (n *Client) CreateMACCommandQueueItem(ctx context.Context, in *ns.CreateMACCommandQueueItemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateMACCommandQueueItemChan <- *in
	return &n.CreateMACCommandQueueItemResponse, nil
}

// SendProprietaryPayload method.
func (n *Client) SendProprietaryPayload(ctx context.Context, in *ns.SendProprietaryPayloadRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.SendProprietaryPayloadChan <- *in
	return &n.SendProprietaryPayloadResponse, nil
}

// CreateGatewayProfile method.
func (n *Client) CreateGatewayProfile(ctx context.Context, in *ns.CreateGatewayProfileRequest, opts ...grpc.CallOption) (*ns.CreateGatewayProfileResponse, error) {
	n.CreateGatewayProfileChan <- *in
	return &n.CreateGatewayProfileResponse, nil
}

// GetGatewayProfile method.
func (n *Client) GetGatewayProfile(ctx context.Context, in *ns.GetGatewayProfileRequest, opts ...grpc.CallOption) (*ns.GetGatewayProfileResponse, error) {
	n.GetGatewayProfileChan <- *in
	return &n.GetGatewayProfileResponse, nil
}

// UpdateGatewayProfile method.
func (n *Client) UpdateGatewayProfile(ctx context.Context, in *ns.UpdateGatewayProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateGatewayProfileChan <- *in
	return &n.UpdateGatewayProfileResponse, nil
}

// DeleteGatewayProfile method.
func (n *Client) DeleteGatewayProfile(ctx context.Context, in *ns.DeleteGatewayProfileRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteGatewayProfileChan <- *in
	return &n.DeleteGatewayProfileResponse, nil
}

// CreateDeviceQueueItem method.
func (n *Client) CreateDeviceQueueItem(ctx context.Context, in *ns.CreateDeviceQueueItemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.CreateDeviceQueueItemChan <- *in
	return &n.CreateDeviceQueueItemResponse, nil
}

// FlushDeviceQueueForDevEUI method.
func (n *Client) FlushDeviceQueueForDevEUI(ctx context.Context, in *ns.FlushDeviceQueueForDevEUIRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.FlushDeviceQueueForDevEUIChan <- *in
	return &n.FlushDeviceQueueForDevEUIResponse, nil
}

// GetDeviceQueueItemsForDevEUI method.
func (n *Client) GetDeviceQueueItemsForDevEUI(ctx context.Context, in *ns.GetDeviceQueueItemsForDevEUIRequest, opts ...grpc.CallOption) (*ns.GetDeviceQueueItemsForDevEUIResponse, error) {
	n.GetDeviceQueueItemsForDevEUIChan <- *in
	return &n.GetDeviceQueueItemsForDevEUIResponse, nil
}

// GetNextDownlinkFCntForDevEUI method.
func (n *Client) GetNextDownlinkFCntForDevEUI(ctx context.Context, in *ns.GetNextDownlinkFCntForDevEUIRequest, opts ...grpc.CallOption) (*ns.GetNextDownlinkFCntForDevEUIResponse, error) {
	n.GetNextDownlinkFCntForDevEUIChan <- *in
	return &n.GetNextDownlinkFCntForDevEUIResponse, nil
}

// GetVersion method.
func (n *Client) GetVersion(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ns.GetVersionResponse, error) {
	return &n.GetVersionResponse, nil
}

// StreamFrameLogsForGateway method.
func (n *Client) StreamFrameLogsForGateway(ctx context.Context, in *ns.StreamFrameLogsForGatewayRequest, opts ...grpc.CallOption) (ns.NetworkServerService_StreamFrameLogsForGatewayClient, error) {
	panic("not implemented")
}

// StreamFrameLogsForDevice method.
func (n *Client) StreamFrameLogsForDevice(ctx context.Context, in *ns.StreamFrameLogsForDeviceRequest, opts ...grpc.CallOption) (ns.NetworkServerService_StreamFrameLogsForDeviceClient, error) {
	panic("not implemented")
}

// CreateMulticastGroup method.
func (n *Client) CreateMulticastGroup(ctx context.Context, in *ns.CreateMulticastGroupRequest, opts ...grpc.CallOption) (*ns.CreateMulticastGroupResponse, error) {
	n.CreateMulticastGroupChan <- *in
	return &n.CreateMulticastGroupResponse, nil
}

// GetMulticastGroup method.
func (n *Client) GetMulticastGroup(ctx context.Context, in *ns.GetMulticastGroupRequest, opts ...grpc.CallOption) (*ns.GetMulticastGroupResponse, error) {
	n.GetMulticastGroupChan <- *in
	return &n.GetMulticastGroupResponse, nil
}

// UpdateMulticastGroup method.
func (n *Client) UpdateMulticastGroup(ctx context.Context, in *ns.UpdateMulticastGroupRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.UpdateMulticastGroupChan <- *in
	return &n.UpdateMulticastGroupResponse, nil
}

// DeleteMulticastGroup method.
func (n *Client) DeleteMulticastGroup(ctx context.Context, in *ns.DeleteMulticastGroupRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.DeleteMulticastGroupChan <- *in
	return &n.DeleteMulticastGroupResponse, nil
}

// AddDeviceToMulticastGroup method.
func (n *Client) AddDeviceToMulticastGroup(ctx context.Context, in *ns.AddDeviceToMulticastGroupRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.AddDeviceToMulticastGroupChan <- *in
	return &n.AddDeviceToMulticastGroupResponse, nil
}

// RemoveDeviceFromMulticastGroup method.
func (n *Client) RemoveDeviceFromMulticastGroup(ctx context.Context, in *ns.RemoveDeviceFromMulticastGroupRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.RemoveDeviceFromMulticastGroupChan <- *in
	return &n.RemoveDeviceFromMulticastGroupResponse, nil
}

// EnqueueMulticastQueueItem method.
func (n *Client) EnqueueMulticastQueueItem(ctx context.Context, in *ns.EnqueueMulticastQueueItemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.EnqueueMulticastQueueItemChan <- *in
	return &n.EnqueueMulticastQueueItemResponse, nil
}

// FlushMulticastQueueForMulticastGroup method.
func (n *Client) FlushMulticastQueueForMulticastGroup(ctx context.Context, in *ns.FlushMulticastQueueForMulticastGroupRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.FlushMulticastQueueForMulticastGroupChan <- *in
	return &n.FlushMulticastQueueForMulticastGroupResponse, nil
}

// GetMulticastQueueItemsForMulticastGroup method.
func (n *Client) GetMulticastQueueItemsForMulticastGroup(ctx context.Context, in *ns.GetMulticastQueueItemsForMulticastGroupRequest, opts ...grpc.CallOption) (*ns.GetMulticastQueueItemsForMulticastGroupResponse, error) {
	n.GetMulticastQueueItemsForMulticastGroupChan <- *in
	return &n.GetMulticastQueueItemsForMulticastGroupResponse, nil
}

// GenerateGatewayClientCertificate method.
func (n *Client) GenerateGatewayClientCertificate(ctx context.Context, in *ns.GenerateGatewayClientCertificateRequest, opts ...grpc.CallOption) (*ns.GenerateGatewayClientCertificateResponse, error) {
	n.GenerateGatewayClientCertificateChan <- *in
	return &n.GenerateGatewayClientCertificateResponse, nil
}

// GetADRAlgorithms method.
func (n *Client) GetADRAlgorithms(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ns.GetADRAlgorithmsResponse, error) {
	return &n.GetADRAlgorithmsResponse, nil
}

// ClearDeviceNonces method.
func (n *Client) ClearDeviceNonces(ctx context.Context, in *ns.ClearDeviceNoncesRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	n.ClearDeviceNoncesChan <- *in
	return &n.ClearDeviceNoncesResponse, nil
}
