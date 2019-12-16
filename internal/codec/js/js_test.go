package js

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestJSDecode(t *testing.T) {
	tests := []struct {
		Name           string
		Script         string
		Payload        []byte
		FPort          uint8
		Variables      map[string]string
		ExpectedObject map[string]interface{}
		ExpectedJSON   string
		ExpectedError  error
	}{
		{
			Name: "valid function",
			Script: `
					function Decode(port, bytes) {
						return {
							"port": port,
							"on": bytes[0] == 1
						};
					}
				`,
			Payload: []byte{1},
			FPort:   3,
			ExpectedObject: map[string]interface{}{
				"port": 3,
				"on":   true,
			},
			ExpectedJSON: `{"on":true,"port":3}`,
		},
		{
			Name:          "function error",
			Script:        ``,
			Payload:       []byte{1},
			FPort:         3,
			ExpectedError: errors.New("execute js error: js vm error: ReferenceError: 'Decode' is not defined"),
		},
		{
			Name: "function timeout",
			Script: `
					function Decode(fPort, bytes) {
						while(true) {}
					}
				`,
			Payload:       []byte{1},
			FPort:         3,
			ExpectedError: errors.New("execute js error: execution timeout"),
		},
		{
			Name: "variables",
			Script: `
					function Decode(port, bytes, variables) {
						return {
							"port": port,
							"on": bytes[0] == 1,
							"calibration": parseFloat(variables["calibration"])
						};
					}
				`,
			Payload: []byte{1},
			FPort:   3,
			Variables: map[string]string{
				"calibration": "1.123",
			},
			ExpectedObject: map[string]interface{}{
				"port": 3,
				"on":   true,
			},
			ExpectedJSON: `{"calibration":1.123,"on":true,"port":3}`,
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			jsonB, err := BinaryToJSON(tst.FPort, tst.Variables, tst.Script, tst.Payload)
			if tst.ExpectedError != nil {
				assert.Equal(tst.ExpectedError.Error(), err.Error())
				return
			}
			assert.NoError(err)

			assert.Equal(tst.ExpectedJSON, string(jsonB))
		})
	}
}

func TestJSEncode(t *testing.T) {
	tests := []struct {
		Name          string
		Script        string
		JSON          string
		FPort         uint8
		Variables     map[string]string
		ExpectedBytes []byte
		ExpectedError error
	}{
		{
			Name: "valid function",
			Script: `
					function Encode(fPort, obj) {
						var bytes = [];
						bytes[0] = obj.Temp;
						bytes[1] = fPort;
						return bytes;
					}
				`,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedBytes: []byte{20, 10},
		},
		{
			Name: "return []int64",
			Script: `
					function Encode(fPort, obj) {
						return [1,2,3];
					}
				`,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedBytes: []byte{1, 2, 3},
		},
		{
			Name: "return float array",
			Script: `
					function Encode(fPort, obj) {
						return [1.123, 2.234];
					}
				`,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedError: errors.New("array value must be in byte range (0 - 255), got: 1.123000"),
		},
		{
			Name: "return invalid bytes",
			Script: `
					function Encode(fPort, obj) {
						return [256, 123];
					}
				`,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedError: errors.New("array value must be in byte range (0 - 255), got: 256"),
		},
		{
			Name:          "invalid function",
			Script:        ``,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedError: errors.New("execute js error: js vm error: ReferenceError: 'Encode' is not defined"),
		},
		{
			Name: "function timeout",
			Script: `
					function Encode(fPort, obj) {
						while(true) {}
					}
				`,
			FPort:         10,
			JSON:          `{"Temp": 20}`,
			ExpectedError: errors.New("execute js error: execution timeout"),
		},
		{
			Name: "variables",
			Script: `
					function Encode(fPort, obj, variables) {
						var bytes = [];
						bytes[0] = obj.Temp - parseFloat(variables["calibration"]);
						bytes[1] = fPort;
						return bytes;
					}
				`,
			FPort: 10,
			Variables: map[string]string{
				"calibration": "5",
			},
			JSON:          `{"Temp": 20}`,
			ExpectedBytes: []byte{15, 10},
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			b, err := JSONToBinary(tst.FPort, tst.Variables, tst.Script, []byte(tst.JSON))
			if tst.ExpectedError != nil {
				assert.Equal(tst.ExpectedError.Error(), err.Error())
				return
			}
			assert.NoError(err)
			assert.Equal(tst.ExpectedBytes, b)
		})
	}
}
