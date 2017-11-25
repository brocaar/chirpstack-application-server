package codec

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"

	"github.com/brocaar/lora-app-server/internal/common"
)

// CustomJS is a scriptable JS codec.
type CustomJS struct {
	fPort        uint8
	encodeScript string
	decodeScript string
	data         interface{}
}

// NewCustomJS creates a new custom JS codec.
func NewCustomJS(fPort uint8, encodeScript, decodeScript string) *CustomJS {
	return &CustomJS{
		fPort:        fPort,
		encodeScript: encodeScript,
		decodeScript: decodeScript,
	}
}

// MarshalJSON implements json.Marshaler.
func (c CustomJS) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.data)
}

// UnmarshalJSON implement json.Unmarshaler.
func (c *CustomJS) UnmarshalJSON(text []byte) error {
	return json.Unmarshal(text, &c.data)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (c *CustomJS) UnmarshalBinary(data []byte) (err error) {
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
		time.Sleep(common.CodecMaxExecTime)
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

	c.data, err = val.Export()
	if err != nil {
		return errors.Wrap(err, "export error")
	}

	return nil
}

// MarshalBinary implements encoding.BinaryMashaler.
func (c CustomJS) MarshalBinary() (b []byte, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%s", caught)
		}
	}()

	script := c.encodeScript + "\n\nEncode(fPort, obj);\n"

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.SetStackDepthLimit(32)
	vm.Set("obj", c.data)
	vm.Set("fPort", c.fPort)

	go func() {
		time.Sleep(common.CodecMaxExecTime)
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
		return nil, errors.New("function must return a slice")
	}
	var out interface{}
	out, err = val.Export()
	if err != nil {
		return nil, errors.Wrap(err, "export error")
	}

	items, ok := out.([]interface{})
	if !ok {
		return nil, errors.New("function must return a slice")
	}

	for _, item := range items {
		switch v := item.(type) {
		case uint8:
			b = append(b, byte(v))
		case int:
			b = append(b, byte(v))
		case float64:
			b = append(b, byte(v))
		default:
			return nil, fmt.Errorf("invalid slice value, %T", v)
		}
	}

	return b, nil
}
