package gcppubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements a GCP Pub/Sub integration.
type Integration struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	marshaler marshaler.Type
	client    *pubsub.Client
	topic     *pubsub.Topic
}

// New creates a new Pub/Sub integration.
func New(m marshaler.Type, conf config.IntegrationGCPConfig) (*Integration, error) {
	i := Integration{
		marshaler: m,
		ctx:       context.Background(),
	}
	var err error
	var o []option.ClientOption

	i.ctx, i.cancel = context.WithCancel(i.ctx)

	if conf.CredentialsFile != "" {
		o = append(o, option.WithCredentialsFile(conf.CredentialsFile))
	}

	log.Info("integration/gcp_pub_sub: setting up client")
	i.client, err = pubsub.NewClient(i.ctx, conf.ProjectID, o...)
	if err != nil {
		return nil, errors.Wrap(err, "new pubsub client error")
	}

	log.WithField("topic", conf.TopicName).Info("integration/gcp_pub_sub: setup topic")
	i.topic = i.client.Topic(conf.TopicName)
	ok, err := i.topic.Exists(i.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "topic exists error")
	}
	if !ok {
		return nil, fmt.Errorf("topic %s does not exist", conf.TopicName)
	}

	return &i, nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	log.Info("integration/gcppubsub: closing integration")
	i.cancel()
	return i.client.Close()
}

// SendDataUp sends an uplink data payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publish(ctx, "up", pl.DevEui, &pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	return i.publish(ctx, "join", pl.DevEui, &pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	return i.publish(ctx, "ack", pl.DevEui, &pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publish(ctx, "error", pl.DevEui, &pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	return i.publish(ctx, "status", pl.DevEui, &pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	return i.publish(ctx, "location", pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func (i *Integration) publish(ctx context.Context, event string, devEUIB []byte, msg proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return errors.Wrap(err, "marshal error")
	}

	res := i.topic.Publish(ctx, &pubsub.Message{
		Data: b,
		Attributes: map[string]string{
			"event":  event,
			"devEUI": devEUI.String(),
		},
	})
	if _, err := res.Get(i.ctx); err != nil {
		return errors.Wrap(err, "get publish result error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/gcppubsub: event published")

	return nil
}
