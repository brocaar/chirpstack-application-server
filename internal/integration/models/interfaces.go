package models

import (
	"context"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
)

// Integration defines the integration interface.
type Integration interface {
	HandleUplinkEvent(ctx context.Context, vars map[string]string, pl integration.UplinkEvent) error
	HandleJoinEvent(ctx context.Context, vars map[string]string, pl integration.JoinEvent) error
	HandleAckEvent(ctx context.Context, vars map[string]string, pl integration.AckEvent) error
	HandleErrorEvent(ctx context.Context, vars map[string]string, pl integration.ErrorEvent) error
	HandleStatusEvent(ctx context.Context, vars map[string]string, pl integration.StatusEvent) error
	HandleLocationEvent(ctx context.Context, vars map[string]string, pl integration.LocationEvent) error
	HandleTxAckEvent(ctx context.Context, vars map[string]string, pl integration.TxAckEvent) error
	HandleIntegrationEvent(ctx context.Context, vars map[string]string, pl integration.IntegrationEvent) error
	DataDownChan() chan DataDownPayload
}

// IntegrationHandler defines the integration handler interface.
// This is different from Integration as one of the arguments is the Integration
// interface. This makes it possible for an IntegrationHandler to generate a new
// event. E.g. a HandleUplinkEvent method could call HandleLocationEvent, which
// then will be handled by all integrations.
type IntegrationHandler interface {
	HandleUplinkEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.UplinkEvent) error
	HandleJoinEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.JoinEvent) error
	HandleAckEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.AckEvent) error
	HandleErrorEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.ErrorEvent) error
	HandleStatusEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.StatusEvent) error
	HandleLocationEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.LocationEvent) error
	HandleTxAckEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.TxAckEvent) error
	HandleIntegrationEvent(ctx context.Context, i Integration, vars map[string]string, pl integration.IntegrationEvent) error
	DataDownChan() chan DataDownPayload
	Close() error
}
