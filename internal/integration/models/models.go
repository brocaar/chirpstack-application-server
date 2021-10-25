package models

import (
	"encoding/json"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/gofrs/uuid"
)

// Location details.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// RXInfo contains the RX information.
type RXInfo struct {
	GatewayID lorawan.EUI64 `json:"gatewayID"`
	UplinkID  uuid.UUID     `json:"uplinkID"`
	Name      string        `json:"name"`
	Time      *time.Time    `json:"time,omitempty"`
	RSSI      int           `json:"rssi"`
	LoRaSNR   float64       `json:"loRaSNR"`
	Location  *Location     `json:"location"`
}

// TXInfo contains the TX information.
type TXInfo struct {
	Frequency int `json:"frequency"`
	DR        int `json:"dr"`
}

// DataUpPayload represents a data-up payload.
type DataUpPayload struct {
	ApplicationID     int64             `json:"applicationID,string"`
	ApplicationName   string            `json:"applicationName"`
	DeviceName        string            `json:"deviceName"`
	DeviceProfileName string            `json:"deviceProfileName"`
	DeviceProfileID   string            `json:"deviceProfileID"`
	DevEUI            lorawan.EUI64     `json:"devEUI"`
	RXInfo            []RXInfo          `json:"rxInfo,omitempty"`
	TXInfo            TXInfo            `json:"txInfo"`
	ADR               bool              `json:"adr"`
	FCnt              uint32            `json:"fCnt"`
	FPort             uint8             `json:"fPort"`
	Data              []byte            `json:"data"`
	Object            interface{}       `json:"object,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	Variables         map[string]string `json:"-"`
}

// DataDownPayload represents a data-down payload.
type DataDownPayload struct {
	ApplicationID int64           `json:"applicationID,string"`
	DevEUI        lorawan.EUI64   `json:"devEUI"`
	Confirmed     bool            `json:"confirmed"`
	FPort         uint8           `json:"fPort"`
	Data          []byte          `json:"data"`
	Object        json.RawMessage `json:"object"`
}

// JoinNotification defines the payload sent to the application on
// a JoinNotificationType event.
type JoinNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	DevAddr         lorawan.DevAddr   `json:"devAddr"`
	RXInfo          []RXInfo          `json:"rxInfo,omitempty"`
	TXInfo          TXInfo            `json:"txInfo"`
	Tags            map[string]string `json:"tags,omitempty"`
	Variables       map[string]string `json:"-"`
}

// ACKNotification defines the payload sent to the application
// on an ACK event.
type ACKNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	Acknowledged    bool              `json:"acknowledged"`
	FCnt            uint32            `json:"fCnt"`
	Tags            map[string]string `json:"tags,omitempty"`
	Variables       map[string]string `json:"-"`
}

// ErrorNotification defines the payload sent to the application
// on an error event.
type ErrorNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	Type            string            `json:"type"`
	Error           string            `json:"error"`
	FCnt            uint32            `json:"fCnt,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	Variables       map[string]string `json:"-"`
}

// StatusNotification defines the payload sent to the application
// on a device-status reporting.
type StatusNotification struct {
	ApplicationID           int64             `json:"applicationID,string"`
	ApplicationName         string            `json:"applicationName"`
	DeviceName              string            `json:"deviceName"`
	DevEUI                  lorawan.EUI64     `json:"devEUI"`
	Margin                  int               `json:"margin"`
	ExternalPowerSource     bool              `json:"externalPowerSource"`
	BatteryLevel            float32           `json:"batteryLevel"`
	BatteryLevelUnavailable bool              `json:"batteryLevelUnavailable"`
	Tags                    map[string]string `json:"tags,omitempty"`
	Variables               map[string]string `json:"-"`
}

// LocationNotification defines the payload sent to the application after
// the device location has been resolved by a geolocation-server.
type LocationNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	Location        Location          `json:"location"`
	Tags            map[string]string `json:"tags,omitempty"`
	Variables       map[string]string `json:"-"`
}

// TxAckNotification defines the payload sent to the application after
// receiving a tx ack from the network-server.
type TxAckNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	FCnt            uint32            `json:"fCnt"`
	Tags            map[string]string `json:"tags,omitempty"`
	Variables       map[string]string `json:"-"`
}

// IntegrationNotification defines the payload for the integration event.
type IntegrationNotification struct {
	ApplicationID   int64             `json:"applicationID,string"`
	ApplicationName string            `json:"applicationName"`
	DeviceName      string            `json:"deviceName"`
	DevEUI          lorawan.EUI64     `json:"devEUI"`
	Tags            map[string]string `json:"tags,omitempty"`
	Object          interface{}       `json:"object"`
}
