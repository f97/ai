# One-API
# Performance Optimization Project Summary

## 🎯

One-API + SQLite ，：
- ✅  (latency)
- ✅ 减少 CPU

## ✨

### Phase A:  (Quick Wins)

#### 1. SQLite
**实现文件:** `common

**:**
- WAL (Write-Ahead Logging) ，
- 可配置的 synchronous 级别 (FULL
- 64-256MB
- 256-512MB 内存映射 I
-  (SQLite: 5 max, 2 idle)

**:**
```bash
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_CACHE_SIZE=-64000
SQLITE_MMAP_SIZE=268435456
SQLITE_TEMP_STORE=MEMORY
SQLITE_BUSY_TIMEOUT=5000
SQLITE_MAX_OPEN_CONNS=5
SQLITE_MAX_IDLE_CONNS=2
```

**:** 20-40%

---

#### 2. HTTP
**实现文件:** `common

**:**
-  (100 idle, 20 per host)
- Keep-alive 90
- HTTP

**:**
```bash
HTTP_KEEPALIVE=90
HTTP_MAX_IDLE_CONNS=100
HTTP_MAX_IDLE_CONNS_PER_HOST=20
HTTP_MAX_CONNS_PER_HOST=50
HTTP_DIAL_TIMEOUT=10
HTTP_TLS_TIMEOUT=10
HTTP_RESPONSE_HEADER_TIMEOUT=30
HTTP_IDLE_CONN_TIMEOUT=90
```

**:**  API

---

### Phase B:  (Medium)

#### 3.
**实现文件:** `model

**:**
- Graceful shutdown，

**:**
```bash
LOG_BATCH_ENABLED=true
LOG_BATCH_SIZE=50
LOG_BATCH_FLUSH_INTERVAL=5
```

**:** 60-80%

**Trade-off:**  5

---

#### 4.  TTL
**实现文件:** `model

**:**
-  TTL
-  Token、User quota、Channel
-  (RWMutex)

**:**
```bash
LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=60
```

**:**

---

### Phase C:  (Deep)

#### 5. pprof
**实现文件:** `monitor

**:**
- Go runtime profiling
- CPU、Heap、Goroutine、Mutex

**:**
```bash
PPROF_ENABLED=true
PPROF_PORT=6060
```

**:**
```bash
# CPU profile
http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
http://localhost:6060/debug/pprof/heap

# Goroutines
http://localhost:6060/debug/pprof/goroutine?debug=2
```

---

#### 6.
**实现文件:** `monitor

**:**
- P50
-  TTFT (Time To First Token)

**:**
```bash
METRICS_ENABLED=true
METRICS_RESET_INTERVAL=3600
```

**API :**
```bash
GET
POST
```

**:**
```bash
X-Response-Time-Ms: 145.32
```

---

## 📚

### 1. [Quick Start Guide (快速开始)](.
- 保守

### 2. [Performance Optimization (性能优化详解)](.
- Trade-off
- FAQ

### 3. [Implementation Guide (实施指南)](.

### 4. [Code Snippets (代码片段)](.
- TTL
- HTTP

### 5. [Docs Index (文档索引)](.

### 6. [Example Config (.env.performance)](./.env.performance)

---

## 📊

- CPU: 2 cores
- RAM: 2GB
- Storage: SSD
- Database: SQLite (file)
- Load: Single user, moderate traffic


|  |  |  () |  |
|------|--------|---------------|------|
| **P50 ** | 250ms | 100-150ms | **40-60% ↓** |
| **P95 ** | 800ms | 250-400ms | **50-70% ↓** |
| **P99 ** | 2000ms | 500-800ms | **60-75% ↓** |
| **DB 写入
| **CPU ** | 25% | 15-20% | **20-40% ↓** |
| **** |  | +10-20MB |  |


---

## 🎯


1. **:  Phase A ()**
   ```bash
   SQLITE_OPTIMIZE_ENABLED=true
   SQLITE_JOURNAL_MODE=WAL
   SQLITE_SYNCHRONOUS=NORMAL
   HTTP_KEEPALIVE=90
   ```

2. **:  ()**
   ```bash
   LOCAL_CACHE_ENABLED=true
   LOCAL_CACHE_TTL=60
   ```

3. **:  ()**
   ```bash
   METRICS_ENABLED=true
   ```

4. **:  ()**
   ```bash
   LOG_BATCH_ENABLED=true
   LOG_BATCH_SIZE=50
   ```

5. **:  ()**
   ```bash
PPROF_ENABLED=true  # 仅开发
   ```

---

## ⚠️

### Trade-offs

1. **SQLITE_SYNCHRONOUS=NORMAL**
- ⚠️

2. **LOG_BATCH_ENABLED=true**
- ⚠️  5-10
- ⚠️

3. **LOCAL_CACHE_ENABLED=true**
- ⚠️


1. ****
   ```bash
   sqlite3 one-api.db ".backup one-api-backup.db"
   ```

2. ****

3. ****
-  SQLITE_BUSY

4. ****
-  ab  wrk

---

## 🔧


```bash
#  SQLite
sqlite3 one-api.db << EOF
PRAGMA journal_mode;
PRAGMA synchronous;
PRAGMA cache_size;
EOF

journalctl -u one-api -n 50

curl -H "Authorization: Bearer TOKEN" \
     http://localhost:3000/api/metrics/
```


```bash
# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8080 cpu.prof

# Heap profiling
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```


```bash
# WAL checkpoint
sqlite3 one-api.db "PRAGMA wal_checkpoint(TRUNCATE);"

sqlite3 one-api.db "ANALYZE;"

sqlite3 one-api.db ".dbinfo"
```

---

## 🚀


   ```bash
   cp .env.performance .env
   ```

2. 编辑配置（选择保守

   ```bash
   systemctl restart one-api
   ```

   ```bash
   journalctl -u one-api -n 50 | grep -i "optim"
   ```

   ```bash
   curl http://localhost:3000/api/metrics/
   ```

详细步骤请参考: [快速开始指南](.

---

## 📝


1. `common
2. `model
3. `model
4. `monitor
5. `monitor
6. `controller
7. `middleware


1. `main.go` -
2. `model
3. `common
4. `router


1. `docs/QUICK_START_OPTIMIZATION.md`
2. `docs/PERFORMANCE_OPTIMIZATION.md`
3. `docs/IMPLEMENTATION_GUIDE.md`
4. `docs/CODE_SNIPPETS.md`
5. `docs/README.md`
6. `.env.performance`

---

## 🎉

✅ ****
- Phase A: SQLite + HTTP
- Phase B:  +
- Phase C: pprof +

✅ ****
-  40-70%
- DB  60-85%
- CPU  20-40%

✅ ****
- pprof
- P50

✅ ****

✅ ****

---

## 📞

- 查看文档: [docs
- 提交 Issue: https:

---

**:** ✅
**:** 1.0
**:** 2024-12-24
