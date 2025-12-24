package monitor

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
)

// PerformanceMetrics tracks performance statistics
type PerformanceMetrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests   int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration

	// Latency buckets for percentile calculation
	latencies []time.Duration

	// Database metrics
	DBQueryCount    int64
	DBTotalDuration time.Duration

	// Streaming metrics
	StreamingRequests int64
	TTFTTotal         time.Duration

	// Last reset time
	LastReset time.Time
}

var (
	globalMetrics     *PerformanceMetrics
	metricsEnabled    = env.Bool("METRICS_ENABLED", false)
	metricsResetInterval = env.Int("METRICS_RESET_INTERVAL", 3600) // seconds
	metricsOnce       sync.Once
)

// InitMetrics initializes the metrics system
func InitMetrics() {
	if !metricsEnabled {
		return
	}

	metricsOnce.Do(func() {
		globalMetrics = &PerformanceMetrics{
			latencies: make([]time.Duration, 0, 10000),
			LastReset: time.Now(),
		}

		logger.SysLog("performance metrics enabled")
		logger.SysLog(fmt.Sprintf("metrics will reset every %d seconds", metricsResetInterval))

		// Start periodic reset
		go periodicReset()
	})
}

// RecordRequest records a request metric
func RecordRequest(ctx context.Context, duration time.Duration, success bool) {
	if !metricsEnabled || globalMetrics == nil {
		return
	}

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.TotalRequests++
	if !success {
		globalMetrics.FailedRequests++
	}

	globalMetrics.TotalDuration += duration

	if globalMetrics.MinDuration == 0 || duration < globalMetrics.MinDuration {
		globalMetrics.MinDuration = duration
	}
	if duration > globalMetrics.MaxDuration {
		globalMetrics.MaxDuration = duration
	}

	// Store latency for percentile calculation
	if len(globalMetrics.latencies) < 10000 {
		globalMetrics.latencies = append(globalMetrics.latencies, duration)
	}
}

// RecordDBQuery records a database query metric
func RecordDBQuery(ctx context.Context, duration time.Duration) {
	if !metricsEnabled || globalMetrics == nil {
		return
	}

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.DBQueryCount++
	globalMetrics.DBTotalDuration += duration
}

// RecordTTFT records Time To First Token for streaming
func RecordTTFT(ctx context.Context, duration time.Duration) {
	if !metricsEnabled || globalMetrics == nil {
		return
	}

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.StreamingRequests++
	globalMetrics.TTFTTotal += duration
}

// GetMetrics returns current metrics
func GetMetrics() *MetricsSnapshot {
	if !metricsEnabled || globalMetrics == nil {
		return nil
	}

	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	// Avoid division by zero
	totalRequests := globalMetrics.TotalRequests
	if totalRequests == 0 {
		totalRequests = 1
	}

	snapshot := &MetricsSnapshot{
		TotalRequests:       globalMetrics.TotalRequests,
		FailedRequests:      globalMetrics.FailedRequests,
		SuccessRate:         float64(globalMetrics.TotalRequests-globalMetrics.FailedRequests) / float64(totalRequests) * 100,
		TotalDuration:       globalMetrics.TotalDuration,
		AverageDuration:     time.Duration(int64(globalMetrics.TotalDuration) / max(globalMetrics.TotalRequests, 1)),
		MinDuration:         globalMetrics.MinDuration,
		MaxDuration:         globalMetrics.MaxDuration,
		DBQueryCount:        globalMetrics.DBQueryCount,
		DBTotalDuration:     globalMetrics.DBTotalDuration,
		DBAvgDuration:       time.Duration(int64(globalMetrics.DBTotalDuration) / max(globalMetrics.DBQueryCount, 1)),
		StreamingRequests:   globalMetrics.StreamingRequests,
		AvgTTFT:             time.Duration(int64(globalMetrics.TTFTTotal) / max(globalMetrics.StreamingRequests, 1)),
		LastReset:           globalMetrics.LastReset,
		CurrentTime:         time.Now(),
	}

	// Calculate percentiles
	if len(globalMetrics.latencies) > 0 {
		snapshot.P50, snapshot.P95, snapshot.P99 = calculatePercentiles(globalMetrics.latencies)
	}

	return snapshot
}

// MetricsSnapshot represents a snapshot of metrics
type MetricsSnapshot struct {
	TotalRequests     int64         `json:"total_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	SuccessRate       float64       `json:"success_rate"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	MinDuration       time.Duration `json:"min_duration"`
	MaxDuration       time.Duration `json:"max_duration"`
	P50               time.Duration `json:"p50_latency"`
	P95               time.Duration `json:"p95_latency"`
	P99               time.Duration `json:"p99_latency"`
	DBQueryCount      int64         `json:"db_query_count"`
	DBTotalDuration   time.Duration `json:"db_total_duration"`
	DBAvgDuration     time.Duration `json:"db_avg_duration"`
	StreamingRequests int64         `json:"streaming_requests"`
	AvgTTFT           time.Duration `json:"avg_ttft"`
	LastReset         time.Time     `json:"last_reset"`
	CurrentTime       time.Time     `json:"current_time"`
}

// calculatePercentiles calculates p50, p95, p99 from latency samples
func calculatePercentiles(latencies []time.Duration) (p50, p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}

	// Use Go's built-in sort for efficiency
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	p50 = sorted[len(sorted)*50/100]
	p95 = sorted[len(sorted)*95/100]
	p99 = sorted[len(sorted)*99/100]

	return p50, p95, p99
}

// periodicReset resets metrics periodically
func periodicReset() {
	ticker := time.NewTicker(time.Duration(metricsResetInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ResetMetrics()
	}
}

// ResetMetrics resets all metrics
func ResetMetrics() {
	if !metricsEnabled || globalMetrics == nil {
		return
	}

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	logger.SysLog("resetting performance metrics")

	globalMetrics.TotalRequests = 0
	globalMetrics.FailedRequests = 0
	globalMetrics.TotalDuration = 0
	globalMetrics.MinDuration = 0
	globalMetrics.MaxDuration = 0
	globalMetrics.latencies = make([]time.Duration, 0, 10000)
	globalMetrics.DBQueryCount = 0
	globalMetrics.DBTotalDuration = 0
	globalMetrics.StreamingRequests = 0
	globalMetrics.TTFTTotal = 0
	globalMetrics.LastReset = time.Now()
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
