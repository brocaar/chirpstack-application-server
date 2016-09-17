package handler

import "github.com/brocaar/lorawan"

// Handler defines the interface of a handler backend.
type Handler interface {
	Close() error                                                                        // closes the handler
	SendDataUp(appEUI, devEUI lorawan.EUI64, payload DataUpPayload) error                // send data-up payload
	SendJoinNotification(appEUI, devEUI lorawan.EUI64, payload JoinNotification) error   // send join notification
	SendACKNotification(appEUI, devEUI lorawan.EUI64, payload ACKNotification) error     // send ack notification
	SendErrorNotification(appEUI, devEUI lorawan.EUI64, payload ErrorNotification) error // send error notification
	DataDownChan() chan DataDownPayload                                                  // returns DataDownPayload channel
}
