package integration

// Handler kinds
const (
	HTTP     = "HTTP"
	InfluxDB = "INFLUXDB"
)

// Integrator defines the interface that an intergration must implement.
type Integrator interface {
	SendDataUp(payload DataUpPayload) error                      // send data-up payload
	SendJoinNotification(payload JoinNotification) error         // send join notification
	SendACKNotification(payload ACKNotification) error           // send ack notification
	SendErrorNotification(payload ErrorNotification) error       // send error notification
	SendStatusNotification(payload StatusNotification) error     // send status notification
	SendLocationNotification(payload LocationNotification) error // send location notofication
	DataDownChan() chan DataDownPayload                          // returns DataDownPayload channel
	Close() error                                                // closes the handler
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
