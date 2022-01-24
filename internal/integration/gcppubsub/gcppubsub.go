package gcppubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

type publishRequest struct {
	Messages []message `json:"messages"`
}

type message struct {
	Attributes map[string]string `json:"attributes"`
	Data       []byte            `json:"data"`
}

// Integration implements a GCP Pub/Sub integration.
type Integration struct {
	marshaler marshaler.Type

	project             string
	topic               string
	jsonCredentialsFile []byte
	client              *http.Client
}

// New creates a new Pub/Sub integration.
func New(m marshaler.Type, conf config.IntegrationGCPConfig) (*Integration, error) {
	if conf.Marshaler != "" {
		switch conf.Marshaler {
		case "PROTOBUF":
			m = marshaler.Protobuf
		case "JSON":
			m = marshaler.ProtobufJSON
		case "JSON_V3":
			m = marshaler.JSONV3
		}
	}

	i := Integration{
		marshaler: m,
		project:   conf.ProjectID,
		topic:     conf.TopicName,
	}

	var err error
	if conf.CredentialsFile != "" {
		i.jsonCredentialsFile, err = ioutil.ReadFile(conf.CredentialsFile)
		if err != nil {
			return nil, errors.Wrap(err, "read credentials file error")
		}
	} else {
		i.jsonCredentialsFile = []byte(conf.CredentialsFileBytes)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), i.jsonCredentialsFile, "https://www.googleapis.com/auth/pubsub")
	if err != nil {
		return nil, errors.Wrap(err, "credentials from json error")
	}
	i.client = oauth2.NewClient(context.Background(), creds.TokenSource)

	return &i, nil
}

// Close is not implemented.
func (i *Integration) Close() error {
	return nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publish(ctx, "up", pl.DevEui, &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return i.publish(ctx, "join", pl.DevEui, &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return i.publish(ctx, "ack", pl.DevEui, &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publish(ctx, "error", pl.DevEui, &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return i.publish(ctx, "status", pl.DevEui, &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return i.publish(ctx, "location", pl.DevEui, &pl)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return i.publish(ctx, "txack", pl.DevEui, &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return i.publish(ctx, "integration", pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

func (i *Integration) publish(ctx context.Context, event string, devEUIB []byte, msg proto.Message) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return errors.Wrap(err, "marshal event error")
	}

	req := publishRequest{
		Messages: []message{
			{
				Attributes: map[string]string{
					"event":  event,
					"devEUI": devEUI.String(),
				},
				Data: b,
			},
		},
	}
	b, err = json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "marshal pub/sub request error")
	}

	ctxCancel, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctxCancel, "POST", fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/topics/%s:publish", i.project, i.topic), bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	resp, err := i.client.Do(httpReq)
	if err != nil {
		return errors.Wrap(err, "pub/sub request error")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("exepcted 2xx code, got: %d (%s)", resp.StatusCode, string(b))
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/gcppubsub: event published")

	return nil
}
