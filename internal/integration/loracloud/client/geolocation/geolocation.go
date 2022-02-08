package geolocation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-api/go/v3/common"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	log "github.com/sirupsen/logrus"
)

const (
	tdoaSingleFrameEndpoint       = `%s/api/v2/tdoa`
	tdoaMultiFrameEndpoint        = `%s/api/v2/tdoaMultiframe`
	rssiSingleFrameEndpoint       = `%s/api/v2/rssi`
	rssiMultiFrameEndpoint        = `%s/api/v2/rssiMultiframe`
	wifiTDOASingleFrameEndpoint   = `%s/api/v2/loraWifi`
	gnssLR1110SingleFrameEndpoint = `%s/api/v3/solve/gnss_lr1110_singleframe`
)

// errors
var (
	ErrNoLocation = errors.New("no location returned")
)

// Client is a LoRa Cloud Geolocation client.
type Client struct {
	uri            string
	token          string
	requestTimeout time.Duration
	migrated       bool
}

// New creates a new Geolocation client.
func New(migrated bool, uri string, token string) *Client {
	return &Client{
		uri:            uri,
		token:          token,
		requestTimeout: time.Second,
		migrated:       migrated,
	}
}

// TDOASingleFrame request.
func (c *Client) TDOASingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo) (common.Location, error) {
	req := NewTDOASingleFrameRequest(rxInfo)
	endpoint := tdoaSingleFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/tdoa"
	}
	resp, err := c.apiRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_TDOA)
}

// TDOAMultiFrame request.
func (c *Client) TDOAMultiFrame(ctx context.Context, rxInfo [][]*gw.UplinkRXInfo) (common.Location, error) {
	req := NewTDOAMultiFrameRequest(rxInfo)
	endpoint := tdoaMultiFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/tdoaMultiframe"
	}
	resp, err := c.apiRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_TDOA)
}

// RSSISingleFrame request.
func (c *Client) RSSISingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo) (common.Location, error) {
	req := NewRSSISingleFrameRequest(rxInfo)
	endpoint := rssiSingleFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/rssi"
	}
	resp, err := c.apiRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_RSSI)
}

// RSSIMultiFrame request.
func (c *Client) RSSIMultiFrame(ctx context.Context, rxInfo [][]*gw.UplinkRXInfo) (common.Location, error) {
	req := NewRSSIMultiFrameRequest(rxInfo)
	endpoint := rssiMultiFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/rssiMultiframe"
	}
	resp, err := c.apiRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_RSSI)
}

// WifiTDOASingleFrame request.
func (c *Client) WifiTDOASingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo, aps []WifiAccessPoint) (common.Location, error) {
	req := NewWifiTDOASingleFrameRequest(rxInfo, aps)
	endpoint := wifiTDOASingleFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/loraWifi"
	}
	resp, err := c.apiRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_WIFI)
}

// GNSSLR1110SingleFrame request.
func (c *Client) GNSSLR1110SingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo, useRxTime bool, pl []byte) (common.Location, error) {
	req := NewGNSSLR1110SingleFrameRequest(rxInfo, useRxTime, pl)
	endpoint := gnssLR1110SingleFrameEndpoint
	if c.migrated {
		endpoint = "%s/api/v1/solve/gnss_lr1110_singleframe"
	}
	resp, err := c.v3APIRequest(ctx, endpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseV3Response(resp, common.LocationSource_GEO_RESOLVER_GNSS)
}

func (c *Client) parseResponse(resp Response, source common.LocationSource) (common.Location, error) {
	if len(resp.Errors) != 0 {
		return common.Location{}, fmt.Errorf("api returned error(s): %s", strings.Join(resp.Errors, ", "))
	}

	if resp.Result == nil {
		return common.Location{}, ErrNoLocation
	}

	return common.Location{
		Latitude:  resp.Result.Latitude,
		Longitude: resp.Result.Longitude,
		Altitude:  resp.Result.Altitude,
		Accuracy:  uint32(resp.Result.Accuracy),
		Source:    source,
	}, nil
}

func (c *Client) parseV3Response(resp V3Response, source common.LocationSource) (common.Location, error) {
	if len(resp.Errors) != 0 {
		return common.Location{}, fmt.Errorf("api returned error(s): %s", strings.Join(resp.Errors, ", "))
	}

	if resp.Result == nil {
		return common.Location{}, ErrNoLocation
	}

	if len(resp.Result.LLH) != 3 {
		return common.Location{}, fmt.Errorf("LLH must contain 3 items, received: %d", len(resp.Result.LLH))
	}

	return common.Location{
		Source:    source,
		Latitude:  resp.Result.LLH[0],
		Longitude: resp.Result.LLH[1],
		Altitude:  resp.Result.LLH[2],
		Accuracy:  uint32(resp.Result.Accuracy),
	}, nil
}

func (c *Client) apiRequest(ctx context.Context, endpoint string, v interface{}) (Response, error) {
	endpoint = fmt.Sprintf(endpoint, c.uri)
	var resp Response

	b, err := json.Marshal(v)
	if err != nil {
		return resp, errors.Wrap(err, "json marshal error")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return resp, errors.Wrap(err, "create request error")
	}

	req.Header.Set("Content-Type", "application/json")

	if c.migrated {
		req.Header.Set("Authorization", c.token)
	} else {
		req.Header.Set("Ocp-Apim-Subscription-Key", c.token)
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	req = req.WithContext(reqCtx)

	log.WithFields(log.Fields{
		"ctx_id":   ctx.Value(logging.ContextIDKey),
		"endpoint": endpoint,
		"migrated": c.migrated,
	}).Debug("integration/das/geolocation: making API request")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "http request error")
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bb, _ := ioutil.ReadAll(httpResp.Body)
		return resp, fmt.Errorf("expected 200, got: %d (%s)", httpResp.StatusCode, string(bb))
	}

	if err = json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return resp, errors.Wrap(err, "unmarshal response error")
	}

	return resp, nil
}

func (c *Client) v3APIRequest(ctx context.Context, endpoint string, v interface{}) (V3Response, error) {
	endpoint = fmt.Sprintf(endpoint, c.uri)
	var resp V3Response

	b, err := json.Marshal(v)
	if err != nil {
		return resp, errors.Wrap(err, "json marshal error")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return resp, errors.Wrap(err, "create request error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

	reqCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	req = req.WithContext(reqCtx)

	log.WithFields(log.Fields{
		"ctx_id":   ctx.Value(logging.ContextIDKey),
		"endpoint": endpoint,
		"migrated": c.migrated,
	}).Debug("integration/das/geolocation: making API request")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "http request error")
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bb, _ := ioutil.ReadAll(httpResp.Body)
		return resp, fmt.Errorf("expected 200, got: %d (%s)", httpResp.StatusCode, string(bb))
	}

	if err = json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return resp, errors.Wrap(err, "unmarshal response error")
	}

	return resp, nil
}
