// Package httphandler implements a HTTP handler
package httphandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/handler"
)

var headerNameValidator = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// HandlerConfig contains the configuration for a HTTP handler.
type HandlerConfig struct {
	Headers              map[string]string `json:"headers"`
	DataUpURL            string            `json:"dataUpURL"`
	JoinNotificationURL  string            `json:"joinNotificationURL"`
	ACKNotificationURL   string            `json:"ackNotificationURL"`
	ErrorNotificationURL string            `json:"errorNotificationURL"`
}

// Validate validates the HandlerConfig data.
func (c HandlerConfig) Validate() error {
	for k := range c.Headers {
		if !headerNameValidator.MatchString(k) {
			return ErrInvalidHeaderName
		}
	}
	return nil
}

// Handler implements a HTTP handler for sending and notifying a HTTP
// endpoint.
type Handler struct {
	config HandlerConfig
}

// NewHandler creates a new HTTPHandler.
func NewHandler(conf HandlerConfig) (*Handler, error) {
	return &Handler{
		config: conf,
	}, nil
}

func (h *Handler) send(url string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range h.config.Headers {
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
func (h *Handler) Close() error {
	return nil
}

// SendDataUp sends a data-up payload.
func (h *Handler) SendDataUp(pl handler.DataUpPayload) error {
	if h.config.DataUpURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.DataUpURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing data-up payload")
	return h.send(h.config.DataUpURL, pl)
}

// SendJoinNotification sends a join notification.
func (h *Handler) SendJoinNotification(pl handler.JoinNotification) error {
	if h.config.JoinNotificationURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.JoinNotificationURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing join notification")
	return h.send(h.config.JoinNotificationURL, pl)
}

// SendACKNotification sends an ACK notification.
func (h *Handler) SendACKNotification(pl handler.ACKNotification) error {
	if h.config.ACKNotificationURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.ACKNotificationURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing ack notification")
	return h.send(h.config.ACKNotificationURL, pl)
}

// SendErrorNotification sends an error notification.
func (h *Handler) SendErrorNotification(pl handler.ErrorNotification) error {
	if h.config.ErrorNotificationURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.ErrorNotificationURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing error notification")
	return h.send(h.config.ErrorNotificationURL, pl)
}
