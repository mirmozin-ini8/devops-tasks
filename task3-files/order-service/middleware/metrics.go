package middleware

import (
	"order-service/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		if path == "/orders/metrics" {
			c.Next()
			return
		}

		metrics.HttpRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		metrics.HttpRequestsInFlight.Dec()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.HttpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
