package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
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

func TestHTTPHandler(t *testing.T) {
	Convey("Given a test HTTP server and a HTTPHandler instance", t, func() {
		h := testHTTPHandler{
			requests: make(chan *http.Request, 100),
		}
		server := httptest.NewServer(&h)
		defer server.Close()

		conf := HTTPHandlerConfig{
			Headers: map[string]string{
				"Foo": "Bar",
			},
			DataUpURL:            server.URL + "/dataup",
			JoinNotificationURL:  server.URL + "/join",
			ACKNotificationURL:   server.URL + "/ack",
			ErrorNotificationURL: server.URL + "/error",
		}
		handler, err := NewHTTPHandler(conf)
		So(err, ShouldBeNil)

		Convey("Then SendDataUp sends the correct notification", func() {
			reqPL := DataUpPayload{
				Data: []byte{1, 2, 3, 4},
			}
			So(handler.SendDataUp(reqPL), ShouldBeNil)

			req := <-h.requests
			So(req.URL.Path, ShouldEqual, "/dataup")

			var pl DataUpPayload
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
		})

		Convey("Then SendJoinNotification sends the correct notification", func() {
			reqPL := JoinNotification{
				DevAddr: lorawan.DevAddr{1, 2, 3, 4},
			}
			So(handler.SendJoinNotification(reqPL), ShouldBeNil)

			req := <-h.requests
			So(req.URL.Path, ShouldEqual, "/join")

			var pl JoinNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
		})

		Convey("Then SendACKNotification sends the correct notification", func() {
			reqPL := ACKNotification{
				Reference: "ack-123",
			}
			So(handler.SendACKNotification(reqPL), ShouldBeNil)

			req := <-h.requests
			So(req.URL.Path, ShouldEqual, "/ack")

			var pl ACKNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
		})

		Convey("Then SendErrorNotification sends the correct notification", func() {
			reqPL := ErrorNotification{
				Error: "boom!",
			}
			So(handler.SendErrorNotification(reqPL), ShouldBeNil)

			req := <-h.requests
			So(req.URL.Path, ShouldEqual, "/error")

			var pl ErrorNotification
			So(json.NewDecoder(req.Body).Decode(&pl), ShouldBeNil)
			So(pl, ShouldResemble, reqPL)
		})
	})
}
