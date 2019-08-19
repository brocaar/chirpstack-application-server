package js

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan/backend"
)

var (
	reqCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_joinserver_request_count",
		Help: "The number of join-server API requests (per message-type and status code)",
	}, []string{"message_type", "status_code"})

	reqTimer = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "api_joinserver_request_duration_seconds",
		Help: "The duration of serving join-server API requests (per message-type and status code)",
	}, []string{"message_type", "status_code"})
)

type prometheusMiddleware struct {
	handler         http.Handler
	timingHistogram bool
}

func (h *prometheusMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var buf bytes.Buffer
	if r.Body != nil {
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.WithError(err).Error("api/js: read request body error")
		}
		r.Body = ioutil.NopCloser(&buf)
	}

	var basePL backend.BasePayload
	if err := json.Unmarshal(buf.Bytes(), &basePL); err != nil {
		log.WithError(err).Error("api/js: unmarshal base payload error")
	}

	sw := statusWriter{ResponseWriter: w}
	h.handler.ServeHTTP(&sw, r)

	labels := prometheus.Labels{"message_type": string(basePL.MessageType), "status_code": strconv.FormatInt(int64(sw.status), 10)}
	reqCount.With(labels).Inc()

	if h.timingHistogram {
		reqTimer.With(labels).Observe(float64(time.Since(start)) / float64(time.Second))
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.ResponseWriter.Write(b)
}
