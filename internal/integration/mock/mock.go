package mock

import (
	"context"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
)

// Integration implements a mock integration.
type Integration struct {
	SendDataUpChan               chan pb.UplinkEvent
	SendJoinNotificationChan     chan pb.JoinEvent
	SendACKNotificationChan      chan pb.AckEvent
	SendErrorNotificationChan    chan pb.ErrorEvent
	DataDownPayloadChan          chan integration.DataDownPayload
	SendStatusNotificationChan   chan pb.StatusEvent
	SendLocationNotificationChan chan pb.LocationEvent
}

// New creates a new mock integration.
func New() *Integration {
	return &Integration{
		SendDataUpChan:               make(chan pb.UplinkEvent, 100),
		SendJoinNotificationChan:     make(chan pb.JoinEvent, 100),
		SendACKNotificationChan:      make(chan pb.AckEvent, 100),
		SendErrorNotificationChan:    make(chan pb.ErrorEvent, 100),
		DataDownPayloadChan:          make(chan integration.DataDownPayload, 100),
		SendStatusNotificationChan:   make(chan pb.StatusEvent, 100),
		SendLocationNotificationChan: make(chan pb.LocationEvent, 100),
	}
}

// Close method.
func (i *Integration) Close() error {
	return nil
}

// SendDataUp method.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, payload pb.UplinkEvent) error {
	i.SendDataUpChan <- payload
	return nil
}

// SendJoinNotification Method.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, payload pb.JoinEvent) error {
	i.SendJoinNotificationChan <- payload
	return nil
}

// SendACKNotification method.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, payload pb.AckEvent) error {
	i.SendACKNotificationChan <- payload
	return nil
}

// SendErrorNotification method.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, payload pb.ErrorEvent) error {
	i.SendErrorNotificationChan <- payload
	return nil
}

// DataDownChan method.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return i.DataDownPayloadChan
}

// SendStatusNotification method.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, payload pb.StatusEvent) error {
	i.SendStatusNotificationChan <- payload
	return nil
}

// SendLocationNotification method.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, payload pb.LocationEvent) error {
	i.SendLocationNotificationChan <- payload
	return nil
}
