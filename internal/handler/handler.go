package handler

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
