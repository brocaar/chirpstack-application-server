package kafka

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"text/template"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements an Kafka integration.
type Integration struct {
	marshaler        marshaler.Type
	writer           *kafka.Writer
	eventKeyTemplate *template.Template
	config           config.IntegrationKafkaConfig
}

// New creates a new Kafka integration.
func New(m marshaler.Type, conf config.IntegrationKafkaConfig) (*Integration, error) {
	wc := kafka.WriterConfig{
		Brokers:  conf.Brokers,
		Topic:    conf.Topic,
		Balancer: &kafka.LeastBytes{},

		// Equal to kafka.DefaultDialer.
		// We do not want to use kafka.DefaultDialer itself, as we might modify
		// it below to setup SASLMechanism.
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		},
	}

	if conf.TLS {
		wc.Dialer.TLS = &tls.Config{}
	}

	if conf.Username != "" || conf.Password != "" {
		switch conf.Mechanism {
		case "plain":
			wc.Dialer.SASLMechanism = plain.Mechanism{
				Username: conf.Username,
				Password: conf.Password,
			}
		case "scram":
			var algorithm scram.Algorithm

			switch conf.Algorithm {
			case "SHA-512":
				algorithm = scram.SHA512
			case "SHA-256":
				algorithm = scram.SHA256
			default:
				return nil, fmt.Errorf("unknown sasl algorithm %s", conf.Algorithm)
			}

			mechanism, err := scram.Mechanism(algorithm, conf.Username, conf.Password)
			if err != nil {
				return nil, errors.Wrap(err, "sasl mechanism")
			}

			wc.Dialer.SASLMechanism = mechanism
		default:
			return nil, fmt.Errorf("unknown sasl mechanism %s", conf.Mechanism)
		}

	}

	log.WithFields(log.Fields{
		"brokers": conf.Brokers,
		"topic":   conf.Topic,
	}).Info("integration/kafka: connecting to kafka broker(s)")

	w := kafka.NewWriter(wc)

	log.Info("integration/kafka: connected to kafka broker(s)")

	kt, err := template.New("key").Parse(conf.EventKeyTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse key template")
	}

	i := Integration{
		marshaler:        m,
		writer:           w,
		eventKeyTemplate: kt,
		config:           conf,
	}
	return &i, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "up", &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "join", &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "ack", &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "error", &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "status", &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "location", &pl)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "txack", &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return i.publish(ctx, pl.ApplicationId, pl.DevEui, "integration", &pl)
}

func (i *Integration) publish(ctx context.Context, applicationID uint64, devEUIB []byte, event string, msg proto.Message) error {
	if i.writer == nil {
		return fmt.Errorf("integration closed")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return err
	}

	keyBuf := bytes.NewBuffer(nil)
	err = i.eventKeyTemplate.Execute(keyBuf, struct {
		ApplicationID uint64
		DevEUI        lorawan.EUI64
		EventType     string
	}{applicationID, devEUI, event})
	if err != nil {
		return errors.Wrap(err, "executing template")
	}
	key := keyBuf.Bytes()

	kmsg := kafka.Message{
		Value: b,
		Headers: []kafka.Header{
			{
				Key:   "event",
				Value: []byte(event),
			},
		},
	}
	if len(key) > 0 {
		kmsg.Key = key
	}

	log.WithFields(log.Fields{
		"key":    string(key),
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("integration/kafka: publishing message")

	if err := i.writer.WriteMessages(ctx, kmsg); err != nil {
		return errors.Wrap(err, "writing message to kafka")
	}
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close shuts down the integration, closing the Kafka writer.
func (i *Integration) Close() error {
	if i.writer == nil {
		return fmt.Errorf("integration already closed")
	}
	err := i.writer.Close()
	i.writer = nil
	return err
}
