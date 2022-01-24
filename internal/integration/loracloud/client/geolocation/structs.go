package geolocation

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/integration/loracloud/client/helpers"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/gps"
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

// WifiTDOASingleFrameRequest implements the LoRa Cloud Wifi / TDOA Single-Frame request.
type WifiTDOASingleFrameRequest struct {
	LoRaWAN          []UplinkTDOA      `json:"lorawan"`
	WifiAccessPoints []WifiAccessPoint `json:"wifiAccessPoints"`
}

// RSSIMultiFrameRequest implements the LoRa Cloud RSSI Multi-Frame request.
type RSSIMultiFrameRequest struct {
	LoRaWAN [][]UplinkRSSI `json:"lorawan"`
}

// GNSSLR1110SingleFrameRequest implements the LoRa Cloud GNSS LR1110 Single-Frame request.
type GNSSLR1110SingleFrameRequest struct {
	Payload                 helpers.HEXBytes `json:"payload"`
	GNSSCaptureTime         *float64         `json:"gnss_capture_time,omitempty"`
	GNSSCaptureTimeAccuracy *float64         `json:"gnss_capture_time_accuracy,omitempty"`
	GNSSAssistPosition      []float64        `json:"gnss_assist_position,omitempty"`
	GNSSAssistAltitude      *float64         `json:"gnss_assist_altitude,omitempty"`
	GNSSUse2DSolver         bool             `json:"gnss_use_2D_solver,omitempty"`
}

// Response implements the LoRa Cloud Response object.
type Response struct {
	Result   *LocationResult `json:"result"`
	Errors   []string        `json:"errors"`
	Warnings []string        `json:"warnings"`
}

// V3Response implements the LoRa Cloud API v3 Response object.
type V3Response struct {
	Result   *LocationSolverResult `json:"result"`
	Errors   []string              `json:"errors"`
	Warnings []string              `json:"warnings"`
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

// LocationSolverResult implements the LoRa Cloud LocationSolverResult object.
type LocationSolverResult struct {
	ECEF           []float64 `json:"ecef"`
	LLH            []float64 `json:"llh"`
	GDOP           float64   `json:"gdop"`
	Accuracy       float64   `json:"accuracy"`
	CaptureTimeGPS float64   `json:"capture_time_gps"`
	CaptureTimeUTC float64   `json:"capture_time_utc"`
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

// WifiAccessPoint implements the LoRa Cloud WifiAccessPoints object.
type WifiAccessPoint struct {
	MacAddress     BSSID `json:"macAddress"`
	SignalStrength int   `json:"signalStrength"`
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

// NewWifiTDOASingleFrameRequest creates a new WifiTDOASingleFrameRequest.
func NewWifiTDOASingleFrameRequest(rxInfo []*gw.UplinkRXInfo, aps []WifiAccessPoint) WifiTDOASingleFrameRequest {
	out := WifiTDOASingleFrameRequest{
		LoRaWAN:          NewUplinkTDOA(rxInfo),
		WifiAccessPoints: aps,
	}

	return out
}

// NewGNSSLR1110SingleFrameRequest creates a new GNSSLR1110SingleFrameRequest.
func NewGNSSLR1110SingleFrameRequest(rxInfo []*gw.UplinkRXInfo, useRxTime bool, pl []byte) GNSSLR1110SingleFrameRequest {
	out := GNSSLR1110SingleFrameRequest{
		Payload: helpers.HEXBytes(pl),
	}

	if useRxTime {
		acc := float64(15)
		out.GNSSCaptureTimeAccuracy = &acc

		if gpsTime := helpers.GetTimeSinceGPSEpoch(rxInfo); gpsTime != nil {
			t := (float64(*gpsTime) / float64(time.Second)) - 6
			out.GNSSCaptureTime = &t
		} else {
			gpsTime := gps.Time(time.Now()).TimeSinceGPSEpoch()
			t := (float64(gpsTime) / float64(time.Second)) - 6
			out.GNSSCaptureTime = &t
		}
	}

	if loc := helpers.GetStartLocation(rxInfo); loc != nil {
		out.GNSSAssistPosition = []float64{loc.Latitude, loc.Longitude}
		out.GNSSAssistAltitude = &loc.Altitude
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

// BSSID defines a BSSID identifier,.
type BSSID [6]byte

// MarshalText implements encoding.TextMarshaler.
func (b BSSID) MarshalText() ([]byte, error) {
	var str []string
	for i := range b[:] {
		str = append(str, hex.EncodeToString([]byte{b[i]}))
	}

	return []byte(strings.Join(str, ":")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *BSSID) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ":")
	if len(parts) != 6 {
		return errors.New("bssid must be 6 bytes")
	}

	for i := range parts {
		bb, err := hex.DecodeString(parts[i])
		if err != nil {
			return errors.Wrap(err, "decode hex error")
		}
		if len(bb) != 1 {
			return errors.New("exactly 1 byte expected")
		}
		b[i] = bb[0]
	}

	return nil
}
