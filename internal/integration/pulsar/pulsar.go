package pulsar

import (
	"bytes"
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"text/template"
	"time"
)

// Integration implements a Pulsar integration.
type Integration struct {
	marshaler        marshaler.Type
	client           pulsar.Client
	producer         pulsar.Producer
	eventKeyTemplate *template.Template
	config           config.IntegrationPulsarConfig
}

// New creates a new Pulsar integration.
func New(m marshaler.Type, conf config.IntegrationPulsarConfig) (*Integration, error) {
	clientOpts := pulsar.ClientOptions{
		URL:                        conf.Brokers,
		OperationTimeout:           30 * time.Second,
		ConnectionTimeout:          30 * time.Second,
		TLSTrustCertsFilePath:      conf.TLSTrustCertsFilePath,
		TLSAllowInsecureConnection: conf.TLSAllowInsecureConnection,
		MaxConnectionsPerBroker:    conf.MaxConnectionsPerBroker,
	}

	if conf.AuthType == config.PulsarAuthTypeOAuth2 {
		oauth2 := pulsar.NewAuthenticationOAuth2(map[string]string{
			"type":       "client_credentials",
			"issuerUrl":  conf.OAuth2.IssuerURL,
			"audience":   conf.OAuth2.Audience,
			"clientId":   conf.OAuth2.ClientID,
			"privateKey": conf.OAuth2.PrivateKey,
		})
		clientOpts.Authentication = oauth2
	}

	client, err := pulsar.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: conf.Topic,
	})
	if err != nil {
		return nil, err
	}

	kt, err := template.New("key").Parse(conf.EventKeyTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parse key template")
	}

	i := Integration{
		marshaler:        m,
		client:           client,
		producer:         producer,
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
	if i.producer == nil {
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
	key := keyBuf.String()

	pulsarMsg := pulsar.ProducerMessage{
		Payload: b,
		Properties: map[string]string{
			"event": event,
		},
	}
	if len(key) > 0 {
		pulsarMsg.Key = key
	}

	log.WithFields(log.Fields{
		"key":    key,
		"ctx_id": ctx.Value(logging.ContextIDKey),
	}).Info("integration/pulsar: publishing message")

	if _, err := i.producer.Send(ctx, &pulsarMsg); err != nil {
		return errors.Wrap(err, "writing message to pulsar")
	}
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close shuts down the integration, closing the Pulsar producer and client.
func (i *Integration) Close() error {
	if i.client == nil {
		return fmt.Errorf("integration/pulsar: client already closed")
	}

	if i.producer == nil {
		return fmt.Errorf("integration/pulsar: producer already closed")
	}

	i.producer.Close()
	i.producer = nil

	i.client.Close()
	i.client = nil
	return nil
}
