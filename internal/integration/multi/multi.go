// Package multi implements a multi-integration handler.
package multi

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

// Integration implements the multi integration.
type Integration struct {
	globalIntegrations []models.IntegrationHandler
	appIntegrations    []models.IntegrationHandler
}

// New creates a new multi-integration.
func New(global, app []models.IntegrationHandler) *Integration {
	return &Integration{
		globalIntegrations: global,
		appIntegrations:    app,
	}
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	for _, ii := range i.integrations() {

		go func(ii models.IntegrationHandler) {
			if err := ii.HandleUplinkEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleJoinEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleAckEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {

	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleErrorEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleStatusEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleLocationEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, vars map[string]string, pl pb.TxAckEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleTxAckEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, vars map[string]string, pl pb.IntegrationEvent) error {
	for _, ii := range i.integrations() {
		go func(ii models.IntegrationHandler) {
			if err := ii.HandleIntegrationEvent(ctx, i, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", ii),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	for _, ii := range i.globalIntegrations {
		if c := ii.DataDownChan(); c != nil {
			return c
		}
	}

	return nil
}

// integrations returns a slice with the global and application-integrations
// combined.
func (i *Integration) integrations() []models.IntegrationHandler {
	var ints []models.IntegrationHandler

	for _, ii := range i.globalIntegrations {
		ints = append(ints, ii)
	}

	for _, ii := range i.appIntegrations {
		ints = append(ints, ii)
	}

	return ints
}
