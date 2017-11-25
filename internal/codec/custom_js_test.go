package codec

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCustomJSDecode(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		tests := []struct {
			Name           string
			Script         string
			Payload        []byte
			FPort          uint8
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
				ExpectedError: errors.New("js vm error: ReferenceError: 'Decode' is not defined"),
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
				ExpectedError: errors.New("execution timeout"),
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				js := NewCustomJS(test.FPort, "", test.Script)
				err := js.UnmarshalBinary(test.Payload)
				if test.ExpectedError != nil {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, test.ExpectedError.Error())
					return
				}

				So(err, ShouldEqual, nil)
				for k, v := range test.ExpectedObject {
					So(js.data.(map[string]interface{})[k], ShouldEqual, v)
				}

				b, err := js.MarshalJSON()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, test.ExpectedJSON)
			})
		}
	})
}

func TestCustomEncodeJS(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		tests := []struct {
			Name          string
			Script        string
			JSON          string
			FPort         uint8
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
				Name:          "invalid function",
				Script:        ``,
				FPort:         10,
				JSON:          `{"Temp": 20}`,
				ExpectedError: errors.New("js vm error: ReferenceError: 'Encode' is not defined"),
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
				ExpectedError: errors.New("execution timeout"),
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				js := NewCustomJS(test.FPort, test.Script, "")
				So(js.UnmarshalJSON([]byte(test.JSON)), ShouldBeNil)

				b, err := js.MarshalBinary()
				if test.ExpectedError != nil {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, test.ExpectedError.Error())
					return
				}

				So(err, ShouldEqual, nil)
				So(b, ShouldResemble, test.ExpectedBytes)
			})
		}
	})
}
