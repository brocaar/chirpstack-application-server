package js

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

var (
	maxExecutionTime = 10 * time.Millisecond
)

// Setup configures the JS codec.
func Setup(conf config.Config) error {
	maxExecutionTime = conf.ApplicationServer.Codec.JS.MaxExecutionTime
	return nil
}

// BinaryToJSON encodes the given binary payload to JSON.
func BinaryToJSON(fPort uint8, variables map[string]string, decodeScript string, b []byte) ([]byte, error) {
	decodeScript = decodeScript + "\n\nDecode(fPort, bytes, variables);\n"

	vars := make(map[string]interface{})

	vars["fPort"] = fPort
	vars["bytes"] = b
	vars["variables"] = variables

	v, err := executeJS(decodeScript, vars)
	if err != nil {
		return nil, errors.Wrap(err, "execute js error")
	}

	return json.Marshal(v)
}

// JSONToBinary encodes the given JSON payload to binary.
func JSONToBinary(fPort uint8, variables map[string]string, encodeScript string, b []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}

	encodeScript = encodeScript + "\n\nEncode(fPort, obj, variables);"

	vars := make(map[string]interface{})

	vars["fPort"] = fPort
	vars["obj"] = v
	vars["variables"] = variables

	v, err := executeJS(encodeScript, vars)
	if err != nil {
		return nil, errors.Wrap(err, "execute js error")
	}

	return interfaceToByteSlice(v)
}

func executeJS(script string, vars map[string]interface{}) (out interface{}, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%s", caught)
		}
	}()

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.SetStackDepthLimit(32)

	for k, v := range vars {
		if err := vm.Set(k, v); err != nil {
			return nil, errors.Wrap(err, "set variable error")
		}
	}

	go func() {
		time.Sleep(maxExecutionTime)
		vm.Interrupt <- func() {
			panic(errors.New("execution timeout"))
		}
	}()

	var val otto.Value
	val, err = vm.Run(script)
	if err != nil {
		fmt.Println(err)
		return nil, errors.Wrap(err, "js vm error")
	}

	return val.Export()
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
