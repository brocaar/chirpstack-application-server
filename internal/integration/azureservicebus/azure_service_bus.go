// Package azureservicebus implements an Azure Service-Bus integration.
package azureservicebus

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Integration implements an Azure Service-Bus integration.
type Integration struct {
	sync.RWMutex

	marshaler   marshaler.Type
	publishName string
	publishMode config.AzurePublishMode

	uri     string
	keyName string
	key     string
}

// New creates a new Azure Service-Bus integration.
func New(m marshaler.Type, conf config.IntegrationAzureConfig) (*Integration, error) {
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

	kv, err := parseConnectionString(conf.ConnectionString)
	if err != nil {
		return nil, errors.Wrap(err, "parse connection string error")
	}

	i := Integration{
		marshaler:   m,
		publishName: conf.PublishName,
		publishMode: conf.PublishMode,

		keyName: kv["SharedAccessKeyName"],
		key:     kv["SharedAccessKey"],
	}

	i.uri = fmt.Sprintf("https://%s%s",
		strings.Replace(kv["Endpoint"], "sb://", "", 1),
		conf.PublishName,
	)

	return &i, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	return i.publishHTTP(ctx, "up", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return i.publishHTTP(ctx, "join", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return i.publishHTTP(ctx, "ack", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return i.publishHTTP(ctx, "error", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	return i.publishHTTP(ctx, "status", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return i.publishHTTP(ctx, "location", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return i.publishHTTP(ctx, "txack", pl.ApplicationId, pl.DevEui, &pl)
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return i.publishHTTP(ctx, "integration", pl.ApplicationId, pl.DevEui, &pl)
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close closes the integration.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) publishHTTP(ctx context.Context, event string, applicationID uint64, devEUIB []byte, v proto.Message) error {
	b, err := marshaler.Marshal(i.marshaler, v)
	if err != nil {
		return errors.Wrap(err, "marshal event error")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], devEUIB)

	token, err := createSASToken(i.uri, i.keyName, i.key, time.Now().Add(time.Minute*5))
	if err != nil {
		return errors.Wrap(err, "create sas token error")
	}

	ctxCancel, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxCancel, "POST", i.uri+"/messages", bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	req.Header.Set("Authorization", token)

	if i.marshaler == marshaler.Protobuf {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("event", fmt.Sprintf("\"%s\"", event))
	req.Header.Set("application_id", fmt.Sprintf("%d", applicationID))
	req.Header.Set("dev_eui", fmt.Sprintf("\"%s\"", devEUI))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("expected 2xx response, got: %d (%s)", resp.StatusCode, string(b))
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"event":   event,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/azureservicebus: event published")

	return nil
}

func parseConnectionString(str string) (map[string]string, error) {
	out := make(map[string]string)
	pairs := strings.Split(str, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("expected two items in: %+v", kv)
		}

		out[kv[0]] = kv[1]
	}

	return out, nil
}

func createSASToken(uri string, keyName, key string, expiration time.Time) (string, error) {
	keyB := []byte(key)

	encoded := url.QueryEscape(uri)
	exp := expiration.Unix()

	signature := fmt.Sprintf("%s\n%d", encoded, exp)

	mac := hmac.New(sha256.New, keyB)
	mac.Write([]byte(signature))
	hash := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	token := fmt.Sprintf("SharedAccessSignature sig=%s&se=%d&skn=%s&sr=%s",
		hash,
		exp,
		keyName,
		encoded,
	)

	return token, nil
}
