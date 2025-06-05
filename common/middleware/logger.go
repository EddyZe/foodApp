package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func Logger(log *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		host := c.Request.Host

		metricsPath := path
		if len(metricsPath) > 50 {
			metricsPath = path[:50] + "..."
		}

		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(method, metricsPath))
		defer timer.ObserveDuration()

		log.Infoln("Request: ", host, method, path, query)
		c.Next()

		status := c.Writer.Status()
		log.Infoln("Response: ", status, method, path, query)

		httpRequestsTotal.WithLabelValues(method, metricsPath, strconv.Itoa(status)).Inc()
	}
}
