package handler

import (
	"time"

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
	MAC     lorawan.EUI64 `json:"mac"`
	Time    *time.Time    `json:"time,omitempty"`
	RSSI    int           `json:"rssi"`
	LoRaSNR float64       `json:"loRaSNR"`
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
	ApplicationID   int64         `json:"applicationID,string"`
	ApplicationName string        `json:"applicationName"`
	NodeName        string        `json:"nodeName"`
	DevEUI          lorawan.EUI64 `json:"devEUI"`
	RXInfo          []RXInfo      `json:"rxInfo"`
	TXInfo          TXInfo        `json:"txInfo"`
	FCnt            uint32        `json:"fCnt"`
	FPort           uint8         `json:"fPort"`
	Data            []byte        `json:"data"`
}

// DataDownPayload represents a data-down payload.
type DataDownPayload struct {
	ApplicationID int64         `json:"applicationID,string"`
	DevEUI        lorawan.EUI64 `json:"devEUI"`
	Reference     string        `json:"reference"`
	Confirmed     bool          `json:"confirmed"`
	FPort         uint8         `json:"fPort"`
	Data          []byte        `json:"data"`
}

// JoinNotification defines the payload sent to the application on
// a JoinNotificationType event.
type JoinNotification struct {
	ApplicationID   int64           `json:"applicationID,string"`
	ApplicationName string          `json:"applicationName"`
	NodeName        string          `json:"nodeName"`
	DevEUI          lorawan.EUI64   `json:"devEUI"`
	DevAddr         lorawan.DevAddr `json:"devAddr"`
}
