// Package http implements a HTTP integration.
package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

var headerNameValidator = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// Config contains the configuration for the HTTP integration.
type Config struct {
	Headers          map[string]string `json:"headers"`
	EventEndpointURL string            `json:"eventEndpointURL"`
	Marshaler        string            `json:"marshaler"`
	Timeout          time.Duration     `json:"timeout"`

	// For backwards compatibility.
	DataUpURL                  string `json:"dataUpURL"`
	JoinNotificationURL        string `json:"joinNotificationURL"`
	ACKNotificationURL         string `json:"ackNotificationURL"`
	ErrorNotificationURL       string `json:"errorNotificationURL"`
	StatusNotificationURL      string `json:"statusNotificationURL"`
	LocationNotificationURL    string `json:"locationNotificationURL"`
	TxAckNotificationURL       string `json:"txAckNotificationURL"`
	IntegrationNotificationURL string `json:"integrationNotificationURL"`
}

// Validate validates the HandlerConfig data.
func (c Config) Validate() error {
	for k := range c.Headers {
		if !headerNameValidator.MatchString(k) {
			return ErrInvalidHeaderName
		}
	}
	return nil
}

// Integration implements a HTTP integration.
type Integration struct {
	marshaler marshaler.Type
	config    Config
}

// New creates a new HTTP integration.
func New(m marshaler.Type, conf Config) (*Integration, error) {
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

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 10
	}

	return &Integration{
		marshaler: m,
		config:    conf,
	}, nil
}

func (i *Integration) send(u string, msg proto.Message) error {
	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req, err := http.NewRequest("POST", u, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	if i.marshaler == marshaler.Protobuf {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range i.config.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: i.config.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}
	defer resp.Body.Close()

	// check that response is in 200 range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("expected 2XX response, got: %d", resp.StatusCode)
	}

	return nil
}

func (i *Integration) sendEvent(ctx context.Context, eventType, u string, devEUI lorawan.EUI64, msg proto.Message) {
	uu, err := url.Parse(u)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"url":        u,
			"dev_eui":    devEUI,
			"ctx_id":     ctx.Value(logging.ContextIDKey),
			"event_type": eventType,
		}).Error("integration/http: parse url error")
		return
	}

	args := uu.Query()
	args.Set("event", eventType)
	u = fmt.Sprintf("%s://%s%s?%s", uu.Scheme, uu.Host, uu.Path, args.Encode())

	log.WithFields(log.Fields{
		"url":        u,
		"dev_eui":    devEUI,
		"ctx_id":     ctx.Value(logging.ContextIDKey),
		"event_type": eventType,
	}).Info("integration/http: publishing event")

	if err := i.send(u, msg); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"url":        u,
			"dev_eui":    devEUI,
			"ctx_id":     ctx.Value(logging.ContextIDKey),
			"event_type": eventType,
		}).Error("integration/http: publish event error")
		return
	}
}

func (i *Integration) getEventEndpointURL(eventType string) string {
	var url string

	// For backwards compatibility.
	switch eventType {
	case "up":
		url = i.config.DataUpURL
	case "join":
		url = i.config.JoinNotificationURL
	case "ack":
		url = i.config.ACKNotificationURL
	case "error":
		url = i.config.ErrorNotificationURL
	case "status":
		url = i.config.StatusNotificationURL
	case "location":
		url = i.config.LocationNotificationURL
	case "txack":
		url = i.config.TxAckNotificationURL
	case "integration":
		url = i.config.IntegrationNotificationURL
	}

	if url == "" {
		url = i.config.EventEndpointURL
	}

	return url
}

// Close closes the handler.
func (i *Integration) Close() error {
	return nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("up")) {
		i.sendEvent(ctx, "up", url, devEUI, &pl)
	}

	return nil
}

// HandleJoinEvent sends a JoinEvent.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("join")) {
		i.sendEvent(ctx, "join", url, devEUI, &pl)
	}

	return nil
}

// HandleAckEvent sends an AckEvent.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("ack")) {
		i.sendEvent(ctx, "ack", url, devEUI, &pl)
	}

	return nil
}

// HandleErrorEvent sends an ErrorEvent.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("error")) {
		i.sendEvent(ctx, "error", url, devEUI, &pl)
	}

	return nil
}

// HandleStatusEvent sends a StatusEvent.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("status")) {
		i.sendEvent(ctx, "status", url, devEUI, &pl)
	}

	return nil
}

// HandleLocationEvent sends a LocationEvent.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("location")) {
		i.sendEvent(ctx, "location", url, devEUI, &pl)
	}

	return nil
}

// HandleTxAckEvent sends a TxAckEvent.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("txack")) {
		i.sendEvent(ctx, "txack", url, devEUI, &pl)
	}

	return nil
}

// HandleIntegrationEvent sends an IntegrationEvent.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.getEventEndpointURL("integration")) {
		i.sendEvent(ctx, "integration", url, devEUI, &pl)
	}

	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

func getURLs(str string) []string {
	urls := strings.Split(str, ",")
	var out []string

	for _, url := range urls {
		if url := strings.TrimSpace(url); url != "" {
			out = append(out, url)
		}
	}
	return out
}
