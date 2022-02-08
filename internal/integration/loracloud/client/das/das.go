package das

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	uplinkSendEndpoint = `%s/api/v1/uplink/send`
)

// Client is a LoRa Cloud DAS client.
type Client struct {
	uri            string
	token          string
	requestTimeout time.Duration
}

// New creates a new DAS client.
func New(uri string, token string) *Client {
	return &Client{
		uri:            uri,
		token:          token,
		requestTimeout: time.Second,
	}
}

// UplinkSend request.
func (c *Client) UplinkSend(ctx context.Context, req UplinkRequest) (UplinkResponse, error) {
	var resp UplinkResponse
	err := c.apiRequest(ctx, uplinkSendEndpoint, req, &resp)
	if err != nil {
		return resp, errors.Wrap(err, "api request error")
	}
	return resp, nil
}

func (c *Client) apiRequest(ctx context.Context, endpoint string, v, resp interface{}) error {
	endpoint = fmt.Sprintf(endpoint, c.uri)

	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "create request error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token)

	reqCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	req = req.WithContext(reqCtx)

	log.WithFields(log.Fields{
		"ctx_id":   ctx.Value(logging.ContextIDKey),
		"endpoint": endpoint,
	}).Debug("integration/das/das: making API request")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http request error")
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bb, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("expected 200, got: %d (%s)", httpResp.StatusCode, string(bb))
	}

	if err = json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return errors.Wrap(err, "unmarshal response error")
	}

	return nil
}
