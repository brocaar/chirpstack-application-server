// Package logger implements an integration which writes logs to Redis.
package logger

import (
	"context"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/eventlog"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Config contains the logger configuration.
type Config struct{}

// Integration implements the logger integration.
type Integration struct {
	config Config
}

// New creates a new logger integration.
func New(conf Config) (*Integration, error) {
	return &Integration{
		config: conf,
	}, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Uplink, devEUI, &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Join, devEUI, &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.ACK, devEUI, &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Error, devEUI, &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Status, devEUI, &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Location, devEUI, &pl)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.TxAck, devEUI, &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	return i.log(ctx, eventlog.Integration, devEUI, &pl)
}

// Close is not implemented.
func (i *Integration) Close() error {
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

func (i *Integration) log(ctx context.Context, typ string, devEUI lorawan.EUI64, msg proto.Message) error {
	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"type":    typ,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/logger: logging event")

	return eventlog.LogEventForDevice(devEUI, typ, msg)
}
