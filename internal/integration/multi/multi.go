// Package multi implements a multi-integration handler.
// This handler can be used to combine the handling of multiple integrations.
package multi

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/awssns"
	"github.com/brocaar/lora-app-server/internal/integration/azureservicebus"
	"github.com/brocaar/lora-app-server/internal/integration/gcppubsub"
	"github.com/brocaar/lora-app-server/internal/integration/http"
	"github.com/brocaar/lora-app-server/internal/integration/influxdb"
	"github.com/brocaar/lora-app-server/internal/integration/mqtt"
)

// Integration implements the multi integration.
type Integration struct {
	integrations []integration.Integrator
}

// New create a new multi integration.
// The argument that must be given is a slice of configuration objects for
// the handlers to setup.
func New(confs []interface{}) (*Integration, error) {
	var integrations []integration.Integrator

	for i := range confs {
		conf := confs[i]
		var ii integration.Integrator
		var err error

		switch v := conf.(type) {
		case awssns.Config:
			ii, err = awssns.New(v)
		case azureservicebus.Config:
			ii, err = azureservicebus.New(v)
		case gcppubsub.Config:
			ii, err = gcppubsub.New(v)
		case http.Config:
			ii, err = http.New(v)
		case influxdb.Config:
			ii, err = influxdb.New(v)
		case mqtt.Config:
			ii, err = mqtt.New(config.C.Redis.Pool, v)
		default:
			return nil, fmt.Errorf("unknown configuration type %T", conf)
		}

		if err != nil {
			return nil, errors.Wrap(err, "new integration error")
		}

		integrations = append(integrations, ii)
	}

	return &Integration{
		integrations: integrations,
	}, nil
}

// Add appends a new integration to the list.
func (i *Integration) Add(intg integration.Integrator) {
	i.integrations = append(i.integrations, intg)
}

// SendDataUp sends a data-up payload.
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendDataUp(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendJoinNotification(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendACKNotification(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendErrorNotification(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendStatusNotification(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendLocationNotification(pl); err != nil {
				log.WithError(err).Errorf("integration/multi: integration %T error", i)
			}
		}(ii)
	}

	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	for _, ii := range i.integrations {
		if c := ii.DataDownChan(); c != nil {
			return c
		}
	}
	return nil
}

// Close closes the handlers.
func (i *Integration) Close() error {
	for _, ii := range i.integrations {
		if err := ii.Close(); err != nil {
			return err
		}
	}

	return nil
}
