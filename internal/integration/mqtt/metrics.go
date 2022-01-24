package mqtt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integration_mqtt_event_count",
		Help: "The number of published events by the MQTT integration (per event type).",
	}, []string{"event"})

	cc = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "integration_mqtt_command_count",
		Help: "The number of received commands by the MQTT integration (per command).",
	}, []string{"command"})
)

func mqttEventCounter(e string) prometheus.Counter {
	return ec.With(prometheus.Labels{"event": e})
}

func mqttCommandCounter(c string) prometheus.Counter {
	return cc.With(prometheus.Labels{"command": c})
}
