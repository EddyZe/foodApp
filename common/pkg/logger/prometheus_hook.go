package logger

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	logMessagesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_log_messages_total",
			Help: "Total number of log messages",
		},
		[]string{"level", "service"},
	)
)

func init() {
	prometheus.MustRegister(logMessagesCounter)
}

type PrometheusHook struct {
	service string
}

func NewPrometheusHook(service string) *PrometheusHook {
	return &PrometheusHook{service: service}
}

func (h *PrometheusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *PrometheusHook) Fire(entry *logrus.Entry) error {
	logMessagesCounter.WithLabelValues(entry.Level.String(), h.service).Inc()
	return nil
}
