package amqp

import (
	"bytes"
	"context"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
	"github.com/streadway/amqp"
)

// Integration implements an AMQP integration.
type Integration struct {
	chPool *pool

	marshaler       marshaler.Type
	eventRoutingKey *template.Template
}

// New creates a new AMQP integration.
func New(m marshaler.Type, conf config.IntegrationAMQPConfig) (*Integration, error) {
	var err error
	i := Integration{
		marshaler: m,
	}

	log.Info("integration/amqp: connecting to amqp broker")
	i.chPool, err = newPool(10, conf.URL)
	if err != nil {
		return nil, errors.Wrap(err, "new amqp channel pool error")
	}

	i.eventRoutingKey, err = template.New("event").Parse(conf.EventRoutingKeyTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse event routing-key template error")
	}

	return &i, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "up", &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "join", &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "ack", &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "error", &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "status", &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "location", &pl)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "txack", &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "integration", &pl)
}

// DataDownChan returns nil
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return i.chPool.close()
}

func (i *Integration) publishEvent(ctx context.Context, applicationID uint64, devEUIB []byte, typ string, msg proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	ch, err := i.chPool.get()
	if err != nil {
		return errors.Wrap(err, "get amqp channel error")
	}
	defer ch.close()

	routingKey := bytes.NewBuffer(nil)
	err = i.eventRoutingKey.Execute(routingKey, struct {
		ApplicationID uint64
		DevEUI        lorawan.EUI64
		EventType     string
	}{applicationID, devEUI, typ})
	if err != nil {
		return errors.Wrap(err, "execute template error")
	}

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"routing_key": routingKey.String(),
		"ctx_id":      ctx.Value(logging.ContextIDKey),
	}).Info("integration/amqp: publishing event")

	var contentType string
	switch i.marshaler {
	case marshaler.ProtobufJSON, marshaler.JSONV3:
		contentType = "application/json"
	case marshaler.Protobuf:
		contentType = "application/octet-stream"
	}

	err = ch.ch.Publish(
		"amq.topic",
		routingKey.String(),
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        b,
		},
	)
	if err != nil {
		ch.markUnusable()
		return errors.Wrap(err, "publish event error")
	}

	return nil
}
