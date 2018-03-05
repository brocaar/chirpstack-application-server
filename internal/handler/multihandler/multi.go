// Package multihandler provides a multi handler.
package multihandler

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/gusseleet/lora-app-server/internal/handler/httphandler"
	"github.com/gusseleet/lora-app-server/internal/storage"
)

// Handler kinds
const (
	HTTPHandlerKind = "HTTP"
)

// Handler wraps multiple handlers inside a single handler so that
// data can be sent to multiple endpoints simultaneously.
// Note that errors are logged, but not returned.
type Handler struct {
	defaultHandler handler.Handler
}

// SendDataUp sends a data-up payload.
func (w Handler) SendDataUp(pl handler.DataUpPayload) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []handler.IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendDataUp(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// SendJoinNotification sends a join notification.
func (w Handler) SendJoinNotification(pl handler.JoinNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []handler.IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendJoinNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// SendACKNotification sends an ACK notification.
func (w Handler) SendACKNotification(pl handler.ACKNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []handler.IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendACKNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// SendErrorNotification sends an error notification.
func (w Handler) SendErrorNotification(pl handler.ErrorNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []handler.IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendErrorNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// Close closes the handlers.
func (w Handler) Close() error {
	return w.defaultHandler.Close()
}

// getHandlersForApplicationID returns all handlers (including the default
// handler for the given application ID.
func (w Handler) getHandlersForApplicationID(id int64) ([]handler.IntegrationHandler, error) {
	handlers := []handler.IntegrationHandler{w.defaultHandler}

	// read integrations
	integrations, err := storage.GetIntegrationsForApplicationID(config.C.PostgreSQL.DB, id)
	if err != nil {
		return nil, errors.Wrap(err, "get integrtions for application id error")
	}

	// map integration to handler + config
	for _, intg := range integrations {
		switch intg.Kind {
		case HTTPHandlerKind:
			var conf httphandler.HandlerConfig
			if err := json.NewDecoder(bytes.NewReader(intg.Settings)).Decode(&conf); err != nil {
				return nil, errors.Wrap(err, "decode http handler config error")
			}
			h, err := httphandler.NewHandler(conf)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, h)
		default:
			return nil, fmt.Errorf("unknown integration %s", intg.Kind)
		}
	}

	return handlers, nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (w Handler) DataDownChan() chan handler.DataDownPayload {
	return w.defaultHandler.DataDownChan()
}

// NewHandler returns a new MultiHandler.
func NewHandler(defaultHandler handler.Handler) handler.Handler {
	return Handler{
		defaultHandler: defaultHandler,
	}
}
