package codec

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

func init() {
	gob.Register(CustomJS{})
}

// CodecMaxExecTime holds the max. time the (custom) codec is allowed to
// run.
var CodecMaxExecTime = 10 * time.Millisecond

// CustomJS is a scriptable JS codec.
type CustomJS struct {
	fPort        uint8
	encodeScript string
	decodeScript string
	Data         interface{}
}

// NewCustomJS creates a new custom JS codec.
func NewCustomJS(fPort uint8, encodeScript, decodeScript string) *CustomJS {
	return &CustomJS{
		fPort:        fPort,
		encodeScript: encodeScript,
		decodeScript: decodeScript,
	}
}

// Object returns the object data.
func (c CustomJS) Object() interface{} {
	return c.Data
}

// MarshalJSON implements json.Marshaler.
func (c CustomJS) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Data)
}

// UnmarshalJSON implement json.Unmarshaler.
func (c *CustomJS) UnmarshalJSON(text []byte) error {
	return json.Unmarshal(text, &c.Data)
}

// DecodeBytes decodes the payload from a slice of bytes.
func (c *CustomJS) DecodeBytes(data []byte) (err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%s", caught)
		}
	}()

	script := c.decodeScript + "\n\nDecode(fPort, bytes);\n"

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.SetStackDepthLimit(32)
	vm.Set("bytes", data)
	vm.Set("fPort", c.fPort)

	go func() {
		time.Sleep(CodecMaxExecTime)
		vm.Interrupt <- func() {
			panic(errors.New("execution timeout"))
		}
	}()

	var val otto.Value
	val, err = vm.Run(script)
	if err != nil {
		return errors.Wrap(err, "js vm error")
	}

	if !val.IsObject() {
		return errors.New("function must return object")
	}

	c.Data, err = val.Export()
	if err != nil {
		return errors.Wrap(err, "export error")
	}

	return nil
}

// EncodeToBytes encodes the payload to a slice of bytes.
func (c CustomJS) EncodeToBytes() (b []byte, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%s", caught)
		}
	}()

	script := c.encodeScript + "\n\nEncode(fPort, obj);\n"

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.SetStackDepthLimit(32)
	vm.Set("obj", c.Data)
	vm.Set("fPort", c.fPort)

	go func() {
		time.Sleep(CodecMaxExecTime)
		vm.Interrupt <- func() {
			panic(errors.New("execution timeout"))
		}
	}()

	var val otto.Value
	val, err = vm.Run(script)
	if err != nil {
		return nil, errors.Wrap(err, "js vm error")
	}
	if !val.IsObject() {
		return nil, errors.New("function must return an array")
	}

	var out interface{}
	out, err = val.Export()
	if err != nil {
		return nil, errors.Wrap(err, "export error")
	}

	return interfaceToByteSlice(out)
}

func interfaceToByteSlice(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nil, errors.New("value must not be nil")
	}

	if reflect.TypeOf(obj).Kind() != reflect.Slice {
		return nil, errors.New("value must be an array")
	}

	s := reflect.ValueOf(obj)
	l := s.Len()

	var out []byte
	for i := 0; i < l; i++ {
		var b int64

		el := s.Index(i).Interface()
		switch v := el.(type) {
		case int:
			b = int64(v)
		case uint:
			b = int64(v)
		case uint8:
			b = int64(v)
		case int8:
			b = int64(v)
		case uint16:
			b = int64(v)
		case int16:
			b = int64(v)
		case uint32:
			b = int64(v)
		case int32:
			b = int64(v)
		case uint64:
			b = int64(v)
			if uint64(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %d", v)
			}
		case int64:
			b = int64(v)
		case float32:
			b = int64(v)
			if float32(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %f", v)
			}
		case float64:
			b = int64(v)
			if float64(b) != v {
				return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %f", v)
			}
		default:
			return nil, fmt.Errorf("array value must be an array of ints or floats, got: %T", el)
		}

		if b < 0 || b > 255 {
			return nil, fmt.Errorf("array value must be in byte range (0 - 255), got: %d", b)
		}

		out = append(out, byte(b))
	}

	return out, nil
}
