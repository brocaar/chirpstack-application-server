package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/segmentio/kafka-go"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements an Kafka integration.
type Integration struct {
	marshaler marshaler.Type
	writer    *kafka.Writer
	wg        sync.WaitGroup
	config    config.IntegrationKafkaConfig
}

// New creates a new Kafka integration.
func New(m marshaler.Type, conf config.IntegrationKafkaConfig) (*Integration, error) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  conf.Brokers,
		Topic:    conf.Topic,
		Balancer: &kafka.LeastBytes{},
	})

	i := Integration{
		marshaler: m,
		writer:    w,
		config:    conf,
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
	key := fmt.Sprintf("application.%d.device.%s", applicationID, devEUI.String())
	kmsg := kafka.Message{
		Key:   []byte(key),
		Value: b,
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI.String(),
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
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
