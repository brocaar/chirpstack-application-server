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
)

const (
	tdoaSingleFrameEndpoint = `%s/api/v2/tdoa`
	tdoaMultiFrameEndpoint  = `%s/api/v2/tdoaMultiframe`
	rssiSingleFrameEndpoint = `%s/api/v2/rssi`
	rssiMultiFrameEndpoint  = `%s/api/v2/rssiMultiframe`
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
}

// New creates a new Geolocation client.
func New(uri string, token string) *Client {
	return &Client{
		uri:            uri,
		token:          token,
		requestTimeout: time.Second,
	}
}

// TDOASingleFrame request.
func (c *Client) TDOASingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo) (common.Location, error) {
	req := NewTDOASingleFrameRequest(rxInfo)
	resp, err := c.apiRequest(ctx, tdoaSingleFrameEndpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_TDOA)
}

// TDOAMultiFrame request.
func (c *Client) TDOAMultiFrame(ctx context.Context, rxInfo [][]*gw.UplinkRXInfo) (common.Location, error) {
	req := NewTDOAMultiFrameRequest(rxInfo)
	resp, err := c.apiRequest(ctx, tdoaMultiFrameEndpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_TDOA)
}

// RSSISingleFrame request.
func (c *Client) RSSISingleFrame(ctx context.Context, rxInfo []*gw.UplinkRXInfo) (common.Location, error) {
	req := NewRSSISingleFrameRequest(rxInfo)
	resp, err := c.apiRequest(ctx, rssiSingleFrameEndpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_RSSI)
}

// RSSIMultiFrame request.
func (c *Client) RSSIMultiFrame(ctx context.Context, rxInfo [][]*gw.UplinkRXInfo) (common.Location, error) {
	req := NewRSSIMultiFrameRequest(rxInfo)
	resp, err := c.apiRequest(ctx, rssiMultiFrameEndpoint, req)
	if err != nil {
		return common.Location{}, errors.Wrap(err, "api request error")
	}

	return c.parseResponse(resp, common.LocationSource_GEO_RESOLVER_RSSI)
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
	req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

	reqCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	req = req.WithContext(reqCtx)
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
