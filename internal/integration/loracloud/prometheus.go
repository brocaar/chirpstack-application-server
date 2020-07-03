package loracloud

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ad = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "integration_loracloud_api_duration_seconds",
		Help: "The duration of LoRa Cloud API calls (per endpoint).",
	}, []string{"endpoint"})
)

func loRaCloudAPIDuration(e string) prometheus.Observer {
	return ad.With(prometheus.Labels{"endpoint": e})
}
