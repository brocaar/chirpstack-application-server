// Package multi implements a multi-integration handler.
// This handler can be used to combine the handling of multiple integrations.
package multi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/amqp"
	"github.com/brocaar/chirpstack-application-server/internal/integration/awssns"
	"github.com/brocaar/chirpstack-application-server/internal/integration/azureservicebus"
	"github.com/brocaar/chirpstack-application-server/internal/integration/gcppubsub"
	"github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/influxdb"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mqtt"
	"github.com/brocaar/chirpstack-application-server/internal/integration/postgresql"
	"github.com/brocaar/chirpstack-application-server/internal/integration/thingsboard"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// Integration implements the multi integration.
type Integration struct {
	integrations []integration.Integrator
}

// New create a new multi integration.
// The argument that must be given is a slice of configuration objects for
// the handlers to setup.
func New(m marshaler.Type, confs []interface{}) (*Integration, error) {
	var integrations []integration.Integrator

	for i := range confs {
		conf := confs[i]
		var ii integration.Integrator
		var err error

		switch v := conf.(type) {
		case config.IntegrationAWSSNSConfig:
			ii, err = awssns.New(m, v)
		case config.IntegrationAzureConfig:
			ii, err = azureservicebus.New(m, v)
		case config.IntegrationGCPConfig:
			ii, err = gcppubsub.New(m, v)
		case http.Config:
			ii, err = http.New(m, v)
		case influxdb.Config:
			ii, err = influxdb.New(v)
		case config.IntegrationMQTTConfig:
			ii, err = mqtt.New(m, storage.RedisPool(), v)
		case config.IntegrationPostgreSQLConfig:
			ii, err = postgresql.New(v)
		case thingsboard.Config:
			ii, err = thingsboard.New(v)
		case config.IntegrationAMQPConfig:
			ii, err = amqp.New(m, v)
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
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendDataUp(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendJoinNotification(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendACKNotification(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendErrorNotification(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendStatusNotification(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
			}
		}(ii)
	}

	return nil
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	for _, ii := range i.integrations {
		go func(i integration.Integrator) {
			if err := i.SendLocationNotification(ctx, vars, pl); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"integration": fmt.Sprintf("%T", i),
					"ctx_id":      ctx.Value(logging.ContextIDKey),
				}).Error("integration/multi: integration error")
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
