package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/monitor"
)

// PerformanceMonitor is a middleware that tracks request performance
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		success := c.Writer.Status() < 400

		// Record metrics
		monitor.RecordRequest(c.Request.Context(), duration, success)

		// Log slow requests
		if duration > 2*time.Second {
			logger.Warnf(c.Request.Context(), 
				"slow request: %s %s took %v (status: %d)",
				c.Request.Method,
				c.Request.URL.Path,
				duration,
				c.Writer.Status())
		}

		// Set response time header
		c.Header("X-Response-Time-Ms", fmt.Sprintf("%.2f", float64(duration.Microseconds())/1000.0))
	}
}
