# One-API Performance Optimization Guide

## 单用户优化指南 / Single-User Optimization Guide

本文档介绍如何为单用户（single-user）场景优化 One-API 的性能，重点降低延迟、减少 CPU/RAM 使用。

This document describes how to optimize One-API performance for single-user scenarios, focusing on reducing latency and CPU/RAM usage.

---

## 📋 目录 / Table of Contents

- [Phase A: Quick Wins (≤ 2 hours)](#phase-a-quick-wins--2-hours)
- [Phase B: Medium Optimizations (≤ 1 day)](#phase-b-medium-optimizations--1-day)
- [Phase C: Deep Optimizations (≤ 2-3 days)](#phase-c-deep-optimizations--2-3-days)
- [Monitoring & Profiling](#monitoring--profiling)
- [Risk Analysis & Rollback](#risk-analysis--rollback)

---

## Phase A: Quick Wins (≤ 2 hours)

### 1. SQLite 优化 / SQLite Optimizations

#### 配置环境变量 / Configuration Environment Variables

```bash
# 启用 SQLite 优化 / Enable SQLite optimizations
export SQLITE_OPTIMIZE_ENABLED=true

# WAL 模式，提高并发性能 / WAL mode for better concurrency
export SQLITE_JOURNAL_MODE=WAL

# 同步级别 (FULL/NORMAL/OFF) / Synchronous level
# NORMAL: 平衡性能与安全 / Balance performance and safety
# OFF: 最快但崩溃可能丢数据 / Fastest but may lose data on crash
export SQLITE_SYNCHRONOUS=NORMAL

# 缓存大小 (负数表示 KB) / Cache size (negative = KB)
export SQLITE_CACHE_SIZE=-64000  # 64MB

# 内存映射大小 / Memory-mapped I/O size
export SQLITE_MMAP_SIZE=268435456  # 256MB

# 临时表存储在内存 / Temp tables in memory
export SQLITE_TEMP_STORE=MEMORY

# 繁忙超时 (毫秒) / Busy timeout (milliseconds)
export SQLITE_BUSY_TIMEOUT=5000

# WAL 自动检查点 / WAL auto-checkpoint
export SQLITE_WAL_AUTO_CHECKPOINT=1000

# 外键约束 (单用户建议关闭以提升性能) / Foreign keys (disable for single-user performance)
export SQLITE_FOREIGN_KEYS=false

# 连接池设置 (SQLite 建议小值) / Connection pool (small values for SQLite)
export SQLITE_MAX_OPEN_CONNS=5
export SQLITE_MAX_IDLE_CONNS=2
```

#### ⚠️ Trade-offs 权衡

| 配置 | 性能提升 | 风险 |
|------|----------|------|
| `SYNCHRONOUS=NORMAL` | ✅ 中等 | ⚠️ 崩溃可能丢失最近几个事务 |
| `SYNCHRONOUS=OFF` | ✅✅ 高 | ❌ 崩溃可能损坏数据库 |
| `FOREIGN_KEYS=false` | ✅ 小 | ⚠️ 不检查引用完整性 |
| `WAL mode` | ✅✅ 高 | ✅ 几乎无风险 |

**推荐配置 / Recommended Settings:**
- **生产环境 / Production**: `SYNCHRONOUS=NORMAL` + `WAL`
- **测试/开发 / Test/Dev**: `SYNCHRONOUS=OFF` 可接受

---

### 2. HTTP 客户端优化 / HTTP Client Optimizations

```bash
# Keep-alive 时间 / Keep-alive duration
export HTTP_KEEPALIVE=90  # 90秒

# 连接池设置 / Connection pool settings
export HTTP_MAX_IDLE_CONNS=100
export HTTP_MAX_IDLE_CONNS_PER_HOST=20
export HTTP_MAX_CONNS_PER_HOST=50

# 超时设置 / Timeout settings
export HTTP_DIAL_TIMEOUT=10
export HTTP_TLS_TIMEOUT=10
export HTTP_RESPONSE_HEADER_TIMEOUT=30
export HTTP_IDLE_CONN_TIMEOUT=90
```

**效果 / Benefits:**
- ✅ 减少 TCP 握手开销 / Reduce TCP handshake overhead
- ✅ 复用连接，降低延迟 / Reuse connections, lower latency
- ✅ 降低 upstream API 压力 / Reduce pressure on upstream APIs

---

## Phase B: Medium Optimizations (≤ 1 day)

### 3. 异步批量日志 / Async Batch Logging

```bash
# 启用批量日志 / Enable batch logging
export LOG_BATCH_ENABLED=true

# 批量大小 / Batch size
export LOG_BATCH_SIZE=50

# 刷新间隔 (秒) / Flush interval (seconds)
export LOG_BATCH_FLUSH_INTERVAL=5
```

**效果 / Benefits:**
- ✅✅ 大幅减少数据库写入次数 / Significantly reduce DB writes
- ✅ 降低流式请求开销 / Reduce streaming request overhead
- ⚠️ 崩溃可能丢失最近 5 秒的日志 / May lose last 5 seconds of logs on crash

**关闭日志功能 / Disable Logging (更激进 / More Aggressive):**

```bash
# 完全禁用消费日志 / Completely disable consume logs
export LOG_CONSUME_ENABLED=false
```

---

### 4. 本地内存缓存 / Local In-Memory Cache

```bash
# 启用本地缓存 / Enable local cache
export LOCAL_CACHE_ENABLED=true

# 缓存 TTL (秒) / Cache TTL (seconds)
export LOCAL_CACHE_TTL=60
```

**缓存内容 / Cached Data:**
- Token → User/Channel mapping
- User quota
- Model routing configuration

**效果 / Benefits:**
- ✅ 减少数据库查询 / Reduce DB queries
- ✅ 降低 token 验证延迟 / Lower token validation latency
- ✅ 适合单用户场景 / Perfect for single-user

---

### 5. 启用批量更新 / Enable Batch Updates

```bash
# 批量更新配额等信息 / Batch update quota and stats
export BATCH_UPDATE_ENABLED=true
export BATCH_UPDATE_INTERVAL=5  # 5秒刷新一次
```

---

## Phase C: Deep Optimizations (≤ 2-3 days)

### 6. 性能监控 / Performance Monitoring

#### 启用 pprof (Go 性能分析)

在 `main.go` 中添加：

```go
import _ "net/http/pprof"

// In main() function
go func() {
    logger.SysLog("pprof server started on :6060")
    http.ListenAndServe(":6060", nil)
}()
```

#### 访问 pprof 端点 / Access pprof Endpoints

```bash
# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Heap profiling
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Blocking
curl http://localhost:6060/debug/pprof/block?debug=2
```

### 7. 指标收集 / Metrics Collection

**关键指标 / Key Metrics:**

- **延迟 / Latency**: p50, p95, p99 response times
- **TTFT**: Time to first token (streaming)
- **数据库时间 / DB Time**: Query duration
- **锁等待 / Lock Wait**: Mutex contention
- **GC 暂停 / GC Pause**: Garbage collection pauses
- **内存分配 / Memory Allocation**: Bytes allocated per request

**实现方式 / Implementation:**

使用 Prometheus + Grafana 或自定义 metrics:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "oneapi_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "status"},
    )
    
    dbQueryDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "oneapi_db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(dbQueryDuration)
}

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

---

## 🎯 Priority Checklist 优先级检查清单

### 高优先级 / High Priority (立即实施)

- [ ] ✅ 启用 SQLite WAL 模式
- [ ] ✅ 设置 `SYNCHRONOUS=NORMAL`
- [ ] ✅ 配置 HTTP 客户端 keep-alive
- [ ] ✅ 启用本地内存缓存 (`LOCAL_CACHE_ENABLED=true`)
- [ ] ✅ 禁用不必要的日志 (`LOG_CONSUME_ENABLED=false`)

### 中优先级 / Medium Priority

- [ ] ⚠️ 启用异步批量日志 (`LOG_BATCH_ENABLED=true`)
- [ ] ⚠️ 启用批量更新 (`BATCH_UPDATE_ENABLED=true`)
- [ ] ⚠️ 调整 SQLite 缓存大小
- [ ] ⚠️ 优化连接池设置

### 低优先级 / Low Priority (可选)

- [ ] 🔧 添加 pprof 监控
- [ ] 🔧 实现 Prometheus metrics
- [ ] 🔧 流式路径优化
- [ ] 🔧 减少 marshal/unmarshal

---

## 🛡️ Risk Analysis & Rollback 风险分析与回滚

### 风险等级 / Risk Levels

| 配置 | 风险 | 数据丢失可能性 | 回滚方法 |
|------|------|----------------|----------|
| WAL mode | 低 | 极低 | `SQLITE_JOURNAL_MODE=DELETE` |
| `SYNCHRONOUS=NORMAL` | 中 | 低 (最近几秒) | `SQLITE_SYNCHRONOUS=FULL` |
| `SYNCHRONOUS=OFF` | 高 | 中 (可能损坏) | ❌ 不推荐使用 |
| Batch logging | 中 | 低 (5秒内) | `LOG_BATCH_ENABLED=false` |
| Local cache | 低 | 无 | `LOCAL_CACHE_ENABLED=false` |
| HTTP optimizations | 低 | 无 | 恢复默认值 |

### 回滚计划 / Rollback Plan

```bash
# 1. 禁用所有优化 / Disable all optimizations
export SQLITE_OPTIMIZE_ENABLED=false
export LOCAL_CACHE_ENABLED=false
export LOG_BATCH_ENABLED=false

# 2. 恢复默认 SQLite 设置 / Restore default SQLite settings
export SQLITE_SYNCHRONOUS=FULL
export SQLITE_JOURNAL_MODE=DELETE

# 3. 重启服务 / Restart service
systemctl restart one-api
```

---

## 📊 性能测试方法 / Performance Testing

### 基准测试 / Benchmark Testing

```bash
# 1. 记录基准性能 / Record baseline performance
ab -n 1000 -c 10 http://localhost:3000/v1/chat/completions

# 2. 启用优化 / Enable optimizations
# ... apply optimizations ...

# 3. 再次测试 / Test again
ab -n 1000 -c 10 http://localhost:3000/v1/chat/completions

# 4. 对比结果 / Compare results
```

### 流式测试 / Streaming Test

```bash
# 测试 TTFT (Time To First Token)
time curl -N http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }' | head -n 1
```

---

## 🔍 常见问题 / FAQ

### Q1: 使用 `SYNCHRONOUS=NORMAL` 安全吗？

**A:** 对于单用户场景，`NORMAL` 是推荐的平衡选项。它在性能和安全性之间取得良好平衡。只有在突然断电的情况下，才可能丢失最近的几个事务。

### Q2: 批量日志会丢失多少数据？

**A:** 最多丢失 `LOG_BATCH_FLUSH_INTERVAL` 秒内的日志（默认 5 秒）。对于单用户场景，这通常是可以接受的。

### Q3: 如何监控性能改善？

**A:** 使用 pprof 监控 CPU 和内存使用，使用 Prometheus 监控请求延迟。对比优化前后的 p95/p99 延迟。

### Q4: 是否应该禁用所有日志？

**A:** 不推荐。至少保留错误日志和关键操作日志。可以禁用详细的 request body 日志和消费日志。

---

## 📚 参考资料 / References

- [SQLite PRAGMA Documentation](https://www.sqlite.org/pragma.html)
- [Go net/http Performance](https://go.dev/blog/http-tracing)
- [WAL Mode](https://www.sqlite.org/wal.html)
- [pprof Guide](https://go.dev/blog/pprof)

---

## 📞 Support 支持

如有问题，请提交 Issue 或参考原项目文档:
- [One-API GitHub](https://github.com/songquanpeng/one-api)
