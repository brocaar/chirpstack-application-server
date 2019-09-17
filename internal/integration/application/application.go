// Package application implements a wrapper for application integrations.
package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/http"
	"github.com/brocaar/lora-app-server/internal/integration/influxdb"
	"github.com/brocaar/lora-app-server/internal/integration/multi"
	"github.com/brocaar/lora-app-server/internal/integration/thingsboard"
	"github.com/brocaar/lora-app-server/internal/storage"
)

// Integration implements the application integration wrapper.
// Per request it will fetch the application integrations and forward the
// request to these integrations.
type Integration struct{}

// New creates a new application integration.
func New() *Integration {
	return &Integration{}
}

// SendDataUp sends an uplink payload.
func (i *Integration) SendDataUp(ctx context.Context, pl integration.DataUpPayload) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application integration error")
	}
	defer multi.Close()

	return multi.SendDataUp(ctx, pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, pl integration.JoinNotification) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendJoinNotification(ctx, pl)
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(ctx context.Context, pl integration.ACKNotification) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendACKNotification(ctx, pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, pl integration.ErrorNotification) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendErrorNotification(ctx, pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, pl integration.StatusNotification) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendStatusNotification(ctx, pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, pl integration.LocationNotification) error {
	multi, err := i.getApplicationIntegration(ctx, pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendLocationNotification(ctx, pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) getApplicationIntegration(ctx context.Context, id int64) (integration.Integrator, error) {
	var configs []interface{}

	// read integrations
	appints, err := storage.GetIntegrationsForApplicationID(ctx, storage.DB(), id)
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

	return multi.New(configs)
}
