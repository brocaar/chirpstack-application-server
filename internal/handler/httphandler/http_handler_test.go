package httphandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/gusseleet/lora-app-server/internal/handler"
	"github.com/brocaar/lorawan"
)

type testHTTPHandler struct {
	requests chan *http.Request
}

func (h *testHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	h.requests <- r
	w.WriteHeader(http.StatusOK)
}

func TestHandlerConfig(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		testTable := []struct {
			Name          string
			HandlerConfig HandlerConfig
			Valid         bool
		}{
			{
				Name: "Valid headers",
				HandlerConfig: HandlerConfig{
					Headers: map[string]string{
						"Foo":     "Bar",
						"Foo-Bar": "Test",
					},
				},
				Valid: true,
			},
			{
				Name: "Invalid space in header name",
				HandlerConfig: HandlerConfig{
					Headers: map[string]string{
						"Invalid Header": "Test",
					},
				},
				Valid: false,
			},
		}

		for i, test := range testTable {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				err := test.HandlerConfig.Validate()
				if test.Valid {
					So(err, ShouldBeNil)
				} else {
					So(err, ShouldNotBeNil)
				}
			})
		}
	})
}

func TestHandler(t *testing.T) {
	Convey("Given a test HTTP server and a Handler instance", t, func() {
		httpHandler := testHTTPHandler{
			requests: make(chan *http.Request, 100),
		}
		server := httptest.NewServer(&httpHandler)
		defer server.Close()

		conf := HandlerConfig{
			Headers: map[string]string{
				"Foo": "Bar",
			},
			DataUpURL:            server.URL + "/dataup",
			JoinNotificationURL:  server.URL + "/join",
			ACKNotificationURL:   server.URL + "/ack",
			ErrorNotificationURL: server.URL + "/error",
		}
		h, err := NewHandler(conf)
		So(err, ShouldBeNil)

		Convey("Then SendDataUp sends the correct notification", func() {
			reqPL := handler.DataUpPayload{
				Data: []byte{1, 2, 3, 4},
			}
			So(h.SendDataUp(reqPL), ShouldBeNil)

			req := <-httpHandler.requests
			So(req.URL.Path, ShouldEqual, "/dataup")

			var pl handler.DataUpPayload
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
			So(req.Header.Get("Foo"), ShouldEqual, "Bar")
			So(req.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})

		Convey("Then SendJoinNotification sends the correct notification", func() {
			reqPL := handler.JoinNotification{
				DevAddr: lorawan.DevAddr{1, 2, 3, 4},
			}
			So(h.SendJoinNotification(reqPL), ShouldBeNil)

			req := <-httpHandler.requests
			So(req.URL.Path, ShouldEqual, "/join")

			var pl handler.JoinNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
			So(req.Header.Get("Foo"), ShouldEqual, "Bar")
			So(req.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})

		Convey("Then SendACKNotification sends the correct notification", func() {
			reqPL := handler.ACKNotification{
				Reference: "ack-123",
			}
			So(h.SendACKNotification(reqPL), ShouldBeNil)

			req := <-httpHandler.requests
			So(req.URL.Path, ShouldEqual, "/ack")

			var pl handler.ACKNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
			So(req.Header.Get("Foo"), ShouldEqual, "Bar")
			So(req.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})

		Convey("Then SendErrorNotification sends the correct notification", func() {
			reqPL := handler.ErrorNotification{
				Error: "boom!",
			}
			So(h.SendErrorNotification(reqPL), ShouldBeNil)

			req := <-httpHandler.requests
			So(req.URL.Path, ShouldEqual, "/error")

			var pl handler.ErrorNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
			So(req.Header.Get("Foo"), ShouldEqual, "Bar")
			So(req.Header.Get("Content-Type"), ShouldEqual, "application/json")
		})
	})
}
