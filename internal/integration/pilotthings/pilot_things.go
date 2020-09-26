package pilotthings

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/pkg/errors"
)

// Config contains the Pilot Things integration configuration.
type Config struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

// Validate verifies the config for errors.
func (cfg Config) Validate() error {
	uri, err := url.Parse(cfg.Server)
	if err != nil {
		return err
	}

	if !uri.IsAbs() {
		return errors.New("server must be absolute url")
	}

	return nil
}

// Integration is the actual implementation of the Pilot Things integration
type Integration struct {
	uplink string
}

type uplinkMetadata struct {
	Rssi    int32   `json:"rssi"`
	LoraSnr float64 `json:"lorasnr"`
	RfChain uint32  `json:"rfchain"`
	Antenna uint32  `json:"antenna"`
	Board   uint32  `json:"board"`
}

type uplinkPayload struct {
	DeviceName string           `json:"deviceName"`
	Data       string           `json:"data"`
	DevEUI     string           `json:"devEUI"`
	FPort      uint32           `json:"fPort"`
	DevAddr    string           `json:"devAddr"`
	FCnt       uint32           `json:"fcnt"`
	Metadata   []uplinkMetadata `json:"metadata"`
}

// New creates the integration from a configuration
func New(cfg Config) (*Integration, error) {
	base, err := url.Parse(cfg.Server)
	if err != nil {
		return nil, errors.Wrap(err, "parse server error")
	}

	if !base.IsAbs() {
		return nil, errors.New("server must be absolute url")
	}

	return &Integration{
		base.ResolveReference(&url.URL{
			Path:     "om2m/ipe-loraserver/up-link",
			RawQuery: "token=" + url.QueryEscape(cfg.Token),
		}).String(),
	}, nil
}

// HandleUplinkEvent sends the data to Pilot Things.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.UplinkEvent) error {
	body := uplinkPayload{
		event.DeviceName,
		hex.EncodeToString(event.Data),
		hex.EncodeToString(event.DevEui),
		event.FPort,
		hex.EncodeToString(event.DevAddr),
		event.FCnt,
		make([]uplinkMetadata, 0, len(event.RxInfo)),
	}

	for _, rxInfo := range event.RxInfo {
		body.Metadata = append(body.Metadata, uplinkMetadata{
			rxInfo.Rssi,
			rxInfo.LoraSnr,
			rxInfo.RfChain,
			rxInfo.Antenna,
			rxInfo.Board,
		})
	}

	bodyStr, err := json.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	req, err := http.NewRequest("POST", i.uplink, bytes.NewReader(bodyStr))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}
	defer resp.Body.Close()

	// check that response is in 200 range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("expected 2xx response, got: %d (%s)", resp.StatusCode, string(b))
	}

	return nil
}

// HandleJoinEvent is not implemented.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.JoinEvent) error {
	return nil
}

// HandleAckEvent is not implemented.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.AckEvent) error {
	return nil
}

// HandleErrorEvent is not implemented.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.ErrorEvent) error {
	return nil
}

// HandleStatusEvent is not implemented.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.StatusEvent) error {
	return nil
}

// HandleLocationEvent is not implemented.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.LocationEvent) error {
	return nil
}

// HandleTxAckEvent is not implemented.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.TxAckEvent) error {
	return nil
}

// HandleIntegrationEvent is not implemented.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, event integration.IntegrationEvent) error {
	return nil
}

// DataDownChan is not implemented.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

// Close is not needed.
func (i *Integration) Close() error {
	return nil
}
