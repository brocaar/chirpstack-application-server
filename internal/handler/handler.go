package handler

import (
	"bytes"
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/storage"
)

// Handler kinds
const (
	HTTPHandlerKind = "HTTP"
)

// Handler defines the interface of a handler backend.
type Handler interface {
	IntegrationHandler
	DataDownChan() chan DataDownPayload // returns DataDownPayload channel
}

// IntegrationHandler defines the interface of an integration handler.
type IntegrationHandler interface {
	SendDataUp(payload DataUpPayload) error                // send data-up payload
	SendJoinNotification(payload JoinNotification) error   // send join notification
	SendACKNotification(payload ACKNotification) error     // send ack notification
	SendErrorNotification(payload ErrorNotification) error // send error notification
	Close() error                                          // closes the handler
}

// MultiHandler wraps multiple handlers inside a single handler so that
// data can be sent to multiple endpoints simultaneously.
// Note that errors are logged, but not returned.
type MultiHandler struct {
	defaultHandler Handler
	db             *sqlx.DB
}

// SendDataUp sends a data-up payload.
func (w MultiHandler) SendDataUp(pl DataUpPayload) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendDataUp(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// SendJoinNotification sends a join notification.
func (w MultiHandler) SendJoinNotification(pl JoinNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendJoinNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// SendACKNotification sends an ACK notification.
func (w MultiHandler) SendACKNotification(pl ACKNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendACKNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

func (w MultiHandler) SendErrorNotification(pl ErrorNotification) error {
	handlers, err := w.getHandlersForApplicationID(pl.ApplicationID)
	if err != nil {
		log.Errorf("get handlers for application-id error: %s", err)
		handlers = []IntegrationHandler{w.defaultHandler}
	}

	for _, h := range handlers {
		if err := h.SendErrorNotification(pl); err != nil {
			log.Errorf("handler %T error: %s", h, err)
		}
	}
	return nil
}

// Close closes the handlers.
func (w MultiHandler) Close() error {
	return w.defaultHandler.Close()
}

// getHandlersForApplicationID returns all handlers (including the default
// handler for the given application ID.
func (w MultiHandler) getHandlersForApplicationID(id int64) ([]IntegrationHandler, error) {
	handlers := []IntegrationHandler{w.defaultHandler}

	// read integrations
	integrations, err := storage.GetIntegrationsForApplicationID(w.db, id)
	if err != nil {
		return nil, errors.Wrap(err, "get integrtions for application id error")
	}

	// map integration to handler + config
	for _, intg := range integrations {
		switch intg.Kind {
		case HTTPHandlerKind:
			var conf HTTPHandlerConfig
			if err := json.NewDecoder(bytes.NewReader(intg.Settings)).Decode(&conf); err != nil {
				return nil, errors.Wrap(err, "decode http handler config error")
			}
			h, err := NewHTTPHandler(conf)
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
func (w MultiHandler) DataDownChan() chan DataDownPayload {
	return w.defaultHandler.DataDownChan()
}

// NewMultiHandler returns a new MultiHandler.
func NewMultiHandler(db *sqlx.DB, defaultHandler Handler) MultiHandler {
	return MultiHandler{
		db:             db,
		defaultHandler: defaultHandler,
	}
}
