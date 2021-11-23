package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/shutdown-listener/internal/config"
	"net/http"
	"strings"
)

var (
	namespace = strings.ReplaceAll(config.AppName, "-", "_")
)

var (
	MetricMessageVerifyErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "server",
		Name:      "message_verify_errors_total",
		Help:      "Total amount of messages that could not be verified",
	})

	MetricMqttReconnections = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "mqtt",
		Name:      "reconnections_triggered_total",
		Help:      "Total amount of reconnecting to the MQTT broker",
	})

	MetricHttpRequestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "http",
		Name:      "message_request_errors_total",
		Help:      "Errors while receiving messages via HTTP",
	})

	MetricHeartbeat = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "heartbeat_seconds",
		Help:      "Continuous heartbeat",
	})

	MetricVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "version",
		Help:      "Version information",
	}, []string{"version", "commit"})
)

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Msgf("Can not start metrics server at %s: %v", addr, err)
	}
}