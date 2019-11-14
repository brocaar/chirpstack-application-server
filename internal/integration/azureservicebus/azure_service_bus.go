// Package azureservicebus implements an Azure Service-Bus integration.
package azureservicebus

import (
	"context"
	"fmt"
	"sync"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements an Azure Service-Bus integration.
type Integration struct {
	sync.RWMutex

	marshaler   marshaler.Type
	ctx         context.Context
	cancel      context.CancelFunc
	ns          *servicebus.Namespace
	publishName string
	publishMode config.AzurePublishMode
	topic       *servicebus.Topic
	queue       *servicebus.Queue
}

// New creates a new Azure Service-Bus integration.
func New(m marshaler.Type, conf config.IntegrationAzureConfig) (*Integration, error) {
	var err error

	i := Integration{
		marshaler:   m,
		ctx:         context.Background(),
		publishName: conf.PublishName,
		publishMode: conf.PublishMode,
	}
	i.ctx, i.cancel = context.WithCancel(i.ctx)

	log.Info("integration/azureservicebus: setting up namespace")
	i.ns, err = servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(conf.ConnectionString))
	if err != nil {
		return nil, errors.Wrap(err, "new namespace error")
	}

	if err := i.setup(); err != nil {
		return nil, errors.Wrap(err, "setup client error")
	}

	return &i, nil
}

// SendDataUp sends an uplink data payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publishRetry(ctx, "up", pl.ApplicationId, pl.DevEui, &pl)
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	return i.publishRetry(ctx, "join", pl.ApplicationId, pl.DevEui, &pl)
}

// SendACKNotification sends an ack notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	return i.publishRetry(ctx, "ack", pl.ApplicationId, pl.DevEui, &pl)
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publishRetry(ctx, "error", pl.ApplicationId, pl.DevEui, &pl)
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	return i.publishRetry(ctx, "status", pl.ApplicationId, pl.DevEui, &pl)
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	return i.publishRetry(ctx, "location", pl.ApplicationId, pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func (i *Integration) Close() error {
	log.Info("integration/azureservicebus: closing integration")
	i.cancel()
	return i.close()
}

func (i *Integration) reconnectLoop() {
	i.Lock()
	defer i.Unlock()

	// try to close, but do not retry on error as we might not be able to
	// close the client
	if err := i.close(); err != nil {
		log.WithError(err).Error("integration/azureservicebus: close client error")
	}

	// try to setup the client and retry on error
	for {
		if err := i.setup(); err != nil {
			log.WithError(err).Error("integration/azureservicebus: setup client error")
			time.Sleep(time.Second)
			continue
		}

		break
	}
}

func (i *Integration) setup() error {
	switch i.publishMode {
	case config.AzurePublishModeTopic:
		if err := i.setTopicClient(); err != nil {
			return errors.Wrap(err, "set topic client error")
		}
	case config.AzurePublishModeQueue:
		if err := i.setQueueClient(); err != nil {
			return errors.Wrap(err, "set queue client error")
		}
	default:
		return fmt.Errorf("unknown publish_mode: %s", i.publishMode)
	}

	return nil
}

func (i *Integration) close() error {
	if i.topic != nil {
		return i.topic.Close(i.ctx)
	}
	if i.queue != nil {
		return i.queue.Close(i.ctx)
	}
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

func (i *Integration) publishRetry(ctx context.Context, event string, applicationID uint64, devEUIB []byte, v proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	for {
		if err := i.publish(ctx, event, applicationID, devEUIB, v); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"dev_eui": devEUI,
				"event":   event,
				"ctx_id":  ctx.Value(logging.ContextIDKey),
			}).Error("integration/azureservicebus: publish event error, will reconnect and retry")
			i.reconnectLoop()
			time.Sleep(time.Second)
			continue
		}

		return nil
	}
}

func (i *Integration) publish(ctx context.Context, event string, applicationID uint64, devEUIB []byte, v proto.Message) error {
	i.RLock()
	defer i.RUnlock()

	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	b, err := marshaler.Marshal(i.marshaler, v)
	if err != nil {
		return errors.Wrap(err, "marshal error")
	}

	msg := servicebus.Message{
		ContentType: "application/json",
		Data:        b,
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
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/azureservicebus: event published")

	return nil
}
