// Package influxdb implements a InfluxDB integration.
package influxdb

import (
	"bytes"
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

	"github.com/brocaar/lora-app-server/internal/integration"
)

var precisionValidator = regexp.MustCompile(`^(ns|u|ms|s|m|h)$`)

// Config contains the configuration for the InfluxDB integration.
type Config struct {
	Endpoint            string `json:"endpoint"`
	DB                  string `json:"db"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	RetentionPolicyName string `json:"retentionPolicyName"`
	Precision           string `json:"precision"`
}

// Validate validates the HandlerConfig data.
func (c Config) Validate() error {
	if !precisionValidator.MatchString(c.Precision) {
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
		tags = append(tags, fmt.Sprintf("%s=%v", k, formatInfluxValue(v, false)))
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
	args.Set("db", i.config.DB)
	args.Set("precision", i.config.Precision)
	args.Set("rp", i.config.RetentionPolicyName)

	req, err := http.NewRequest("POST", i.config.Endpoint+"?"+args.Encode(), bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "text/plain")

	if i.config.Username != "" || i.config.Password != "" {
		req.SetBasicAuth(i.config.Username, i.config.Password)
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

// SendDataUp stores the uplink data into InfluxDB.
func (i *Integration) SendDataUp(pl integration.DataUpPayload) error {
	if pl.Object == nil {
		return nil
	}

	var measurements []measurement

	// add data-rate measurement
	measurements = append(measurements, measurement{
		Name: "device_uplink",
		Tags: map[string]string{
			"application_name": pl.ApplicationName,
			"device_name":      pl.DeviceName,
			"dev_eui":          pl.DevEUI.String(),
			"dr":               strconv.FormatInt(int64(pl.TXInfo.DR), 10),
			"frequency":        strconv.FormatInt(int64(pl.TXInfo.Frequency), 10),
		},
		Values: map[string]interface{}{
			"value": 1,
			"f_cnt": pl.FCnt,
		},
	})

	if len(pl.RXInfo) != 0 {
		var rssi int
		for i, rxInfo := range pl.RXInfo {
			if i == 0 || rxInfo.RSSI > rssi {
				rssi = rxInfo.RSSI
			}
		}

		var snr float64
		for i, rxInfo := range pl.RXInfo {
			if i == 0 || rxInfo.LoRaSNR > snr {
				snr = rxInfo.LoRaSNR
			}
		}

		measurements[0].Values["rssi"] = rssi
		measurements[0].Values["snr"] = snr
	}

	// parse object to measurements
	measurements = append(measurements, objectToMeasurements(pl, "device_frmpayload_data", pl.Object)...)

	if len(measurements) == 0 {
		return nil
	}

	if err := i.send(measurements); err != nil {
		return errors.Wrap(err, "sending measurements error")
	}

	log.WithFields(log.Fields{
		"dev_eui": pl.DevEUI,
	}).Info("integration/influxdb: uplink measurements written")

	return nil
}

// SendStatusNotification writes the device-status.
func (i *Integration) SendStatusNotification(pl integration.StatusNotification) error {
	var measurements []measurement

	measurements = append(measurements, measurement{
		Name: "device_status_battery",
		Tags: map[string]string{
			"application_name": pl.ApplicationName,
			"device_name":      pl.DeviceName,
			"dev_eui":          pl.DevEUI.String(),
		},
		Values: map[string]interface{}{
			"value": pl.Battery,
		},
	})

	if !pl.ExternalPowerSource && !pl.BatteryLevelUnavailable {
		measurements = append(measurements, measurement{
			Name: "device_status_battery_level",
			Tags: map[string]string{
				"application_name": pl.ApplicationName,
				"device_name":      pl.DeviceName,
				"dev_eui":          pl.DevEUI.String(),
			},
			Values: map[string]interface{}{
				"value": pl.BatteryLevel,
			},
		})
	}

	measurements = append(measurements, measurement{
		Name: "device_status_margin",
		Tags: map[string]string{
			"application_name": pl.ApplicationName,
			"device_name":      pl.DeviceName,
			"dev_eui":          pl.DevEUI.String(),
		},
		Values: map[string]interface{}{
			"value": pl.Margin,
		},
	})

	if err := i.send(measurements); err != nil {
		return errors.Wrap(err, "sending measurements error")
	}

	log.WithFields(log.Fields{
		"dev_eui": pl.DevEUI,
	}).Info("integration/influxdb: status measurements written")

	return nil
}

// SendJoinNotification is not implemented.
func (i *Integration) SendJoinNotification(pl integration.JoinNotification) error {
	return nil
}

// SendACKNotification is not implemented.
func (i *Integration) SendACKNotification(pl integration.ACKNotification) error {
	return nil
}

// SendErrorNotification is not implemented.
func (i *Integration) SendErrorNotification(pl integration.ErrorNotification) error {
	return nil
}

// SendLocationNotification is not implemented.
func (i *Integration) SendLocationNotification(pl integration.LocationNotification) error {
	return nil
}

// DataDownChan return nil.
func (i *Integration) DataDownChan() chan integration.DataDownPayload {
	return nil
}

func objectToMeasurements(pl integration.DataUpPayload, prefix string, obj interface{}) []measurement {
	var out []measurement

	if obj == nil {
		return out
	}

	switch o := obj.(type) {
	case int, uint, float32, float64, uint8, int8, uint16, int16, uint32, int32, uint64, int64, string, bool:
		out = append(out, measurement{
			Name: prefix,
			Tags: map[string]string{
				"application_name": pl.ApplicationName,
				"device_name":      pl.DeviceName,
				"dev_eui":          pl.DevEUI.String(),
				"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
			},
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

func mapToLocation(pl integration.DataUpPayload, prefix string, obj reflect.Value) []measurement {
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

	return []measurement{
		{
			Name: prefix + "_location",
			Tags: map[string]string{
				"application_name": pl.ApplicationName,
				"device_name":      pl.DeviceName,
				"dev_eui":          pl.DevEUI.String(),
				"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
			},
			Values: map[string]interface{}{
				"latitude":  latFloat,
				"longitude": longFloat,
				"geohash":   geohash.Encode(latFloat, longFloat),
			},
		},
	}
}

func structToLocation(pl integration.DataUpPayload, prefix string, obj reflect.Value) []measurement {
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

	return []measurement{
		{
			Name: prefix + "_location",
			Tags: map[string]string{
				"application_name": pl.ApplicationName,
				"device_name":      pl.DeviceName,
				"dev_eui":          pl.DevEUI.String(),
				"f_port":           strconv.FormatInt(int64(pl.FPort), 10),
			},
			Values: map[string]interface{}{
				"latitude":  latFloat,
				"longitude": longFloat,
				"geohash":   geohash.Encode(latFloat, longFloat),
			},
		},
	}
}
