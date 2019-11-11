package thingsboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

// Config holds the Thingsboard integration configuration.
type Config struct {
	Server string `json:"server"`
}

// Validate validates the Config.
func (c Config) Validate() error {
	return nil
}

// Integration implements the Thingsboard integration.
type Integration struct {
	server string
}

// New creates a new Thingsboard integration.
func New(conf Config) (*Integration, error) {
	return &Integration{
		server: conf.Server,
	}, nil
}

// SendDataUp sends the (decoded) uplink payload to the Thingsboard endpoint.
func (i *Integration) SendDataUp(ctx context.Context, vars map[string]string, pl pb.UplinkEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	accessToken, ok := vars["ThingsBoardAccessToken"]
	if !ok {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"ctx_id":  ctx.Value(logging.ContextIDKey),
		}).Warning("integration/thingsboard: device does not have a 'ThingsBoardAccessToken' variable")
		return nil
	}

	attributes := make(map[string]interface{})
	for k, v := range pl.Tags {
		attributes[k] = v
	}
	attributes["application_name"] = pl.ApplicationName
	attributes["application_id"] = strconv.FormatInt(int64(pl.ApplicationId), 10)
	attributes["device_name"] = pl.DeviceName
	attributes["dev_eui"] = devEUI

	var obj interface{}
	if pl.ObjectJson != "" {
		if err := json.Unmarshal([]byte(pl.ObjectJson), &obj); err != nil {
			return errors.Wrap(err, "unmarshal json error")
		}
	}

	telemetry := objToMap("data", obj)

	if err := i.send(accessToken, attributes, telemetry); err != nil {
		return errors.Wrap(err, "send event error")
	}

	log.WithFields(log.Fields{
		"event":   "up",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/thingsboard: attributes and telemetry uploaded")

	return nil
}

// SendJoinNotification returns nil.
func (i *Integration) SendJoinNotification(ctx context.Context, vars map[string]string, pl pb.JoinEvent) error {
	return nil
}

// SendACKNotification returns nil.
func (i *Integration) SendACKNotification(ctx context.Context, vars map[string]string, pl pb.AckEvent) error {
	return nil
}

// SendErrorNotification returns nil.
func (i *Integration) SendErrorNotification(ctx context.Context, vars map[string]string, pl pb.ErrorEvent) error {
	return nil
}

// SendStatusNotification sends the device-status fields to the Thingsboard endpoint.
func (i *Integration) SendStatusNotification(ctx context.Context, vars map[string]string, pl pb.StatusEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	accessToken, ok := vars["ThingsBoardAccessToken"]
	if !ok {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"ctx_id":  ctx.Value(logging.ContextIDKey),
		}).Warning("integration/thingsboard: device does not have a 'ThingsBoardAccessToken' variable")
		return nil
	}

	attributes := make(map[string]interface{})
	for k, v := range pl.Tags {
		attributes[k] = v
	}
	attributes["application_name"] = pl.ApplicationName
	attributes["application_id"] = strconv.FormatInt(int64(pl.ApplicationId), 10)
	attributes["device_name"] = pl.DeviceName
	attributes["dev_eui"] = devEUI

	telemetry := map[string]interface{}{
		"status_margin":                    pl.Margin,
		"status_external_power_source":     pl.ExternalPowerSource,
		"status_battery_level":             pl.BatteryLevel,
		"status_battery_level_unavailable": pl.BatteryLevelUnavailable,
	}
	if err := i.send(accessToken, attributes, telemetry); err != nil {
		return errors.Wrap(err, "send event error")
	}

	log.WithFields(log.Fields{
		"event":   "status",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/thingsboard: attributes and telemetry uploaded")

	return nil
}

// SendLocationNotification sends the device location to the Thingsboard endpoint.
func (i *Integration) SendLocationNotification(ctx context.Context, vars map[string]string, pl pb.LocationEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	accessToken, ok := vars["ThingsBoardAccessToken"]
	if !ok {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"ctx_id":  ctx.Value(logging.ContextIDKey),
		}).Warning("integration/thingsboard: device does not have a 'ThingsBoardAccessToken' variable")
		return nil
	}

	attributes := make(map[string]interface{})
	for k, v := range pl.Tags {
		attributes[k] = v
	}
	attributes["application_name"] = pl.ApplicationName
	attributes["application_id"] = strconv.FormatInt(int64(pl.ApplicationId), 10)
	attributes["device_name"] = pl.DeviceName
	attributes["dev_eui"] = devEUI

	telemetry := map[string]interface{}{
		"location_latitude":  pl.GetLocation().GetLatitude(),
		"location_longitude": pl.GetLocation().GetLongitude(),
		"location_altitude":  pl.GetLocation().GetAltitude(),
	}

	if err := i.send(accessToken, attributes, telemetry); err != nil {
		return errors.Wrap(err, "send event error")
	}

	log.WithFields(log.Fields{
		"event":   "location",
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/thingsboard: attributes and telemetry uploaded")

	return nil
}

// DataDownChan returns nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

// Close returns nil.
func (i *Integration) Close() error {
	return nil
}

func (i *Integration) send(token string, attributes, telemetry map[string]interface{}) error {
	calls := []struct {
		payload  map[string]interface{}
		endpoint string
	}{
		{
			payload:  attributes,
			endpoint: "%s/api/v1/%s/attributes",
		},
		{
			payload:  telemetry,
			endpoint: "%s/api/v1/%s/telemetry",
		},
	}

	for _, call := range calls {
		b, err := json.Marshal(call.payload)
		if err != nil {
			return errors.Wrap(err, "marshal json error")
		}

		url := fmt.Sprintf(call.endpoint, i.server, token)
		req, err := http.NewRequest("POST", url, bytes.NewReader(b))
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
	}

	return nil
}

func objToMap(prefix string, v interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	if v == nil {
		return out
	}

	switch o := v.(type) {
	case int, uint, float32, float64, uint8, int8, uint16, int16, uint32, int32, uint64, int64, string, bool:
		out[prefix] = o
	default:
		switch reflect.TypeOf(o).Kind() {
		case reflect.Map:
			v := reflect.ValueOf(o)
			keys := v.MapKeys()

			for _, k := range keys {
				keyName := fmt.Sprintf("%v", k.Interface())

				for k, v := range objToMap(prefix+"_"+keyName, v.MapIndex(k).Interface()) {
					out[k] = v
				}
			}

		case reflect.Struct:
			v := reflect.ValueOf(o)
			l := v.NumField()

			for i := 0; i < l; i++ {
				if !v.Field(i).CanInterface() {
					continue
				}

				fieldName := v.Type().Field(i).Tag.Get("influxdb")
				if fieldName == "" {
					fieldName = strings.ToLower(v.Type().Field(i).Name)
				}

				for k, v := range objToMap(prefix+"_"+fieldName, v.Field(i).Interface()) {
					out[k] = v
				}
			}

		case reflect.Ptr:
			v := reflect.Indirect(reflect.ValueOf(o))
			for k, v := range objToMap(prefix, v.Interface()) {
				out[k] = v
			}

		default:
			log.WithField("type_name", fmt.Sprintf("%T", o)).Warning("integration/thingsboard: unhandled type!")
		}
	}

	return out
}
