// Package http implements a HTTP integration.
package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
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

// Close closes the handler.
func (i *Integration) Close() error {
	return nil
}

// SendDataUp sends a data-up payload.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	if i.config.DataUpURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.DataUpURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing data-up payload")
	if err := i.send(i.config.DataUpURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// SendJoinNotification sends a join notification.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	if i.config.JoinNotificationURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.JoinNotificationURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing join notification")
	if err := i.send(i.config.JoinNotificationURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// SendACKNotification sends an ACK notification.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	if i.config.ACKNotificationURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.ACKNotificationURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing ack notification")
	if err := i.send(i.config.ACKNotificationURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// SendErrorNotification sends an error notification.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	if i.config.ErrorNotificationURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.ErrorNotificationURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing error notification")
	if err := i.send(i.config.ErrorNotificationURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// SendStatusNotification sends a status notification.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	if i.config.StatusNotificationURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.StatusNotificationURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing status notification")
	if err := i.send(i.config.StatusNotificationURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// SendLocationNotification sends a location notification.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	if i.config.LocationNotificationURL == "" {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	log.WithFields(log.Fields{
		"url":     i.config.LocationNotificationURL,
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/http: publishing location notification")
	if err := i.send(i.config.LocationNotificationURL, &pl); err != nil {
		return errors.Wrap(err, "send error")
	}
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}
