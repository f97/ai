package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
)

// Batch logging system for reducing DB write overhead
// This is especially important for streaming requests that can generate many log entries

var (
	logBatchEnabled     = env.Bool("LOG_BATCH_ENABLED", false)
	logBatchSize        = env.Int("LOG_BATCH_SIZE", 50)
	logBatchFlushInterval = env.Int("LOG_BATCH_FLUSH_INTERVAL", 5) // seconds
	
	logBatchChannel chan *Log
	logBatchBuffer  []*Log
	logBatchMutex   sync.Mutex
	logBatchOnce    sync.Once
	logBatchWg      sync.WaitGroup  // Add WaitGroup for proper synchronization
)

// InitLogBatchProcessor initializes the async batch log processor
func InitLogBatchProcessor() {
	if !logBatchEnabled {
		return
	}
	
	logBatchOnce.Do(func() {
		logBatchChannel = make(chan *Log, logBatchSize*2)
		logBatchBuffer = make([]*Log, 0, logBatchSize)
		
		logger.SysLog("async batch log processor enabled")
		logger.SysLog("⚠️  TRADE-OFF: Logs are batched, may lose recent logs on crash")
		logger.SysLog(fmt.Sprintf("batch size: %d, flush interval: %ds", logBatchSize, logBatchFlushInterval))
		
		// Start batch processor goroutine
		logBatchWg.Add(2)
		go processBatchLogs()
		
		// Start periodic flush goroutine
		go periodicFlushLogs()
	})
}

// processBatchLogs processes log entries from the channel
func processBatchLogs() {
	defer logBatchWg.Done()
	
	for log := range logBatchChannel {
		logBatchMutex.Lock()
		logBatchBuffer = append(logBatchBuffer, log)
		shouldFlush := len(logBatchBuffer) >= logBatchSize
		logBatchMutex.Unlock()
		
		if shouldFlush {
			flushLogBatch()
		}
	}
	
	// Flush remaining logs when channel is closed
	flushLogBatch()
}

// periodicFlushLogs flushes logs periodically
func periodicFlushLogs() {
	defer logBatchWg.Done()
	
	ticker := time.NewTicker(time.Duration(logBatchFlushInterval) * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		flushLogBatch()
	}
}

// flushLogBatch flushes accumulated logs to database
func flushLogBatch() {
	logBatchMutex.Lock()
	if len(logBatchBuffer) == 0 {
		logBatchMutex.Unlock()
		return
	}
	
	// Swap buffer
	toFlush := logBatchBuffer
	logBatchBuffer = make([]*Log, 0, logBatchSize)
	logBatchMutex.Unlock()
	
	// Batch insert
	if len(toFlush) > 0 {
		if err := LOG_DB.CreateInBatches(toFlush, logBatchSize).Error; err != nil {
			logger.SysError(fmt.Sprintf("failed to batch insert logs: %s", err.Error()))
		} else {
			logger.SysLog(fmt.Sprintf("flushed %d logs to database", len(toFlush)))
		}
	}
}

// RecordLogAsync records a log asynchronously if batch mode is enabled
func RecordLogAsync(ctx context.Context, log *Log) {
	if !logBatchEnabled {
		// Fall back to synchronous logging
		recordLogHelper(ctx, log)
		return
	}
	
	select {
	case logBatchChannel <- log:
		// Successfully queued
	default:
		// Channel full, log synchronously as fallback
		logger.Warn(ctx, "log batch channel full, falling back to sync logging")
		recordLogHelper(ctx, log)
	}
}

// FlushPendingLogs flushes any pending logs (should be called on shutdown)
func FlushPendingLogs() {
	if !logBatchEnabled {
		return
	}
	
	logger.SysLog("flushing pending batch logs...")
	
	// Close the channel to signal goroutines to finish
	close(logBatchChannel)
	
	// Wait for all goroutines to finish
	logBatchWg.Wait()
	
	logger.SysLog("all batch logs flushed successfully")
}
