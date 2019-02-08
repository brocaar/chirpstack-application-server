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

	"github.com/brocaar/lora-app-server/internal/handler"
)

var headerNameValidator = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// HandlerConfig contains the configuration for a HTTP handler.
type HandlerConfig struct {
	Headers                 map[string]string `json:"headers"`
	DataUpURL               string            `json:"dataUpURL"`
	JoinNotificationURL     string            `json:"joinNotificationURL"`
	ACKNotificationURL      string            `json:"ackNotificationURL"`
	ErrorNotificationURL    string            `json:"errorNotificationURL"`
	StatusNotificationURL   string            `json:"statusNotificationURL"`
	LocationNotificationURL string            `json:"locationNotificationURL"`
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

func (h *Handler) sendAsync(url string, payload interface{}) {
	go func() {
		if err := h.send(url, payload); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"url": url,
			}).Error("handler/http error")
		}
	}()
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
	h.sendAsync(h.config.DataUpURL, pl)
	return nil
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
	h.sendAsync(h.config.JoinNotificationURL, pl)
	return nil
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
	h.sendAsync(h.config.ACKNotificationURL, pl)
	return nil
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
	h.sendAsync(h.config.ErrorNotificationURL, pl)
	return nil
}

// SendStatusNotification sends a status notification.
func (h *Handler) SendStatusNotification(pl handler.StatusNotification) error {
	if h.config.StatusNotificationURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.StatusNotificationURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing status notification")
	h.sendAsync(h.config.StatusNotificationURL, pl)
	return nil
}

// SendLocationNotification sends a location notification.
func (h *Handler) SendLocationNotification(pl handler.LocationNotification) error {
	if h.config.LocationNotificationURL == "" {
		return nil
	}

	log.WithFields(log.Fields{
		"url":     h.config.LocationNotificationURL,
		"dev_eui": pl.DevEUI,
	}).Info("handler/http: publishing location notification")
	h.sendAsync(h.config.LocationNotificationURL, pl)
	return nil
}
