// Package http implements a HTTP integration.
package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

var headerNameValidator = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// Config contains the configuration for the HTTP integration.
type Config struct {
	Headers                 map[string]string `json:"headers"`
	DataUpURL               string            `json:"dataUpURL"`
	JoinNotificationURL     string            `json:"joinNotificationURL"`
	ACKNotificationURL      string            `json:"ackNotificationURL"`
	ErrorNotificationURL    string            `json:"errorNotificationURL"`
	StatusNotificationURL   string            `json:"statusNotificationURL"`
	LocationNotificationURL string            `json:"locationNotificationURL"`
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
	return &Integration{
		marshaler: m,
		config:    conf,
	}, nil
}

func (i *Integration) send(url string, msg proto.Message) error {
	b, err := marshaler.Marshal(i.marshaler, msg)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
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

	resp, err := http.DefaultClient.Do(req)
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

func (i *Integration) sendEvent(ctx context.Context, event, url string, devEUI lorawan.EUI64, msg proto.Message) {
	log.WithFields(log.Fields{
		"url":     url,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
		"event":   event,
	}).Info("integration/http: publishing event")
	if err := i.send(url, msg); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"url":     url,
			"dev_eui": devEUI,
			"ctx_id":  ctx.Value(logging.ContextIDKey),
			"event":   event,
		}).Error("integration/http: publish event error")
	}
}

// Close closes the handler.
func (i *Integration) Close() error {
	return nil
}

// SendDataUp sends a data-up payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.DataUpURL) {
		i.sendEvent(ctx, "up", url, devEUI, &pl)
	}

	return nil
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.JoinNotificationURL) {
		i.sendEvent(ctx, "join", url, devEUI, &pl)
	}

	return nil
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.ACKNotificationURL) {
		i.sendEvent(ctx, "ack", url, devEUI, &pl)
	}

	return nil
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.ErrorNotificationURL) {
		i.sendEvent(ctx, "error", url, devEUI, &pl)
	}

	return nil
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.StatusNotificationURL) {
		i.sendEvent(ctx, "status", url, devEUI, &pl)
	}

	return nil
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	for _, url := range getURLs(i.config.LocationNotificationURL) {
		i.sendEvent(ctx, "location", url, devEUI, &pl)
	}

	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
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
