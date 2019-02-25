package mock

import "github.com/brocaar/lora-app-server/internal/integration"

// Integration implements a mock integration.
type Integration struct {
	SendDataUpChan               chan integration.DataUpPayload
	SendJoinNotificationChan     chan integration.JoinNotification
	SendACKNotificationChan      chan integration.ACKNotification
	SendErrorNotificationChan    chan integration.ErrorNotification
	DataDownPayloadChan          chan integration.DataDownPayload
	SendStatusNotificationChan   chan integration.StatusNotification
	SendLocationNotificationChan chan integration.LocationNotification
}

// New creates a new mock integration.
func New() *Integration {
	return &Integration{
		SendDataUpChan:               make(chan integration.DataUpPayload, 100),
		SendJoinNotificationChan:     make(chan integration.JoinNotification, 100),
		SendACKNotificationChan:      make(chan integration.ACKNotification, 100),
		SendErrorNotificationChan:    make(chan integration.ErrorNotification, 100),
		DataDownPayloadChan:          make(chan integration.DataDownPayload, 100),
		SendStatusNotificationChan:   make(chan integration.StatusNotification, 100),
		SendLocationNotificationChan: make(chan integration.LocationNotification, 100),
	}
}

// Close method.
func (i *Integration) Close() error {
	return nil
}

// SendDataUp method.
func (i *Integration) SendDataUp(payload integration.DataUpPayload) error {
	i.SendDataUpChan <- payload
	return nil
}

// SendJoinNotification Method.
func (i *Integration) SendJoinNotification(payload integration.JoinNotification) error {
	i.SendJoinNotificationChan <- payload
	return nil
}

// SendACKNotification method.
func (i *Integration) SendACKNotification(payload integration.ACKNotification) error {
	i.SendACKNotificationChan <- payload
	return nil
}

// SendErrorNotification method.
func (i *Integration) SendErrorNotification(payload integration.ErrorNotification) error {
	i.SendErrorNotificationChan <- payload
	return nil
}

// DataDownChan method.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return i.DataDownPayloadChan
}

// SendStatusNotification method.
func (i *Integration) SendStatusNotification(payload integration.StatusNotification) error {
	i.SendStatusNotificationChan <- payload
	return nil
}

// SendLocationNotification method.
func (i *Integration) SendLocationNotification(payload integration.LocationNotification) error {
	i.SendLocationNotificationChan <- payload
	return nil
}
