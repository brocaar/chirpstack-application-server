package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/amqp"
	"github.com/brocaar/chirpstack-application-server/internal/integration/awssns"
	"github.com/brocaar/chirpstack-application-server/internal/integration/azureservicebus"
	"github.com/brocaar/chirpstack-application-server/internal/integration/gcppubsub"
	"github.com/brocaar/chirpstack-application-server/internal/integration/http"
	"github.com/brocaar/chirpstack-application-server/internal/integration/influxdb"
	"github.com/brocaar/chirpstack-application-server/internal/integration/kafka"
	"github.com/brocaar/chirpstack-application-server/internal/integration/logger"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mqtt"
	"github.com/brocaar/chirpstack-application-server/internal/integration/multi"
	"github.com/brocaar/chirpstack-application-server/internal/integration/mydevices"
	"github.com/brocaar/chirpstack-application-server/internal/integration/pilotthings"
	"github.com/brocaar/chirpstack-application-server/internal/integration/postgresql"
	"github.com/brocaar/chirpstack-application-server/internal/integration/thingsboard"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// Handler kinds
const (
	HTTP            = "HTTP"
	InfluxDB        = "INFLUXDB"
	ThingsBoard     = "THINGSBOARD"
	MyDevices       = "MYDEVICES"
	LoRaCloud       = "LORACLOUD"
	GCPPubSub       = "GCP_PUBSUB"
	AWSSNS          = "AWS_SNS"
	AzureServiceBus = "AZURE_SERVICE_BUS"
	PilotThings     = "PILOT_THINGS"
)

var (
	mockIntegration    models.Integration
	marshalType        marshaler.Type
	globalIntegrations []models.IntegrationHandler
)

// Setup configures the integration package.
func Setup(conf config.Config) error {
	log.Info("integration: configuring global integrations")

	if err := mqtt.Setup(conf); err != nil {
		return err
	}

	var ints []models.IntegrationHandler

	// setup marshaler
	switch conf.ApplicationServer.Integration.Marshaler {
	case "protobuf":
		marshalType = marshaler.Protobuf
	case "json":
		marshalType = marshaler.ProtobufJSON
	case "json_v3":
		marshalType = marshaler.JSONV3
	}

	// configure logger integration (for device events in web-interface)
	i, err := logger.New(logger.Config{})
	if err != nil {
		return errors.Wrap(err, "new logger integration error")
	}
	ints = append(ints, i)

	// setup global integrations, to be used by all applications
	for _, name := range conf.ApplicationServer.Integration.Enabled {
		var i models.IntegrationHandler
		var err error

		switch name {
		case "aws_sns":
			i, err = awssns.New(marshalType, conf.ApplicationServer.Integration.AWSSNS)
		case "azure_service_bus":
			i, err = azureservicebus.New(marshalType, conf.ApplicationServer.Integration.AzureServiceBus)
		case "mqtt":
			i, err = mqtt.New(marshalType, conf.ApplicationServer.Integration.MQTT)
		case "gcp_pub_sub":
			i, err = gcppubsub.New(marshalType, conf.ApplicationServer.Integration.GCPPubSub)
		case "kafka":
			i, err = kafka.New(marshalType, conf.ApplicationServer.Integration.Kafka)
		case "postgresql":
			i, err = postgresql.New(marshalType, conf.ApplicationServer.Integration.PostgreSQL)
		case "amqp":
			i, err = amqp.New(marshalType, conf.ApplicationServer.Integration.AMQP)
		default:
			return fmt.Errorf("unknonwn integration type: %s", name)
		}

		if err != nil {
			return errors.Wrap(err, "new integration error")
		}

		ints = append(ints, i)
	}
	globalIntegrations = ints

	return nil
}

// ForApplicationID returns the integration handler for the given application ID.
// The returned handler will be a "multi-handler", containing both the global
// integrations and the integrations setup specifically for the given
// application ID.
// When the given application ID equals 0, only the global integrations are
// returned.
func ForApplicationID(id int64) models.Integration {
	// for testing, return mock integration
	if mockIntegration != nil {
		return mockIntegration
	}

	var appints []storage.Integration
	var err error

	// retrieve application integrations when ID != 0
	if id != 0 {
		appints, err = storage.GetIntegrationsForApplicationID(context.TODO(), storage.DB(), id)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"application_id": id,
			}).Error("integrations: get application integrations error")
		}
	}

	// parse integration configs and setup integrations
	var ints []models.IntegrationHandler
	for _, appint := range appints {
		var i models.IntegrationHandler
		var err error

		switch appint.Kind {
		case HTTP:
			// read config
			var conf http.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read http configuration error")
				continue
			}

			// create new http integration
			i, err = http.New(marshalType, conf)
		case InfluxDB:
			// read config
			var conf influxdb.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read influxdb configuration error")
				continue
			}

			// create new influxdb integration
			i, err = influxdb.New(conf)
		case ThingsBoard:
			// read config
			var conf thingsboard.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read thingsboard configuration error")
				continue
			}

			// create new thingsboard integration
			i, err = thingsboard.New(conf)
		case MyDevices:
			// read config
			var conf mydevices.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read mydevices configuration error")
				continue
			}

			// create new mydevices integration
			i, err = mydevices.New(conf)
		case LoRaCloud:
			// read config
			var conf loracloud.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read loracloud configuration error")
				continue
			}

			// create new loracloud integration
			i, err = loracloud.New(conf)
		case GCPPubSub:
			// read config
			var conf config.IntegrationGCPConfig
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read gcp pubsub configuration error")
				continue
			}

			// create new gcp pubsub integration
			i, err = gcppubsub.New(marshalType, conf)
		case AWSSNS:
			// read config
			var conf config.IntegrationAWSSNSConfig
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read aws sns configuration error")
				continue
			}

			// create new aws sns integration
			i, err = awssns.New(marshalType, conf)
		case AzureServiceBus:
			// read config
			var conf config.IntegrationAzureConfig
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&conf); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read azure service-bus configuration error")
				continue
			}

			// create new aws sns integration
			i, err = azureservicebus.New(marshalType, conf)
		case PilotThings:
			var config pilotthings.Config
			if err := json.NewDecoder(bytes.NewReader(appint.Settings)).Decode(&config); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"application_id": id,
				}).Error("integrtations: read pilot things configuration error")
				continue
			}

			// create new pilot things integration
			i, err = pilotthings.New(config)
		default:
			log.WithFields(log.Fields{
				"application_id": id,
				"kind":           appint.Kind,
			}).Error("integrations: unknown integration type")
			continue
		}

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"application_id": id,
				"kind":           appint.Kind,
			}).Error("integrations: new integration error")
			continue
		}

		ints = append(ints, i)
	}

	return multi.New(globalIntegrations, ints)
}

// SetMockIntegration mocks the integration.
func SetMockIntegration(i models.Integration) {
	mockIntegration = i
}
