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

	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lorawan"
)

// Config holds the GCP Pub/Sub integration configuration.
type Config struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	ProjectID       string `mapstructure:"project_id"`
	TopicName       string `mapstructure:"topic_name"`
}

// Handler implements a Google Cloud Pub/Sub handler.
type Handler struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	client *pubsub.Client
	topic  *pubsub.Topic
}

// NewHandler creates a new Pub/Sub handler.
func NewHandler(conf Config) (handler.Handler, error) {
	h := Handler{
		ctx: context.Background(),
	}
	var err error
	var o []option.ClientOption

	h.ctx, h.cancel = context.WithCancel(h.ctx)

	if conf.CredentialsFile != "" {
		o = append(o, option.WithCredentialsFile(conf.CredentialsFile))
	}

	log.Info("handler/gcp_pub_sub: setting up client")
	h.client, err = pubsub.NewClient(h.ctx, conf.ProjectID, o...)
	if err != nil {
		return nil, errors.Wrap(err, "new pubsub client error")
	}

	log.WithField("topic", conf.TopicName).Info("handler/gcp_pub_sub: setup topic")
	h.topic = h.client.Topic(conf.TopicName)
	ok, err := h.topic.Exists(h.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "topic exists error")
	}
	if !ok {
		return nil, fmt.Errorf("topic %s does not exist", conf.TopicName)
	}

	return &h, nil
}

// Close closes the handler.
func (h *Handler) Close() error {
	log.Info("handler/gcp_pub_sub: closing handler")
	h.cancel()
	return h.client.Close()
}

// SendDataUp sends an uplink data payload.
func (h *Handler) SendDataUp(pl handler.DataUpPayload) error {
	return h.publish("up", pl.DevEUI, pl)
}

// SendJoinNotification sends a join notification.
func (h *Handler) SendJoinNotification(pl handler.JoinNotification) error {
	return h.publish("join", pl.DevEUI, pl)
}

// SendACKNotification sends an ack notification.
func (h *Handler) SendACKNotification(pl handler.ACKNotification) error {
	return h.publish("ack", pl.DevEUI, pl)
}

// SendErrorNotification sends an error notification.
func (h *Handler) SendErrorNotification(pl handler.ErrorNotification) error {
	return h.publish("error", pl.DevEUI, pl)
}

// SendStatusNotification sends a status notification.
func (h *Handler) SendStatusNotification(pl handler.StatusNotification) error {
	return h.publish("status", pl.DevEUI, pl)
}

// SendLocationNotification sends a location notification.
func (h *Handler) SendLocationNotification(pl handler.LocationNotification) error {
	return h.publish("location", pl.DevEUI, pl)
}

// DataDownChan return nil.
func (h *Handler) DataDownChan() chan handler.DataDownPayload {
	return nil
}

func (h *Handler) publish(event string, devEUI lorawan.EUI64, v interface{}) error {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	res := h.topic.Publish(h.ctx, &pubsub.Message{
		Data: jsonB,
		Attributes: map[string]string{
			"event":  event,
			"devEUI": devEUI.String(),
		},
	})
	if _, err := res.Get(h.ctx); err != nil {
		return errors.Wrap(err, "get publish result error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
	}).Info("handler/gcp_pub_sub: event published")

	return nil
}
