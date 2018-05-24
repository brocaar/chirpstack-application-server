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
	gob.Register(CayenneLPP{})
}

// CayenneLPP types.
const (
	lppDigitalInput      byte = 0
	lppDigitalOutput     byte = 1
	lppAnalogInput       byte = 2
	lppAnalogOutput      byte = 3
	lppIlluminanceSensor byte = 101
	lppPresenseSensor    byte = 102
	lppTemperatureSensor byte = 103
	lppHumiditySensor    byte = 104
	lppAccelerometer     byte = 113
	lppBarometer         byte = 115
	lppGyrometer         byte = 134
	lppGPSLocation       byte = 136
)

// Accelerometer defines the accelerometer data.
type Accelerometer struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Gyrometer defines the gyrometer data.
type Gyrometer struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// GPSLocation defines the GPS location data.
type GPSLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// CayenneLPP defines the Cayenne LPP data structure.
type CayenneLPP struct {
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
}

// Object returns the CayenneLPP data object.
func (c CayenneLPP) Object() interface{} {
	return c
}

// DecodeBytes decodes the payload from a slice of bytes.
func (c *CayenneLPP) DecodeBytes(data []byte) error {
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
			err = lppDigitalInputDecode(buf[0], r, c)
		case lppDigitalOutput:
			err = lppDigitalOutputDecode(buf[0], r, c)
		case lppAnalogInput:
			err = lppAnalogInputDecode(buf[0], r, c)
		case lppAnalogOutput:
			err = lppAnalogOutputDecode(buf[0], r, c)
		case lppIlluminanceSensor:
			err = lppIlluminanceSensorDecode(buf[0], r, c)
		case lppPresenseSensor:
			err = lppPresenseSensorDecode(buf[0], r, c)
		case lppTemperatureSensor:
			err = lppTemperatureSensorDecode(buf[0], r, c)
		case lppHumiditySensor:
			err = lppHumiditySensorDecode(buf[0], r, c)
		case lppAccelerometer:
			err = lppAccelerometerDecode(buf[0], r, c)
		case lppBarometer:
			err = lppBarometerDecode(buf[0], r, c)
		case lppGyrometer:
			err = lppGyrometerDecode(buf[0], r, c)
		case lppGPSLocation:
			err = lppGPSLocationDecode(buf[0], r, c)
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
func (c CayenneLPP) EncodeToBytes() ([]byte, error) {
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

	return w.Bytes(), nil
}

func lppDigitalInputDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppDigitalInputEncode(channel uint8, w io.Writer, data uint8) error {
	w.Write([]byte{channel, lppDigitalInput})
	if err := binary.Write(w, binary.BigEndian, data); err != nil {
		return errors.Wrap(err, "write uint8 error")
	}
	return nil
}

func lppDigitalOutputDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppDigitalOutputEncode(channel uint8, w io.Writer, data uint8) error {
	w.Write([]byte{channel, lppDigitalOutput})
	if err := binary.Write(w, binary.BigEndian, data); err != nil {
		return errors.Wrap(err, "write uint8 error")
	}
	return nil
}

func lppAnalogInputDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppAnalogInputEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppAnalogInput})
	if err := binary.Write(w, binary.BigEndian, int16(data*100)); err != nil {
		return errors.Wrap(err, "write int16 error")
	}
	return nil
}

func lppAnalogOutputDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppAnalogOutputEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppAnalogOutput})
	if err := binary.Write(w, binary.BigEndian, int16(data*100)); err != nil {
		return errors.Wrap(err, "write int16 error")
	}
	return nil
}

func lppIlluminanceSensorDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppIlluminanceSensorEncode(channel uint8, w io.Writer, data uint16) error {
	w.Write([]byte{channel, lppIlluminanceSensor})
	if err := binary.Write(w, binary.BigEndian, data); err != nil {
		return errors.Wrap(err, "write uint16 error")
	}
	return nil
}

func lppPresenseSensorDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppPresenseSensorEncode(channel uint8, w io.Writer, data uint8) error {
	w.Write([]byte{channel, lppPresenseSensor})
	if err := binary.Write(w, binary.BigEndian, data); err != nil {
		return errors.Wrap(err, "write uint8 error")
	}
	return nil
}

func lppTemperatureSensorDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppTemperatureSensorEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppTemperatureSensor})
	if err := binary.Write(w, binary.BigEndian, int16(data*10)); err != nil {
		return errors.Wrap(err, "write int16 error")
	}
	return nil
}

func lppHumiditySensorDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppHumiditySensorEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppHumiditySensor})
	if err := binary.Write(w, binary.BigEndian, uint8(data*2)); err != nil {
		return errors.Wrap(err, "write uint8 error")
	}
	return nil
}

func lppAccelerometerDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppAccelerometerEncode(channel uint8, w io.Writer, data Accelerometer) error {
	w.Write([]byte{channel, lppAccelerometer})
	vals := []int16{
		int16(data.X * 1000),
		int16(data.Y * 1000),
		int16(data.Z * 1000),
	}
	for _, v := range vals {
		if err := binary.Write(w, binary.BigEndian, v); err != nil {
			return errors.Wrap(err, "write int16 error")
		}
	}
	return nil
}

func lppBarometerDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppBarometerEncode(channel uint8, w io.Writer, data float64) error {
	w.Write([]byte{channel, lppBarometer})
	if err := binary.Write(w, binary.BigEndian, uint16(data*10)); err != nil {
		return errors.Wrap(err, "write uint16 error")
	}
	return nil
}

func lppGyrometerDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppGyrometerEncode(channel uint8, w io.Writer, data Gyrometer) error {
	w.Write([]byte{channel, lppGyrometer})
	vals := []int16{
		int16(data.X * 100),
		int16(data.Y * 100),
		int16(data.Z * 100),
	}
	for _, v := range vals {
		if err := binary.Write(w, binary.BigEndian, v); err != nil {
			return errors.Wrap(err, "write int16 error")
		}
	}
	return nil
}

func lppGPSLocationDecode(channel uint8, r io.Reader, out *CayenneLPP) error {
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

func lppGPSLocationEncode(channel uint8, w io.Writer, data GPSLocation) error {
	w.Write([]byte{channel, lppGPSLocation})
	vals := []int32{
		int32(data.Latitude * 10000),
		int32(data.Longitude * 10000),
		int32(data.Altitude * 100),
	}
	for _, v := range vals {
		b := make([]byte, 4)
		v = v << 8
		binary.BigEndian.PutUint32(b, uint32(v))
		w.Write(b[0:3])
	}
	return nil
}
