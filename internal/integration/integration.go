package integration

import (
	"context"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
)

// Handler kinds
const (
	HTTP        = "HTTP"
	InfluxDB    = "INFLUXDB"
	ThingsBoard = "THINGSBOARD"
)

// Integrator defines the interface that an intergration must implement.
type Integrator interface {
	SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error                 // send data-up payload
	SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error         // send join notification
	SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error           // send ack notification
	SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error       // send error notification
	SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error     // send status notification
	SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error // send location notofication
	DataDownChan() chan DataDownPayload                                                              // returns DataDownPayload channel
	Close() error                                                                                    // closes the handler
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
