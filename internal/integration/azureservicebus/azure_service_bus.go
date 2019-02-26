// Package azureservicebus implements an Azure Service-Bus integration.
package azureservicebus

import (
	"context"
	"encoding/json"
	"fmt"

	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lorawan"
)

// PublishMode defines the publish-mode type.
type PublishMode string

// Publish modes.
const (
	PublishModeTopic PublishMode = "topic"
	PublishModeQueue PublishMode = "queue"
)

// Config holds the Azure Service-Bus integration configuration.
type Config struct {
	ConnectionString string      `mapstructure:"connection_string"`
	PublishMode      PublishMode `mapstructure:"publish_mode"`
	PublishName      string      `mapstructure:"publish_name"`
}

// Integration implements an Azure Service-Bus integration.
type Integration struct {
	ctx         context.Context
	cancel      context.CancelFunc
	ns          *servicebus.Namespace
	publishName string
	topic       *servicebus.Topic
	queue       *servicebus.Queue
}

// New creates a new Azure Service-Bus integration.
func New(conf Config) (*Integration, error) {
	var err error

	i := Integration{
		ctx:         context.Background(),
		publishName: conf.PublishName,
	}
	i.ctx, i.cancel = context.WithCancel(i.ctx)

	log.Info("integration/azureservicebus: setting up namespace")
	i.ns, err = servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(conf.ConnectionString))
	if err != nil {
		return nil, errors.Wrap(err, "new namespace error")
	}

	switch conf.PublishMode {
	case PublishModeTopic:
		if err := i.setTopicClient(); err != nil {
			return nil, errors.Wrap(err, "set topic client error")
		}
	case PublishModeQueue:
		if err := i.setQueueClient(); err != nil {
			return nil, errors.Wrap(err, "set queue client error")
		}
	default:
		return nil, fmt.Errorf("unknown publish_mode: %s", conf.PublishMode)
	}

	return &i, nil
}

// SendDataUp sends an uplink data payload.
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	return i.publish("up", pl.ApplicationID, pl.DevEUI, pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	return i.publish("join", pl.ApplicationID, pl.DevEUI, pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	return i.publish("ack", pl.ApplicationID, pl.DevEUI, pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	return i.publish("error", pl.ApplicationID, pl.DevEUI, pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	return i.publish("status", pl.ApplicationID, pl.DevEUI, pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	return i.publish("location", pl.ApplicationID, pl.DevEUI, pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func (i *Integration) setTopicClient() error {
	tm := i.ns.NewTopicManager()

	log.WithField("topic", i.publishName).Info("integration/azureservicebus: testing if topic exists")
	t, err := tm.Get(i.ctx, i.publishName)
	if err != nil {
		return errors.Wrap(err, "get topic error")
	}

	if t == nil {
		log.WithField("topic", i.publishName).Info("integration/azureservicebus: topic does not exist, creating it")
		_, err := tm.Put(i.ctx, i.publishName)
		if err != nil {
			return errors.Wrap(err, "create topic error")
		}
	}

	i.topic, err = i.ns.NewTopic(i.publishName)
	if err != nil {
		return errors.Wrap(err, "new topic error")
	}

	return nil
}

func (i *Integration) setQueueClient() error {
	qm := i.ns.NewQueueManager()

	log.WithField("queue", i.publishName).Info("integration/azureservicebus: testing if queue exists")
	q, err := qm.Get(i.ctx, i.publishName)
	if err != nil {
		return errors.Wrap(err, "get queue error")
	}

	if q == nil {
		log.WithField("queue", i.publishName).Info("integration/azureservicebus: queue does not exist, creating it")
		_, err := qm.Put(i.ctx, i.publishName)
		if err != nil {
			return errors.Wrap(err, "create queue error")
		}
	}

	i.queue, err = i.ns.NewQueue(i.publishName)
	if err != nil {
		return errors.Wrap(err, "new queue error")
	}

	return nil
}

func (i *Integration) Close() error {
	log.Info("integration/azureservicebus: closing integration")
	i.cancel()
	if i.topic != nil {
		return i.topic.Close(i.ctx)
	}
	if i.queue != nil {
		return i.queue.Close(i.ctx)
	}
	return nil
}

func (i *Integration) publish(event string, applicationID int64, devEUI lorawan.EUI64, v interface{}) error {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	msg := servicebus.Message{
		ContentType: "application/json",
		Data:        jsonB,
		UserProperties: map[string]interface{}{
			"event":          event,
			"application_id": applicationID,
			"dev_eui":        devEUI.String(),
		},
	}

	if i.queue != nil {
		err = i.queue.Send(i.ctx, &msg)
	}
	if i.topic != nil {
		err = i.topic.Send(i.ctx, &msg)
	}
	if err != nil {
		return errors.Wrap(err, "send error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
	}).Info("integration/azureservicebus: event published")

	return nil
}
