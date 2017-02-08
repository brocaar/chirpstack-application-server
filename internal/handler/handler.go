package handler

// Handler defines the interface of a handler backend.
type Handler interface {
	Close() error                                          // closes the handler
	SendDataUp(payload DataUpPayload) error                // send data-up payload
	SendJoinNotification(payload JoinNotification) error   // send join notification
	SendACKNotification(payload ACKNotification) error     // send ack notification
	SendErrorNotification(payload ErrorNotification) error // send error notification
	DataDownChan() chan DataDownPayload                    // returns DataDownPayload channel
}
