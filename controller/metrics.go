package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/monitor"
)

// GetMetrics returns current performance metrics
func GetMetrics(c *gin.Context) {
	metrics := monitor.GetMetrics()
	if metrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "metrics not enabled",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// ResetMetrics resets all metrics
func ResetMetrics(c *gin.Context) {
	monitor.ResetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "metrics reset successfully",
	})
}
