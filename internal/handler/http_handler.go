package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// HTTPHandlerConfig contains the configuration for a HTTP handler.
type HTTPHandlerConfig struct {
	Headers              map[string]string `json:"headers"`
	DataUpURL            string            `json:"dataUpURL"`
	JoinNotificationURL  string            `json:"joinNotificationURL"`
	ACKNotificationURL   string            `json:"ackNotificationURL"`
	ErrorNotificationURL string            `json:"errorNotificationURL"`
}

// HTTPHandler implements a HTTP handler for sending and notifying a HTTP
// endpoint.
type HTTPHandler struct {
	config HTTPHandlerConfig
}

// NewHTTPHandler creates a new HTTPHandler.
func NewHTTPHandler(conf HTTPHandlerConfig) (HTTPHandler, error) {
	return HTTPHandler{
		config: conf,
	}, nil
}

func (h *HTTPHandler) send(url string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	for k, v := range h.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}

	// check that response is in 200 range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("expected 2XX response, got: %d", resp.StatusCode)
	}

	return nil
}

// Close closes the handler.
func (h *HTTPHandler) Close() error {
	return nil
}

// SendDataUp sends a data-up payload.
func (h *HTTPHandler) SendDataUp(pl DataUpPayload) error {
	return h.send(h.config.DataUpURL, pl)
}

// SendJoinNotification sends a join notification.
func (h *HTTPHandler) SendJoinNotification(pl JoinNotification) error {
	return h.send(h.config.JoinNotificationURL, pl)
}

// SendACKNotification sends an ACK notification.
func (h *HTTPHandler) SendACKNotification(pl ACKNotification) error {
	return h.send(h.config.ACKNotificationURL, pl)
}

// SendErrorNotification sends an error notification.
func (h *HTTPHandler) SendErrorNotification(pl ErrorNotification) error {
	return h.send(h.config.ErrorNotificationURL, pl)
}
