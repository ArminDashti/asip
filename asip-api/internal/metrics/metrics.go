package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asip_http_requests_total",
			Help: "Total HTTP requests handled by asip-api.",
		},
		[]string{"method", "path", "code"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asip_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

// Register mounts /metrics and request instrumentation on the engine.
func Register(engine *gin.Engine) {
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.Use(Middleware())
}

// Middleware records request counts and latency. Skips /metrics.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		method := c.Request.Method
		code := strconv.Itoa(c.Writer.Status())

		httpRequests.WithLabelValues(method, path, code).Inc()
		httpDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	}
}
