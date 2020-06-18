package mock

import (
	"context"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
)

// Integration implements a mock integration.
type Integration struct {
	SendDataUpChan                  chan pb.UplinkEvent
	SendJoinNotificationChan        chan pb.JoinEvent
	SendACKNotificationChan         chan pb.AckEvent
	SendErrorNotificationChan       chan pb.ErrorEvent
	DataDownPayloadChan             chan models.DataDownPayload
	SendStatusNotificationChan      chan pb.StatusEvent
	SendLocationNotificationChan    chan pb.LocationEvent
	SendTxAckNotificationChan       chan pb.TxAckEvent
	SendIntegrationNotificationChan chan pb.IntegrationEvent
}

// New creates a new mock integration.
func New() *Integration {
	return &Integration{
		SendDataUpChan:                  make(chan pb.UplinkEvent, 100),
		SendJoinNotificationChan:        make(chan pb.JoinEvent, 100),
		SendACKNotificationChan:         make(chan pb.AckEvent, 100),
		SendErrorNotificationChan:       make(chan pb.ErrorEvent, 100),
		DataDownPayloadChan:             make(chan models.DataDownPayload, 100),
		SendStatusNotificationChan:      make(chan pb.StatusEvent, 100),
		SendLocationNotificationChan:    make(chan pb.LocationEvent, 100),
		SendTxAckNotificationChan:       make(chan pb.TxAckEvent, 100),
		SendIntegrationNotificationChan: make(chan pb.IntegrationEvent, 100),
	}
}

// Close method.
func (i *Integration) Close() error {
	return nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, vars map[string]string, payload pb.UplinkEvent) error {
	i.SendDataUpChan <- payload
	return nil
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, vars map[string]string, payload pb.JoinEvent) error {
	i.SendJoinNotificationChan <- payload
	return nil
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, vars map[string]string, payload pb.AckEvent) error {
	i.SendACKNotificationChan <- payload
	return nil
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, vars map[string]string, payload pb.ErrorEvent) error {
	i.SendErrorNotificationChan <- payload
	return nil
}

// DataDownChan method.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return i.DataDownPayloadChan
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, vars map[string]string, payload pb.StatusEvent) error {
	i.SendStatusNotificationChan <- payload
	return nil
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, vars map[string]string, payload pb.LocationEvent) error {
	i.SendLocationNotificationChan <- payload
	return nil
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, vars map[string]string, payload pb.TxAckEvent) error {
	i.SendTxAckNotificationChan <- payload
	return nil
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, vars map[string]string, payload pb.IntegrationEvent) error {
	i.SendIntegrationNotificationChan <- payload
	return nil
}
