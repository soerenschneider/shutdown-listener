package internal

import (
	"bytes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/shutdown-listener/internal/config"
	"io/ioutil"
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
		Help:      "Total amount of reconnecting to the MQTT broker",
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

func WriteMetrics(path string) error {
	log.Info().Msgf("Dumping metrics to %s", path)
	metrics, err := dumpMetrics()
	if err != nil {
		log.Info().Msgf("Error dumping metrics: %v", err)
		return err
	}

	err = ioutil.WriteFile(path, []byte(metrics), 0644)
	if err != nil {
		log.Info().Msgf("Error writing metrics to '%s': %v", path, err)
	}
	return err
}

func dumpMetrics() (string, error) {
	var buf = &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return "", err
	}

	for _, f := range families {
		if err := enc.Encode(f); err != nil {
			log.Info().Msgf("could not encode metric: %s", err.Error())
		}
	}

	return buf.String(), nil
}
