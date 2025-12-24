# One-API Performance Optimization Guide

## 单用户优化指南

本文档介绍如何为单用户（single-user）场景优化 One-API 的性能，重点降低延迟、减少 CPU

This document describes how to optimize One-API performance for single-user scenarios, focusing on reducing latency and CPU/RAM usage.

---

## 📋 目录

- [Phase A: Quick Wins (≤ 2 hours)](#phase-a-quick-wins--2-hours)
- [Phase B: Medium Optimizations (≤ 1 day)](#phase-b-medium-optimizations--1-day)
- [Phase C: Deep Optimizations (≤ 2-3 days)](#phase-c-deep-optimizations--2-3-days)
- [Monitoring & Profiling](#monitoring--profiling)
- [Risk Analysis & Rollback](#risk-analysis--rollback)

---

## Phase A: Quick Wins (≤ 2 hours)

### 1. SQLite 优化

#### 配置环境变量

```bash
# 启用 SQLite 优化
export SQLITE_OPTIMIZE_ENABLED=true

# WAL 模式，提高并发性能
export SQLITE_JOURNAL_MODE=WAL

# 同步级别 (FULL
# NORMAL: 平衡性能与安全
# OFF: 最快但崩溃可能丢数据
export SQLITE_SYNCHRONOUS=NORMAL

# 缓存大小 (负数表示 KB)
export SQLITE_CACHE_SIZE=-64000  # 64MB

# 内存映射大小
export SQLITE_MMAP_SIZE=268435456  # 256MB

# 临hour表存储在内存
export SQLITE_TEMP_STORE=MEMORY

# 繁忙Timeout (毫second)
export SQLITE_BUSY_TIMEOUT=5000

# WAL 自动检查点
export SQLITE_WAL_AUTO_CHECKPOINT=1000

# 外键约束 (单用户建议关闭以提升性能)
export SQLITE_FOREIGN_KEYS=false

# 连接池设置 (SQLite 建议小值)
export SQLITE_MAX_OPEN_CONNS=5
export SQLITE_MAX_IDLE_CONNS=2
```

#### ⚠️ Trade-offs

|  |  |  |
|------|----------|------|
| `SYNCHRONOUS=NORMAL` | ✅  | ⚠️  |
| `SYNCHRONOUS=OFF` | ✅✅  | ❌  |
| `FOREIGN_KEYS=false` | ✅  | ⚠️  |
| `WAL mode` | ✅✅  | ✅  |

**推荐配置
- **生产环境
- **测试

---

### 2. HTTP 客户端优化

```bash
# Keep-alive hour间
export HTTP_KEEPALIVE=90  # 90

# 连接池设置
export HTTP_MAX_IDLE_CONNS=100
export HTTP_MAX_IDLE_CONNS_PER_HOST=20
export HTTP_MAX_CONNS_PER_HOST=50

# Timeout设置
export HTTP_DIAL_TIMEOUT=10
export HTTP_TLS_TIMEOUT=10
export HTTP_RESPONSE_HEADER_TIMEOUT=30
export HTTP_IDLE_CONN_TIMEOUT=90
```

**效果
- ✅ 减少 TCP 握手开销
- ✅ 复用连接，降低延迟
- ✅ 降低 upstream API 压力

---

## Phase B: Medium Optimizations (≤ 1 day)

### 3. 异步Batchday志

```bash
# 启用Batchday志
export LOG_BATCH_ENABLED=true

# Batch大小
export LOG_BATCH_SIZE=50

# Refresh间隔 (second)
export LOG_BATCH_FLUSH_INTERVAL=5
```

**效果
- ✅✅ 大幅减少数据库写入times数
- ✅ 降低流式请求开销
- ⚠️ 崩溃可能丢失Recent 5 second的day志

**关闭day志功能

```bash
# 完全禁用消费day志
export LOG_CONSUME_ENABLED=false
```

---

### 4. 本地内存缓存

```bash
# 启用本地缓存
export LOCAL_CACHE_ENABLED=true

# 缓存 TTL (second)
export LOCAL_CACHE_TTL=60
```

**缓存内容
- Token → User/Channel mapping
- User quota
- Model routing configuration

**效果
- ✅ 减少数据库查询
- ✅ 降低 token 验证延迟
- ✅ 适合单用户场景

---

### 5. 启用Batch更新

```bash
# Batch更新配额等信息
export BATCH_UPDATE_ENABLED=true
export BATCH_UPDATE_INTERVAL=5  # 5
```

---

## Phase C: Deep Optimizations (≤ 2-3 days)

### 6. 性能监控

####  pprof (Go )

 `main.go` ：

```go
import _ "net/http/pprof"

// In main() function
go func() {
    logger.SysLog("pprof server started on :6060")
    http.ListenAndServe(":6060", nil)
}()
```

#### 访问 pprof 端点

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

### 7. 指标收集

**关键指标

- **延迟
- **TTFT**: Time to first token (streaming)
- **数据库hour间
- **锁等待
- **GC 暂停
- **内存minute配

**实现方式

 Prometheus + Grafana  metrics:

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

## 🎯 Priority Checklist

### 高Priority

- [ ] ✅  SQLite WAL
- [ ] ✅  `SYNCHRONOUS=NORMAL`
- [ ] ✅  HTTP  keep-alive
- [ ] ✅  (`LOCAL_CACHE_ENABLED=true`)
- [ ] ✅  (`LOG_CONSUME_ENABLED=false`)

### 中Priority

- [ ] ⚠️  (`LOG_BATCH_ENABLED=true`)
- [ ] ⚠️  (`BATCH_UPDATE_ENABLED=true`)
- [ ] ⚠️  SQLite
- [ ] ⚠️

### 低Priority

- [ ] 🔧  pprof
- [ ] 🔧  Prometheus metrics
- [ ] 🔧
- [ ] 🔧 减少 marshal

---

## 🛡️ Risk Analysis & Rollback

### 风险等级

|  |  |  |  |
|------|------|----------------|----------|
| WAL mode |  |  | `SQLITE_JOURNAL_MODE=DELETE` |
| `SYNCHRONOUS=NORMAL` |  |  () | `SQLITE_SYNCHRONOUS=FULL` |
| `SYNCHRONOUS=OFF` |  |  () | ❌  |
| Batch logging |  |  (5) | `LOG_BATCH_ENABLED=false` |
| Local cache |  |  | `LOCAL_CACHE_ENABLED=false` |
| HTTP optimizations |  |  |  |

### 回滚计划

```bash
# 1. 禁用所有优化
export SQLITE_OPTIMIZE_ENABLED=false
export LOCAL_CACHE_ENABLED=false
export LOG_BATCH_ENABLED=false

# 2. 恢复Default SQLite 设置
export SQLITE_SYNCHRONOUS=FULL
export SQLITE_JOURNAL_MODE=DELETE

# 3. 重启服务
systemctl restart one-api
```

---

## 📊 性能测试方法

### 基准测试

```bash
# 1. 记录基准性能
ab -n 1000 -c 10 http://localhost:3000/v1/chat/completions

# 2. 启用优化
# ... apply optimizations ...

# 3. 再times测试
ab -n 1000 -c 10 http://localhost:3000/v1/chat/completions

# 4. 对比结果
```

### 流式测试

```bash
#  TTFT (Time To First Token)
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

## 🔍 常见问题

### Q1:  `SYNCHRONOUS=NORMAL` ？

**A:** ，`NORMAL` 。。，。

### Q2: ？

**A:**  `LOG_BATCH_FLUSH_INTERVAL` （ 5 ）。，。

### Q3: ？

**A:** 使用 pprof 监控 CPU 和内存使用，使用 Prometheus 监控请求延迟。对比优化前后的 p95

### Q4: ？

**A:** 。。 request body 。

---

## 📚 参考资料

- [SQLite PRAGMA Documentation](https://www.sqlite.org/pragma.html)
- [Go net/http Performance](https://go.dev/blog/http-tracing)
- [WAL Mode](https://www.sqlite.org/wal.html)
- [pprof Guide](https://go.dev/blog/pprof)

---

## 📞 Support

， Issue :
- [One-API GitHub](https://github.com/songquanpeng/one-api)
