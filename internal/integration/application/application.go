// Package application implements a wrapper for application integrations.
package application

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/http"
	"github.com/brocaar/lora-app-server/internal/integration/influxdb"
	"github.com/brocaar/lora-app-server/internal/integration/multi"
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
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get application integration error")
	}
	defer multi.Close()

	return multi.SendDataUp(pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendJoinNotification(pl)
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendACKNotification(pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendErrorNotification(pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendStatusNotification(pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	multi, err := i.getApplicationIntegration(pl.ApplicationID)
	if err != nil {
		return errors.Wrap(err, "get appplication integration error")
	}
	defer multi.Close()

	return multi.SendLocationNotification(pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) getApplicationIntegration(id int64) (integration.Integrator, error) {
	var configs []interface{}

	// read integrations
	appints, err := storage.GetIntegrationsForApplicationID(config.C.PostgreSQL.DB, id)
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
		default:
			return nil, fmt.Errorf("unknown integration type: %s", appint.Kind)
		}
	}

	return multi.New(configs)
}
