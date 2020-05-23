package geolocation

import (
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/lorawan"
)

// TDOASingleFrameRequest implements the LoRa Cloud TDOA Single-Frame request.
type TDOASingleFrameRequest struct {
	LoRaWAN []UplinkTDOA `json:"lorawan"`
}

// TDOAMultiFrameRequest implements the LoRa Cloud TDOA Multi-Frame request.
type TDOAMultiFrameRequest struct {
	LoRaWAN [][]UplinkTDOA `json:"lorawan"`
}

// RSSISingleFrameRequest implements the LoRa Cloud RSSI Single-Frame request.
type RSSISingleFrameRequest struct {
	LoRaWAN []UplinkRSSI `json:"lorawan"`
}

// RSSIMultiFrameRequest implements the LoRa Cloud RSSI Multi-Frame request.
type RSSIMultiFrameRequest struct {
	LoRaWAN [][]UplinkRSSI `json:"lorawan"`
}

// Response implements the LoRa Cloud Response object.
type Response struct {
	Result   *LocationResult `json:"result"`
	Errors   []string        `json:"errors"`
	Warnings []string        `json:"warnings"`
}

// LocationResult implements the LoRa Cloud LocationResult object.
type LocationResult struct {
	Latitude                 float64 `json:"latitude"`
	Longitude                float64 `json:"longitude"`
	Altitude                 float64 `json:"altitude"`
	Accuracy                 int     `json:"accuracty"`
	AlgorithmType            string  `json:"algorithmType"`
	NumberOfGatewaysReceived int     `json:"numberOfGatewaysReceived"`
	NumberOfGatewaysUsed     int     `json:"numberOfGatewaysUsed"`
}

// UplinkTDOA implements the LoRa Cloud UplinkTdoa object.
type UplinkTDOA struct {
	GatewayID       lorawan.EUI64   `json:"gatewayId"`
	RSSI            float64         `json:"rssi"`
	SNR             float64         `json:"snr"`
	TOA             uint32          `json:"toa"`
	AntennaID       int             `json:"antennaId"`
	AntennaLocation AntennaLocation `json:"antennaLocation"`
}

// UplinkRSSI implements the LoRa Cloud UplinkRssi object.
type UplinkRSSI struct {
	GatewayID       lorawan.EUI64   `json:"gatewayId"`
	RSSI            float64         `json:"rssi"`
	SNR             float64         `json:"snr"`
	AntennaID       int             `json:"antennaId"`
	AntennaLocation AntennaLocation `json:"antennaLocation"`
}

// AntennaLocation implements the LoRa Cloud AntennaLocation object.
type AntennaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// NewTDOASingleFrameRequest creates a new TDOASingleFrameRequest.
func NewTDOASingleFrameRequest(rxInfo []*gw.UplinkRXInfo) TDOASingleFrameRequest {
	return TDOASingleFrameRequest{
		LoRaWAN: NewUplinkTDOA(rxInfo),
	}
}

// NewTDOAMultiFrameRequest creates a new TDOAMultiFrameRequest.
func NewTDOAMultiFrameRequest(rxInfo [][]*gw.UplinkRXInfo) TDOAMultiFrameRequest {
	var out TDOAMultiFrameRequest

	for i := range rxInfo {
		out.LoRaWAN = append(out.LoRaWAN, NewUplinkTDOA(rxInfo[i]))
	}

	return out
}

// NewRSSISingleFrameRequest creates a new RSSISingleFrameRequest.
func NewRSSISingleFrameRequest(rxInfo []*gw.UplinkRXInfo) RSSISingleFrameRequest {
	return RSSISingleFrameRequest{
		LoRaWAN: NewUplinkRSSI(rxInfo),
	}
}

// NewRSSIMultiFrameRequest creates a new RSSIMultiFrameRequest.
func NewRSSIMultiFrameRequest(rxInfo [][]*gw.UplinkRXInfo) RSSIMultiFrameRequest {
	var out RSSIMultiFrameRequest

	for i := range rxInfo {
		out.LoRaWAN = append(out.LoRaWAN, NewUplinkRSSI(rxInfo[i]))
	}

	return out
}

// NewUplinkTDOA creates a new UplinkTDOA.
func NewUplinkTDOA(rxInfo []*gw.UplinkRXInfo) []UplinkTDOA {
	var out []UplinkTDOA

	for i := range rxInfo {
		var gatewayID lorawan.EUI64
		copy(gatewayID[:], rxInfo[i].GatewayId)

		var toa uint32
		if plainTS := rxInfo[i].GetPlainFineTimestamp(); plainTS != nil {
			toa = uint32(plainTS.GetTime().Nanos)
		}

		out = append(out, UplinkTDOA{
			GatewayID: gatewayID,
			RSSI:      float64(rxInfo[i].Rssi),
			SNR:       rxInfo[i].LoraSnr,
			TOA:       toa,
			AntennaID: int(rxInfo[i].Antenna),
			AntennaLocation: AntennaLocation{
				Latitude:  rxInfo[i].GetLocation().Latitude,
				Longitude: rxInfo[i].GetLocation().Longitude,
				Altitude:  rxInfo[i].GetLocation().Altitude,
			},
		})
	}

	return out
}

// NewUplinkRSSI creates a new UplinkRSSI.
func NewUplinkRSSI(rxInfo []*gw.UplinkRXInfo) []UplinkRSSI {
	var out []UplinkRSSI

	for i := range rxInfo {
		var gatewayID lorawan.EUI64
		copy(gatewayID[:], rxInfo[i].GatewayId)

		out = append(out, UplinkRSSI{
			GatewayID: gatewayID,
			RSSI:      float64(rxInfo[i].Rssi),
			SNR:       rxInfo[i].LoraSnr,
			AntennaID: int(rxInfo[i].Antenna),
			AntennaLocation: AntennaLocation{
				Latitude:  rxInfo[i].GetLocation().Latitude,
				Longitude: rxInfo[i].GetLocation().Longitude,
				Altitude:  rxInfo[i].GetLocation().Altitude,
			},
		})
	}

	return out
}
