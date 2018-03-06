package handler

import (
	"encoding/json"
	"time"

	"github.com/gusseleet/lora-app-server/internal/codec"
	"github.com/brocaar/lorawan"
)

// DataRate contains the data-rate related fields.
type DataRate struct {
	Modulation   string `json:"modulation"`
	Bandwidth    int    `json:"bandwidth"`
	SpreadFactor int    `json:"spreadFactor,omitempty"`
	Bitrate      int    `json:"bitrate,omitempty"`
}

// RXInfo contains the RX information.
type RXInfo struct {
	MAC       lorawan.EUI64 `json:"mac"`
	Time      *time.Time    `json:"time,omitempty"`
	RSSI      int           `json:"rssi"`
	LoRaSNR   float64       `json:"loRaSNR"`
	Name      string        `json:"name"`
	Latitude  float64       `json:"latitude"`
	Longitude float64       `json:"longitude"`
	Altitude  float64       `json:"altitude"`
}

// TXInfo contains the TX information.
type TXInfo struct {
	Frequency int      `json:"frequency"`
	DataRate  DataRate `json:"dataRate"`
	ADR       bool     `json:"adr"`
	CodeRate  string   `json:"codeRate"`
}

// DataUpPayload represents a data-up payload.
type DataUpPayload struct {
	ApplicationID       int64         `json:"applicationID,string"`
	ApplicationName     string        `json:"applicationName"`
	DeviceName          string        `json:"deviceName"`
	DevEUI              lorawan.EUI64 `json:"devEUI"`
	DeviceStatusBattery *int          `json:"deviceStatusBattery,omitempty"`
	DeviceStatusMargin  *int          `json:"deviceStatusMargin,omitempty"`
	RXInfo              []RXInfo      `json:"rxInfo,omitempty"`
	TXInfo              TXInfo        `json:"txInfo"`
	FCnt                uint32        `json:"fCnt"`
	FPort               uint8         `json:"fPort"`
	Data                []byte        `json:"data"`
	Object              codec.Payload `json:"object,omitempty"`
}

// DataDownPayload represents a data-down payload.
type DataDownPayload struct {
	ApplicationID int64           `json:"applicationID,string"`
	DevEUI        lorawan.EUI64   `json:"devEUI"`
	Reference     string          `json:"reference"`
	Confirmed     bool            `json:"confirmed"`
	FPort         uint8           `json:"fPort"`
	Data          []byte          `json:"data"`
	Object        json.RawMessage `json:"object"`
}

// JoinNotification defines the payload sent to the application on
// a JoinNotificationType event.
type JoinNotification struct {
	ApplicationID   int64           `json:"applicationID,string"`
	ApplicationName string          `json:"applicationName"`
	DeviceName      string          `json:"deviceName"`
	DevEUI          lorawan.EUI64   `json:"devEUI"`
	DevAddr         lorawan.DevAddr `json:"devAddr"`
}

// ACKNotification defines the payload sent to the application
// on an ACK event.
type ACKNotification struct {
	ApplicationID   int64         `json:"applicationID,string"`
	ApplicationName string        `json:"applicationName"`
	DeviceName      string        `json:"deviceName"`
	DevEUI          lorawan.EUI64 `json:"devEUI"`
	Reference       string        `json:"reference"`
	Acknowledged    bool          `json:"acknowledged"`
	FCnt            uint32        `json:"fCnt"`
}

// ErrorNotification defines the payload sent to the application
// on an error event.
type ErrorNotification struct {
	ApplicationID   int64         `json:"applicationID,string"`
	ApplicationName string        `json:"applicationName"`
	DeviceName      string        `json:"deviceName"`
	DevEUI          lorawan.EUI64 `json:"devEUI"`
	Type            string        `json:"type"`
	Error           string        `json:"error"`
	FCnt            uint32        `json:"fCnt"`
}
