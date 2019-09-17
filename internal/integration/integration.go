package integration

import "context"

// Handler kinds
const (
	HTTP        = "HTTP"
	InfluxDB    = "INFLUXDB"
	ThingsBoard = "THINGSBOARD"
)

// Integrator defines the interface that an intergration must implement.
type Integrator interface {
	SendDataUp(ctx context.Context, pl DataUpPayload) error                      // send data-up payload
	SendJoinNotification(ctx context.Context, pl JoinNotification) error         // send join notification
	SendACKNotification(ctx context.Context, pl ACKNotification) error           // send ack notification
	SendErrorNotification(ctx context.Context, pl ErrorNotification) error       // send error notification
	SendStatusNotification(ctx context.Context, pl StatusNotification) error     // send status notification
	SendLocationNotification(ctx context.Context, pl LocationNotification) error // send location notofication
	DataDownChan() chan DataDownPayload                                          // returns DataDownPayload channel
	Close() error                                                                // closes the handler
}

var integration Integrator

// Integration returns the integration object.
func Integration() Integrator {
	if integration == nil {
		panic("integration package must be initialized")
	}
	return integration
}

// SetIntegration sets the given integration.
func SetIntegration(i Integrator) {
	integration = i
}
