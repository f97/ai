# Code Snippets for One-API Performance Optimization
# 性能优化代码片段集合

这个文档包含了可以直接使用的代码片段，用于各种性能优化场景。

---

## 1. Async Batch SQLite Writer (异步批量写入器)

### 基础版本

```go
package model

import (
	"context"
	"fmt"
	"sync"
	"time"
	"gorm.io/gorm"
)

// AsyncBatchWriter handles batched async writes to database
type AsyncBatchWriter struct {
	db            *gorm.DB
	batchSize     int
	flushInterval time.Duration
	channel       chan interface{}
	buffer        []interface{}
	mutex         sync.Mutex
	wg            sync.WaitGroup
}

// NewAsyncBatchWriter creates a new async batch writer
func NewAsyncBatchWriter(db *gorm.DB, batchSize int, flushInterval time.Duration) *AsyncBatchWriter {
	writer := &AsyncBatchWriter{
		db:            db,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		channel:       make(chan interface{}, batchSize*2),
		buffer:        make([]interface{}, 0, batchSize),
	}
	
	// Start processors
	writer.wg.Add(2)
	go writer.processRecords()
	go writer.periodicFlush()
	
	return writer
}

// Write queues a record for async writing
func (w *AsyncBatchWriter) Write(record interface{}) {
	select {
	case w.channel <- record:
		// Successfully queued
	default:
		// Channel full, write synchronously as fallback
		w.db.Create(record)
	}
}

// processRecords processes records from channel
func (w *AsyncBatchWriter) processRecords() {
	defer w.wg.Done()
	
	for record := range w.channel {
		w.mutex.Lock()
		w.buffer = append(w.buffer, record)
		shouldFlush := len(w.buffer) >= w.batchSize
		w.mutex.Unlock()
		
		if shouldFlush {
			w.flush()
		}
	}
}

// periodicFlush flushes buffer periodically
func (w *AsyncBatchWriter) periodicFlush() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		w.flush()
	}
}

// flush writes buffer to database
func (w *AsyncBatchWriter) flush() {
	w.mutex.Lock()
	if len(w.buffer) == 0 {
		w.mutex.Unlock()
		return
	}
	
	toFlush := w.buffer
	w.buffer = make([]interface{}, 0, w.batchSize)
	w.mutex.Unlock()
	
	if len(toFlush) > 0 {
		if err := w.db.CreateInBatches(toFlush, w.batchSize).Error; err != nil {
			fmt.Printf("batch write error: %v\n", err)
		}
	}
}

// Close closes the writer and flushes remaining records
func (w *AsyncBatchWriter) Close() {
	close(w.channel)
	w.wg.Wait()
	w.flush()
}
```

### 使用示例

```go
// 初始化
batchWriter := NewAsyncBatchWriter(db, 50, 5*time.Second)
defer batchWriter.Close()

// 写入日志
log := &Log{
	UserId:    userId,
	Type:      LogTypeConsume,
	Content:   "API call",
	CreatedAt: time.Now().Unix(),
}
batchWriter.Write(log)
```

---

## 2. TTL Cache for Key Lookup (带 TTL 的键查找缓存)

### 通用 TTL Cache 实现

```go
package cache

import (
	"sync"
	"time"
)

// TTLCache is a simple in-memory cache with TTL support
type TTLCache struct {
	mu         sync.RWMutex
	items      map[string]*cacheItem
	defaultTTL time.Duration
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewTTLCache creates a new TTL cache
func NewTTLCache(defaultTTL time.Duration) *TTLCache {
	cache := &TTLCache{
		items:      make(map[string]*cacheItem),
		defaultTTL: defaultTTL,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a value from cache
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(item.expiration) {
		return nil, false
	}
	
	return item.value, true
}

// Set stores a value in cache with default TTL
func (c *TTLCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with custom TTL
func (c *TTLCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Delete removes a key from cache
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}

// cleanup periodically removes expired items
func (c *TTLCache) cleanup() {
	ticker := time.NewTicker(c.defaultTTL)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Clear removes all items from cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*cacheItem)
}

// Len returns the number of items in cache
func (c *TTLCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}
```

### Token Cache 使用示例

```go
// 初始化 token cache
var tokenCache = NewTTLCache(60 * time.Second)

// 获取 token（先查缓存，再查数据库）
func GetTokenWithCache(key string) (*Token, error) {
	// Try cache first
	if cached, found := tokenCache.Get("token:" + key); found {
		if token, ok := cached.(*Token); ok {
			return token, nil
		}
	}
	
	// Cache miss, fetch from database
	var token Token
	if err := db.Where("key = ?", key).First(&token).Error; err != nil {
		return nil, err
	}
	
	// Store in cache
	tokenCache.Set("token:"+key, &token)
	
	return &token, nil
}

// 使缓存失效
func InvalidateTokenCache(key string) {
	tokenCache.Delete("token:" + key)
}
```

---

## 3. 优化的 HTTP Transport 配置

```go
package client

import (
	"net"
	"net/http"
	"time"
)

// CreateOptimizedHTTPClient creates an HTTP client optimized for single-user gateway
func CreateOptimizedHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		// Connection pool settings
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		
		// Dialer settings
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 90 * time.Second,
		}).DialContext,
		
		// TLS and timing settings
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		
		// Enable HTTP/2
		ForceAttemptHTTP2: true,
		
		// Disable compression for streaming (important for SSE)
		DisableCompression: false,
	}
	
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// 使用示例
var (
	// For regular API calls
	HTTPClient = CreateOptimizedHTTPClient(120 * time.Second)
	
	// For fast operations
	FastHTTPClient = CreateOptimizedHTTPClient(5 * time.Second)
	
	// For streaming (no timeout)
	StreamingHTTPClient = CreateOptimizedHTTPClient(0)
)
```

---

## 4. SQLite PRAGMA 初始化

```go
package database

import (
	"database/sql"
	"fmt"
)

// OptimizeSQLiteForGateway applies optimal PRAGMA settings for gateway workload
func OptimizeSQLiteForGateway(db *sql.DB) error {
	pragmas := []struct {
		name  string
		value string
	}{
		// WAL mode for better concurrency
		{"journal_mode", "WAL"},
		
		// NORMAL synchronous for balance
		{"synchronous", "NORMAL"},
		
		// 64MB cache
		{"cache_size", "-64000"},
		
		// 256MB memory-mapped I/O
		{"mmap_size", "268435456"},
		
		// Store temp tables in memory
		{"temp_store", "MEMORY"},
		
		// Busy timeout
		{"busy_timeout", "5000"},
		
		// Auto vacuum
		{"auto_vacuum", "INCREMENTAL"},
		
		// WAL checkpoint
		{"wal_autocheckpoint", "1000"},
		
		// Disable foreign keys for single-user performance
		{"foreign_keys", "OFF"},
	}
	
	for _, pragma := range pragmas {
		query := fmt.Sprintf("PRAGMA %s = %s", pragma.name, pragma.value)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to set %s: %w", pragma.name, err)
		}
		fmt.Printf("PRAGMA %s = %s\n", pragma.name, pragma.value)
	}
	
	return nil
}

// VerifySQLiteSettings verifies current PRAGMA settings
func VerifySQLiteSettings(db *sql.DB) {
	pragmas := []string{
		"journal_mode",
		"synchronous",
		"cache_size",
		"mmap_size",
		"temp_store",
	}
	
	fmt.Println("Current SQLite settings:")
	for _, pragma := range pragmas {
		var value string
		query := fmt.Sprintf("PRAGMA %s", pragma)
		db.QueryRow(query).Scan(&value)
		fmt.Printf("  %s = %s\n", pragma, value)
	}
}
```

---

## 5. 连接池优化配置

```go
package database

import (
	"database/sql"
	"time"
)

// ConfigureSQLiteConnectionPool sets optimal connection pool for SQLite
func ConfigureSQLiteConnectionPool(db *sql.DB) {
	// SQLite: use small connection pool to avoid SQLITE_BUSY
	db.SetMaxOpenConns(5)   // Allow some parallel reads with WAL
	db.SetMaxIdleConns(2)   // Keep 2 connections idle
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)
}

// ConfigureMySQLConnectionPool sets optimal connection pool for MySQL
func ConfigureMySQLConnectionPool(db *sql.DB) {
	// MySQL: can handle more connections
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)
}

// ConfigurePostgreSQLConnectionPool sets optimal connection pool for PostgreSQL
func ConfigurePostgreSQLConnectionPool(db *sql.DB) {
	// PostgreSQL: similar to MySQL
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)
}
```

---

## 6. 性能监控中间件

```go
package middleware

import (
	"fmt"
	"time"
	
	"github.com/gin-gonic/gin"
)

// PerformanceMetrics tracks request performance metrics
type PerformanceMetrics struct {
	Count       int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	P50Time     time.Duration
	P95Time     time.Duration
	P99Time     time.Duration
}

// PerformanceMonitor is a middleware for monitoring request performance
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Log slow requests
		if duration > 1*time.Second {
			fmt.Printf("[SLOW] %s %s took %v\n", 
				c.Request.Method, 
				c.Request.URL.Path, 
				duration)
		}
		
		// Set header for debugging
		c.Header("X-Response-Time", duration.String())
	}
}

// DBTimeTracker tracks database query time
func DBTimeTracker() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set start time in context
		c.Set("db_start_time", time.Now())
		
		c.Next()
		
		// Calculate DB time if available
		if startTime, exists := c.Get("db_start_time"); exists {
			if start, ok := startTime.(time.Time); ok {
				dbTime := time.Since(start)
				c.Header("X-DB-Time", dbTime.String())
			}
		}
	}
}
```

---

## 7. 流式响应优化

```go
package streaming

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
)

// OptimizedSSEWriter is an optimized SSE (Server-Sent Events) writer
type OptimizedSSEWriter struct {
	writer  *bufio.Writer
	flusher http.Flusher
}

// NewOptimizedSSEWriter creates a new optimized SSE writer
func NewOptimizedSSEWriter(w http.ResponseWriter) *OptimizedSSEWriter {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("ResponseWriter doesn't support flushing")
	}
	
	return &OptimizedSSEWriter{
		writer:  bufio.NewWriterSize(w, 4096), // 4KB buffer
		flusher: flusher,
	}
}

// WriteEvent writes an SSE event without unnecessary allocations
func (w *OptimizedSSEWriter) WriteEvent(data []byte) error {
	// Write event data
	if _, err := w.writer.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.writer.Write(data); err != nil {
		return err
	}
	if _, err := w.writer.Write([]byte("\n\n")); err != nil {
		return err
	}
	
	// Flush immediately for low latency
	if err := w.writer.Flush(); err != nil {
		return err
	}
	w.flusher.Flush()
	
	return nil
}

// Close flushes any remaining data
func (w *OptimizedSSEWriter) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	w.flusher.Flush()
	return nil
}

// ProxyStreamingResponse proxies a streaming response efficiently
func ProxyStreamingResponse(dst http.ResponseWriter, src io.Reader) error {
	writer := NewOptimizedSSEWriter(dst)
	defer writer.Close()
	
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 4096), 1024*1024) // 1MB max
	
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		
		if err := writer.WriteEvent(line); err != nil {
			return err
		}
	}
	
	return scanner.Err()
}
```

---

## 使用建议 / Usage Recommendations

1. **AsyncBatchWriter**: 用于日志、使用记录等可以容忍短暂延迟的写入
2. **TTLCache**: 用于 token、用户信息、渠道配置等频繁读取的数据
3. **HTTP Client**: 全局初始化一次，复用客户端
4. **SQLite PRAGMA**: 在数据库连接初始化时设置
5. **Connection Pool**: 根据数据库类型选择合适的配置
6. **Performance Monitor**: 作为 Gin 中间件使用
7. **SSE Writer**: 用于流式响应，减少 TTFT

---

## 性能注意事项 / Performance Notes

- Async writers 增加内存使用，但大幅减少 DB 写入
- TTL cache 需要定期清理，避免内存泄漏
- HTTP client 的 keep-alive 可以显著减少连接开销
- SQLite WAL 模式需要更多磁盘空间，但性能更好
- 流式响应要及时 flush，减少 TTFT

---

需要更多示例？查看完整文档: [PERFORMANCE_OPTIMIZATION.md](./PERFORMANCE_OPTIMIZATION.md)
