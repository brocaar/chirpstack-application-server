package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func init() {
	gob.Register(ExtendedCayenneLPP{})
}

// ExtendedCayenneLPP types.
const (
	lppExtendedCounterInput   byte = 4
	lppExtendedBatteryVoltage byte = 255
)

// ExtendedCayenneLPP defines the Cayenne LPP data structure.
type ExtendedCayenneLPP struct {
	DigitalInput      map[byte]uint8         `json:"digitalInput,omitempty" influxdb:"digital_input"`
	DigitalOutput     map[byte]uint8         `json:"digitalOutput,omitempty" influxdb:"digital_output"`
	AnalogInput       map[byte]float64       `json:"analogInput,omitempty" influxdb:"analog_input"`
	AnalogOutput      map[byte]float64       `json:"analogOutput,omitempty" influxdb:"analog_output"`
	IlluminanceSensor map[byte]uint16        `json:"illuminanceSensor,omitempty" influxdb:"illuminance_sensor"`
	PresenceSensor    map[byte]uint8         `json:"presenceSensor,omitempty" influxdb:"presence_sensor"`
	TemperatureSensor map[byte]float64       `json:"temperatureSensor,omitempty" influxdb:"temperature_sensor"`
	HumiditySensor    map[byte]float64       `json:"humiditySensor,omitempty" influxdb:"humidity_sensor"`
	Accelerometer     map[byte]Accelerometer `json:"accelerometer,omitempty" influxdb:"accelerometer"`
	Barometer         map[byte]float64       `json:"barometer,omitempty" influxdb:"barometer"`
	Gyrometer         map[byte]Gyrometer     `json:"gyrometer,omitempty" influxdb:"gyrometer"`
	GPSLocation       map[byte]GPSLocation   `json:"gpsLocation,omitempty" influxdb:"gps_location"`
	CounterInput      map[byte]uint16        `json:"counterInput,omitempty" influxdb:"counter_input"`
	BatteryVoltage    map[byte]float64       `json:"batteryVoltage,omitempty" influxdb:"battery_voltage"`
}

// Object returns the ExtendedCayenneLPP data object.
func (c ExtendedCayenneLPP) Object() interface{} {
	return c
}

// DecodeBytes decodes the payload from a slice of bytes.
func (c *ExtendedCayenneLPP) DecodeBytes(data []byte) error {
	var err error
	buf := make([]byte, 2)
	r := bytes.NewReader(data)

	for {
		_, err = io.ReadFull(r, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "read full error")
		}

		switch buf[1] {
		case lppDigitalInput:
			err = lppExtendedDigitalInputDecode(buf[0], r, c)
		case lppDigitalOutput:
			err = lppExtendedDigitalOutputDecode(buf[0], r, c)
		case lppAnalogInput:
			err = lppExtendedAnalogInputDecode(buf[0], r, c)
		case lppAnalogOutput:
			err = lppExtendedAnalogOutputDecode(buf[0], r, c)
		case lppIlluminanceSensor:
			err = lppExtendedIlluminanceSensorDecode(buf[0], r, c)
		case lppPresenseSensor:
			err = lppExtendedPresenseSensorDecode(buf[0], r, c)
		case lppTemperatureSensor:
			err = lppExtendedTemperatureSensorDecode(buf[0], r, c)
		case lppHumiditySensor:
			err = lppExtendedHumiditySensorDecode(buf[0], r, c)
		case lppAccelerometer:
			err = lppExtendedAccelerometerDecode(buf[0], r, c)
		case lppBarometer:
			err = lppExtendedBarometerDecode(buf[0], r, c)
		case lppGyrometer:
			err = lppExtendedGyrometerDecode(buf[0], r, c)
		case lppGPSLocation:
			err = lppExtendedGPSLocationDecode(buf[0], r, c)
		case lppExtendedCounterInput:
			err = lppExtendedCounterInputDecode(buf[0], r, c)
		case lppExtendedBatteryVoltage:
			err = lppExtendedBatteryVoltageDecode(buf[0], r, c)
		default:
			return fmt.Errorf("invalid data type: %d", buf[1])
		}

		if err != nil {
			return errors.Wrap(err, "decode error")
		}
	}

	return nil
}

// EncodeToBytes encodes the payload to a slice of bytes.
func (c ExtendedCayenneLPP) EncodeToBytes() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})

	for k, v := range c.DigitalInput {
		if err := lppDigitalInputEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.DigitalOutput {
		if err := lppDigitalOutputEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.AnalogInput {
		if err := lppAnalogInputEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.AnalogOutput {
		if err := lppAnalogOutputEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.IlluminanceSensor {
		if err := lppIlluminanceSensorEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.PresenceSensor {
		if err := lppPresenseSensorEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.TemperatureSensor {
		if err := lppTemperatureSensorEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.HumiditySensor {
		if err := lppHumiditySensorEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.Accelerometer {
		if err := lppAccelerometerEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.Barometer {
		if err := lppBarometerEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.Gyrometer {
		if err := lppGyrometerEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.GPSLocation {
		if err := lppGPSLocationEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.CounterInput {
		if err := lppExtendedCounterInputEncode(k, w, v); err != nil {
			return nil, err
		}
	}
	for k, v := range c.BatteryVoltage {
		if err := lppExtendedBatteryVoltageEncode(k, w, v); err != nil {
			return nil, err
		}
	}

	return w.Bytes(), nil
}

func lppExtendedCounterInputDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return errors.Wrap(err, "read uint16 error")
	}
	if out.CounterInput == nil {
		out.CounterInput = make(map[uint8]uint16)
	}
	out.CounterInput[channel] = count
	return nil
}

func lppExtendedCounterInputEncode(channel uint8, w io.Writer, data uint16) error {
	w.Write([]byte{channel, lppExtendedCounterInput})
	if err := binary.Write(w, binary.BigEndian, data); err != nil {
		return errors.Wrap(err, "write uint16 error")
	}
	return nil
}

func lppExtendedBatteryVoltageDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var battery int16
	if err := binary.Read(r, binary.BigEndian, &battery); err != nil {
		return errors.Wrap(err, "read int16 error")
	}
	if out.BatteryVoltage == nil {
		out.BatteryVoltage = make(map[uint8]float64)
	}
	out.BatteryVoltage[channel] = float64(battery) / 100
	return nil
}

func lppExtendedBatteryVoltageEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppExtendedBatteryVoltage})
	if err := binary.Write(w, binary.BigEndian, int16(data*100)); err != nil {
		return errors.Wrap(err, "write int16 error")
	}
	return nil
}

func lppExtendedDigitalInputDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var b uint8
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return errors.Wrap(err, "read uint8 error")
	}
	if out.DigitalInput == nil {
		out.DigitalInput = make(map[uint8]uint8)
	}
	out.DigitalInput[channel] = b
	return nil
}

func lppExtendedDigitalOutputDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var b uint8
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return errors.Wrap(err, "read uint8 error")
	}
	if out.DigitalOutput == nil {
		out.DigitalOutput = make(map[uint8]uint8)
	}
	out.DigitalOutput[channel] = b
	return nil
}

func lppExtendedAnalogInputDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var analog int16
	if err := binary.Read(r, binary.BigEndian, &analog); err != nil {
		return errors.Wrap(err, "read int16 error")
	}
	if out.AnalogInput == nil {
		out.AnalogInput = make(map[uint8]float64)
	}
	out.AnalogInput[channel] = float64(analog) / 100
	return nil
}

func lppExtendedAnalogOutputDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var analog int16
	if err := binary.Read(r, binary.BigEndian, &analog); err != nil {
		return errors.Wrap(err, "read int16 error")
	}
	if out.AnalogOutput == nil {
		out.AnalogOutput = make(map[uint8]float64)
	}
	out.AnalogOutput[channel] = float64(analog) / 100
	return nil
}

func lppExtendedIlluminanceSensorDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var illum uint16
	if err := binary.Read(r, binary.BigEndian, &illum); err != nil {
		return errors.Wrap(err, "read uint16 error")
	}
	if out.IlluminanceSensor == nil {
		out.IlluminanceSensor = make(map[uint8]uint16)
	}
	out.IlluminanceSensor[channel] = illum
	return nil
}

func lppExtendedPresenseSensorDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var b uint8
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return errors.Wrap(err, "read uint8 error")
	}
	if out.PresenceSensor == nil {
		out.PresenceSensor = make(map[uint8]uint8)
	}
	out.PresenceSensor[channel] = b
	return nil
}

func lppExtendedTemperatureSensorDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var temp int16
	if err := binary.Read(r, binary.BigEndian, &temp); err != nil {
		return errors.Wrap(err, "read int16 error")
	}
	if out.TemperatureSensor == nil {
		out.TemperatureSensor = make(map[uint8]float64)
	}
	out.TemperatureSensor[channel] = float64(temp) / 10
	return nil
}

func lppExtendedHumiditySensorDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var b uint8
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return errors.Wrap(err, "read uint8 error")
	}
	if out.HumiditySensor == nil {
		out.HumiditySensor = make(map[uint8]float64)
	}
	out.HumiditySensor[channel] = float64(b) / 2
	return nil
}

func lppExtendedAccelerometerDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	data := make([]int16, 3)
	for i := range data {
		if err := binary.Read(r, binary.BigEndian, &data[i]); err != nil {
			return errors.Wrap(err, "read int16 error")
		}
	}
	if out.Accelerometer == nil {
		out.Accelerometer = make(map[uint8]Accelerometer)
	}
	out.Accelerometer[channel] = Accelerometer{
		X: float64(data[0]) / 1000,
		Y: float64(data[1]) / 1000,
		Z: float64(data[2]) / 1000,
	}
	return nil
}

func lppExtendedBarometerDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	var bar uint16
	if err := binary.Read(r, binary.BigEndian, &bar); err != nil {
		return errors.Wrap(err, "read int16 error")
	}
	if out.Barometer == nil {
		out.Barometer = make(map[uint8]float64)
	}
	out.Barometer[channel] = float64(bar) / 10
	return nil
}

func lppExtendedGyrometerDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	data := make([]int16, 3)
	for i := range data {
		if err := binary.Read(r, binary.BigEndian, &data[i]); err != nil {
			return errors.Wrap(err, "read int16 error")
		}
	}
	if out.Gyrometer == nil {
		out.Gyrometer = make(map[uint8]Gyrometer)
	}
	out.Gyrometer[channel] = Gyrometer{
		X: float64(data[0]) / 100,
		Y: float64(data[1]) / 100,
		Z: float64(data[2]) / 100,
	}
	return nil
}

func lppExtendedGPSLocationDecode(channel uint8, r io.Reader, out *ExtendedCayenneLPP) error {
	data := make([]int32, 3)
	buf := make([]byte, 9)

	if _, err := io.ReadFull(r, buf); err != nil {
		return errors.Wrap(err, "read error")
	}

	for i := range data {
		b := make([]byte, 4)
		copy(b, buf[i*3:(i*3)+3])
		data[i] = int32(binary.BigEndian.Uint32(b)) >> 8
	}

	if out.GPSLocation == nil {
		out.GPSLocation = make(map[uint8]GPSLocation)
	}
	out.GPSLocation[channel] = GPSLocation{
		Latitude:  float64(data[0]) / 10000,
		Longitude: float64(data[1]) / 10000,
		Altitude:  float64(data[2]) / 100,
	}
	return nil
}
