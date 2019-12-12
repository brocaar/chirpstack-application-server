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
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
	"github.com/streadway/amqp"
)

// Integration implements an AMQP integration.
type Integration struct {
	conn   *amqp.Connection
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
	i.conn, err = amqp.Dial(conf.URL)
	if err != nil {
		return nil, errors.Wrap(err, "dial amqp url error")
	}

	i.chPool, err = newPool(10, i.conn)
	if err != nil {
		return nil, errors.Wrap(err, "new amqp channel pool error")
	}

	i.eventRoutingKey, err = template.New("event").Parse(conf.EventRoutingKeyTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse event routing-key template error")
	}

	return &i, nil
}

func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "up", &pl)
}

func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "join", &pl)
}

func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "ack", &pl)
}

func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "error", &pl)
}

func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "status", &pl)
}

func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	return i.publishEvent(ctx, pl.ApplicationId, pl.DevEui, "location", &pl)
}

func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func (i *Integration) Close() error {
	i.chPool.close()
	return i.conn.Close()
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
		return errors.Wrap(err, "publish event error")
	}

	return nil
}
