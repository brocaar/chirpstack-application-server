// Package influxdb implements a InfluxDB integration.
package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mmcloughlin/geohash"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	pb "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/models"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
	"github.com/brocaar/lorawan"
)

var precisionValidator = regexp.MustCompile(`^(ns|u|ms|s|m|h)$`)

// Config contains the configuration for the InfluxDB integration.
type Config struct {
	Endpoint string `json:"endpoint"`
	Version  int    `json:"version"`

	// v1 options
	DB                  string `json:"db"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	RetentionPolicyName string `json:"retentionPolicyName"`
	Precision           string `json:"precision"`

	// v2 options
	Token        string `json:"token"`
	Organization string `json:"org"`
	Bucket       string `json:"bucket"`
}

// Validate validates the HandlerConfig data.
func (c Config) Validate() error {
	if c.Precision != "" && !precisionValidator.MatchString(c.Precision) {
		return ErrInvalidPrecision
	}
	return nil
}

type measurement struct {
	Name   string
	Tags   map[string]string
	Values map[string]interface{}
}

func (m measurement) String() string {
	var tags []string
	var values []string

	for k, v := range m.Tags {
		tags = append(tags, fmt.Sprintf("%s=%v", escapeInfluxTag(k), formatInfluxValue(escapeInfluxTag(v), false)))
	}

	for k, v := range m.Values {
		values = append(values, fmt.Sprintf("%s=%v", k, formatInfluxValue(v, true)))
	}

	// as maps are unsorted the order of tags and values is random.
	// this is not an issue for influxdb, but makes testing more complex.
	sort.Strings(tags)
	sort.Strings(values)

	return fmt.Sprintf("%s,%s %s", m.Name, strings.Join(tags, ","), strings.Join(values, ","))
}

func formatInfluxValue(v interface{}, quote bool) string {
	switch v := v.(type) {
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case int, uint, uint8, int8, uint16, int16, uint32, int32, uint64, int64:
		return fmt.Sprintf("%di", v)
	case string:
		if quote {
			return strconv.Quote(v)
		}
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// see https://docs.influxdata.com/influxdb/v1.7/write_protocols/line_protocol_tutorial/#special-characters
func escapeInfluxTag(str string) string {
	replace := map[string]string{
		",": `\,`,
		"=": `\=`,
		" ": `\ `,
	}

	for k, v := range replace {
		str = strings.ReplaceAll(str, k, v)
	}

	return str
}

// Integration implements an InfluxDB integration.
type Integration struct {
	config Config
}

// New creates a new InfluxDB integration.
func New(conf Config) (*Integration, error) {
	return &Integration{
		config: conf,
	}, nil
}

func (i *Integration) send(measurements []measurement) error {
	var measStr []string
	for _, m := range measurements {
		measStr = append(measStr, m.String())
	}
	sort.Strings(measStr)

	b := []byte(strings.Join(measStr, "\n"))

	args := url.Values{}

	if i.config.Version == 2 {
		args.Set("org", i.config.Organization)
		args.Set("bucket", i.config.Bucket)
	} else {
		// Use else as version is a new field which might not be set.
		// It is safe to assume that in this case v1 must be used.
		args.Set("db", i.config.DB)
		args.Set("precision", i.config.Precision)
		args.Set("rp", i.config.RetentionPolicyName)
	}

	req, err := http.NewRequest("POST", i.config.Endpoint+"?"+args.Encode(), bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "text/plain")

	if i.config.Version == 2 {
		req.Header.Set("Authorization", "Token "+i.config.Token)
	} else {
		// Use else as version is a new field which might not be set.
		// It is safe to assume that in this case v1 must be used.
		if i.config.Username != "" || i.config.Password != "" {
			req.SetBasicAuth(i.config.Username, i.config.Password)
		}
	}

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

// Close closes the handler.
func (i *Integration) Close() error {
	return nil
}

// HandleUplinkEvent writes the uplink into InfluxDB.
func (i *Integration) HandleUplinkEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.UplinkEvent) error {
	if pl.ObjectJson == "" {
		return nil
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(pl.ObjectJson), &obj); err != nil {
		return errors.Wrap(err, "unmarshal json error")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	tags := map[string]string{
		"application_name": pl.ApplicationName,
		"device_name":      pl.DeviceName,
		"dev_eui":          devEUI.String(),
		"dr":               strconv.FormatInt(int64(pl.Dr), 10),
		"frequency":        strconv.FormatInt(int64(pl.GetTxInfo().GetFrequency()), 10),
	}
	for k, v := range pl.Tags {
		tags[k] = v
	}

	var measurements []measurement

	// add data-rate measurement
	measurements = append(measurements, measurement{
		Name: "device_uplink",
		Tags: tags,
		Values: map[string]interface{}{
			"value": 1,
			"f_cnt": pl.FCnt,
		},
	})

	if len(pl.RxInfo) != 0 {
		var rssi int32
		for i, rxInfo := range pl.RxInfo {
			if i == 0 || rxInfo.Rssi > rssi {
				rssi = rxInfo.Rssi
			}
		}

		var snr float64
		for i, rxInfo := range pl.RxInfo {
			if i == 0 || rxInfo.LoraSnr > snr {
				snr = rxInfo.LoraSnr
			}
		}

		measurements[0].Values["rssi"] = rssi
		measurements[0].Values["snr"] = snr
	}

	// parse object to measurements
	measurements = append(measurements, objectToMeasurements(pl, "device_frmpayload_data", obj)...)

	if len(measurements) == 0 {
		return nil
	}

	if err := i.send(measurements); err != nil {
		return errors.Wrap(err, "sending measurements error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/influxdb: uplink measurements written")

	return nil
}

// HandleStatusEvent writes the device-status into InfluxDB.
func (i *Integration) HandleStatusEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.StatusEvent) error {
	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	tags := map[string]string{
		"application_name": pl.ApplicationName,
		"device_name":      pl.DeviceName,
		"dev_eui":          devEUI.String(),
	}
	for k, v := range pl.Tags {
		tags[k] = v
	}

	var measurements []measurement

	if !pl.ExternalPowerSource && !pl.BatteryLevelUnavailable {
		measurements = append(measurements, measurement{
			Name: "device_status_battery_level",
			Tags: tags,
			Values: map[string]interface{}{
				"value": pl.BatteryLevel,
			},
		})
	}

	measurements = append(measurements, measurement{
		Name: "device_status_margin",
		Tags: tags,
		Values: map[string]interface{}{
			"value": pl.Margin,
		},
	})

	if err := i.send(measurements); err != nil {
		return errors.Wrap(err, "sending measurements error")
	}

	log.WithFields(log.Fields{
		"dev_eui": devEUI,
		"ctx_id":  ctx.Value(logging.ContextIDKey),
	}).Info("integration/influxdb: status measurements written")

	return nil
}

// HandleJoinEvent is not implemented.
func (i *Integration) HandleJoinEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.JoinEvent) error {
	return nil
}

// HandleAckEvent is not implemented.
func (i *Integration) HandleAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.AckEvent) error {
	return nil
}

// HandleErrorEvent is not implemented.
func (i *Integration) HandleErrorEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.ErrorEvent) error {
	return nil
}

// HandleLocationEvent is not implemented.
func (i *Integration) HandleLocationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.LocationEvent) error {
	return nil
}

// HandleTxAckEvent is not implemented.
func (i *Integration) HandleTxAckEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.TxAckEvent) error {
	return nil
}

// HandleIntegrationEvent is not implemented.
func (i *Integration) HandleIntegrationEvent(ctx context.Context, _ models.Integration, vars map[string]string, pl pb.IntegrationEvent) error {
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan models.DataDownPayload {
	return nil
}

func objectToMeasurements(pl pb.UplinkEvent, prefix string, obj interface{}) []measurement {
	var out []measurement

	if obj == nil {
		return out
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	switch o := obj.(type) {
	case int, uint, float32, float64, uint8, int8, uint16, int16, uint32, int32, uint64, int64, string, bool:
		tags := map[string]string{
			"application_name": pl.ApplicationName,
			"device_name":      pl.DeviceName,
			"dev_eui":          devEUI.String(),
			"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
		}
		for k, v := range pl.Tags {
			tags[k] = v
		}

		out = append(out, measurement{
			Name: prefix,
			Tags: tags,
			Values: map[string]interface{}{
				"value": o,
			},
		})

	default:
		switch reflect.TypeOf(o).Kind() {
		case reflect.Map:
			v := reflect.ValueOf(o)
			keys := v.MapKeys()

			out = append(out, mapToLocation(pl, prefix, v)...)

			for _, k := range keys {
				keyName := fmt.Sprintf("%v", k.Interface())
				if _, ignore := map[string]struct{}{
					"latitude":  struct{}{},
					"longitude": struct{}{},
				}[keyName]; ignore {
					continue
				}

				out = append(out, objectToMeasurements(pl, prefix+"_"+keyName, v.MapIndex(k).Interface())...)
			}

		case reflect.Struct:
			v := reflect.ValueOf(o)
			l := v.NumField()

			out = append(out, structToLocation(pl, prefix, v)...)

			for i := 0; i < l; i++ {
				if !v.Field(i).CanInterface() {
					continue
				}

				fieldName := v.Type().Field(i).Tag.Get("influxdb")
				if fieldName == "" {
					fieldName = strings.ToLower(v.Type().Field(i).Name)
				}

				if _, ignore := map[string]struct{}{
					"latitude":  struct{}{},
					"longitude": struct{}{},
				}[fieldName]; ignore {
					continue
				}

				out = append(out, objectToMeasurements(pl, prefix+"_"+fieldName, v.Field(i).Interface())...)
			}

		case reflect.Ptr:
			v := reflect.Indirect(reflect.ValueOf(o))
			out = append(out, objectToMeasurements(pl, prefix, v.Interface())...)

		default:
			log.WithField("type_name", fmt.Sprintf("%T", o)).Warning("influxdb integration: unhandled type!")
		}

	}

	return out
}

func mapToLocation(pl pb.UplinkEvent, prefix string, obj reflect.Value) []measurement {
	var latFloat, longFloat float64

	keys := obj.MapKeys()
	for _, k := range keys {
		if strings.ToLower(k.String()) == "latitude" {
			switch v := obj.MapIndex(k).Interface().(type) {
			case float32:
				latFloat = float64(v)
			case float64:
				latFloat = v
			}
		}

		if strings.ToLower(k.String()) == "longitude" {
			switch v := obj.MapIndex(k).Interface().(type) {
			case float32:
				longFloat = float64(v)
			case float64:
				longFloat = v
			}
		}
	}

	if latFloat == 0 && longFloat == 0 {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	tags := map[string]string{
		"application_name": pl.ApplicationName,
		"device_name":      pl.DeviceName,
		"dev_eui":          devEUI.String(),
		"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
	}
	for k, v := range pl.Tags {
		tags[k] = v
	}

	return []measurement{
		{
			Name: prefix + "_location",
			Tags: tags,
			Values: map[string]interface{}{
				"latitude":  latFloat,
				"longitude": longFloat,
				"geohash":   geohash.Encode(latFloat, longFloat),
			},
		},
	}
}

func structToLocation(pl pb.UplinkEvent, prefix string, obj reflect.Value) []measurement {
	var latFloat, longFloat float64

	l := obj.NumField()
	for i := 0; i < l; i++ {
		fieldName := strings.ToLower(obj.Type().Field(i).Name)
		if fieldName == "latitude" {
			switch v := obj.Field(i).Interface().(type) {
			case float32:
				latFloat = float64(v)
			case float64:
				latFloat = v
			}
		}

		if fieldName == "longitude" {
			switch v := obj.Field(i).Interface().(type) {
			case float32:
				longFloat = float64(v)
			case float64:
				longFloat = v
			}
		}
	}

	if latFloat == 0 && longFloat == 0 {
		return nil
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], pl.DevEui)

	tags := map[string]string{
		"application_name": pl.ApplicationName,
		"device_name":      pl.DeviceName,
		"dev_eui":          devEUI.String(),
		"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
	}
	for k, v := range pl.Tags {
		tags[k] = v
	}

	return []measurement{
		{
			Name: prefix + "_location",
			Tags: tags,
			Values: map[string]interface{}{
				"latitude":  latFloat,
				"longitude": longFloat,
				"geohash":   geohash.Encode(latFloat, longFloat),
			},
		},
	}
}
