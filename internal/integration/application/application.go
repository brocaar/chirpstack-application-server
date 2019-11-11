// Package application implements a wrapper for application integrations.
package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/influxdb"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/multi"
	"github.com/brocaar/chirpstack-application-server/internal/integration/thingsboard"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// Integration implements the application integration wrapper.
// Per request it will fetch the application integrations and forward the
// request to these integrations.
type Integration struct {
	marshaler marshaler.Type
}

// New creates a new application integration.
func New(m marshaler.Type) *Integration {
	return &Integration{
		marshaler: m,
	}
}

// SendDataUp sends an uplink payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get application integration error")
	}
	defer multi.Close()

	return multi.SendDataUp(ctx, vars, pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendJoinNotification(ctx, vars, pl)
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendACKNotification(ctx, vars, pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendErrorNotification(ctx, vars, pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendStatusNotification(ctx, vars, pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationId)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendLocationNotification(ctx, vars, pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) getApplicationIntegration(ctx context.Context, id uint64) (integration.Integrator, error) {
	var configs []interface{}

	// read integrations
	appints, err := storage.GetIntegrationsForApplicationID(ctx, storage.DB(), int64(id))
	if err != nil {
		return nil, errors.Wrap(err, "get integrations for application id error")
	}

	// unmarshal configurations
	for _, appint := range appints {
		switch appint.Kind {
		case integration.HTTP:
			var conf http.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				return nil, errors.Wrap(err, "decode http integration config error")
			}
			configs = append(configs, conf)
		case integration.InfluxDB:
			var conf influxdb.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				return nil, errors.Wrap(err, "decode http integration config error")
			}
			configs = append(configs, conf)
		case integration.ThingsBoard:
			var conf thingsboard.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				return nil, errors.Wrap(err, "decode thingsboard integration config error")
			}
			configs = append(configs, conf)
		default:
			return nil, fmt.Errorf("unknown integration type: %s", appint.Kind)
		}
	}

	return multi.New(i.marshaler, configs)
}
