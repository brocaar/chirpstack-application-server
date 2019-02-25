package gcppubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lorawan"
)

// Config holds the GCP Pub/Sub integration configuration.
type Config struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	ProjectID       string `mapstructure:"project_id"`
	TopicName       string `mapstructure:"topic_name"`
}

// Integration implements a GCP Pub/Sub integration.
type Integration struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	client *pubsub.Client
	topic  *pubsub.Topic
}

// New creates a new Pub/Sub integration.
func New(conf Config) (*Integration, error) {
	i := Integration{
		ctx: context.Background(),
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
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	return i.publish("up", pl.DevEUI, pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	return i.publish("join", pl.DevEUI, pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	return i.publish("ack", pl.DevEUI, pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	return i.publish("error", pl.DevEUI, pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	return i.publish("status", pl.DevEUI, pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	return i.publish("location", pl.DevEUI, pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func (i *Integration) publish(event string, devEUI lorawan.EUI64, v interface{}) error {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	res := i.topic.Publish(i.ctx, &pubsub.Message{
		Data: jsonB,
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
	}).Info("integration/gcppubsub: event published")

	return nil
}
