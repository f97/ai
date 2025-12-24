# One-API
# Complete Performance Optimization Implementation

，、。

---

## 📋

1. [](#)
2. [Phase A: ](#phase-a-)
3. [Phase B: ](#phase-b-)
4. [Phase C: ](#phase-c-)
5. [](#)
6. [](#)
7. [](#)

---



#### ✅ Phase A:  ()
- SQLite PRAGMA （WAL, synchronous, cache_size, mmap）
- HTTP
-  keep-alive

#### ✅ Phase B:  ()
-  TTL

#### ✅ Phase C:  ()
- pprof
- 性能指标收集系统（p50


|  |  |  |  |
|------|--------|--------|------|
| P50  | 250ms | 100-150ms | 40-60% ↓ |
| P95  | 800ms | 250-400ms | 50-70% ↓ |
| P99  | 2000ms | 500-800ms | 60-75% ↓ |
| DB 写入
| CPU  | 25% | 15-20% | 20-40% ↓ |
|  | +10-20MB | +10-20MB |  |

---

## Phase A:

### 1. SQLite

####

 `.env` ：

```bash
# SQLite  -
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_CACHE_SIZE=-64000      # 64MB
SQLITE_MMAP_SIZE=268435456    # 256MB
SQLITE_TEMP_STORE=MEMORY
SQLITE_BUSY_TIMEOUT=5000
SQLITE_WAL_AUTO_CHECKPOINT=1000
SQLITE_FOREIGN_KEYS=false
SQLITE_MAX_OPEN_CONNS=5
SQLITE_MAX_IDLE_CONNS=2
```

####

```bash
journalctl -u one-api -n 100 | grep -i "sqlite"

# [SYS] applying SQLite optimizations for single-user workload
# [SYS] SQLite PRAGMA: journal_mode = WAL
# [SYS] SQLite PRAGMA: synchronous = NORMAL
```

####

```bash
sqlite3 one-api.db << EOF
PRAGMA journal_mode;
PRAGMA synchronous;
PRAGMA cache_size;
PRAGMA mmap_size;
EOF

# wal
# 1
# -64000
# 268435456
```

### 2. HTTP

```bash
# HTTP
HTTP_KEEPALIVE=90
HTTP_MAX_IDLE_CONNS=100
HTTP_MAX_IDLE_CONNS_PER_HOST=20
HTTP_MAX_CONNS_PER_HOST=50
HTTP_DIAL_TIMEOUT=10
HTTP_TLS_TIMEOUT=10
HTTP_RESPONSE_HEADER_TIMEOUT=30
HTTP_IDLE_CONN_TIMEOUT=90
```

---

## Phase B:

### 3.

```bash
#  ()
LOG_BATCH_ENABLED=true
LOG_BATCH_SIZE=50
LOG_BATCH_FLUSH_INTERVAL=5

# LOG_CONSUME_ENABLED=false
```

**⚠️ :**
-  5

### 4.

```bash
#  TTL
LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=60  # 60
```

**:**
- Token → User

### 5.

```bash
#  ()
BATCH_UPDATE_ENABLED=true
BATCH_UPDATE_INTERVAL=5
```

---

## Phase C:

### 6.  (pprof)

####  pprof

```bash
#  pprof
PPROF_ENABLED=true
PPROF_PORT=6060
```

####  pprof

```bash
# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
(pprof) top10
(pprof) web

# Heap profiling
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

#  goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=2

curl http://localhost:6060/debug/pprof/block?debug=2
curl http://localhost:6060/debug/pprof/mutex?debug=2
```

### 7.

####

```bash
METRICS_ENABLED=true
METRICS_RESET_INTERVAL=3600  # 
```

####

```bash
curl -H "Authorization: Bearer ADMIN_TOKEN" \
     http://localhost:3000/api/metrics/

{
  "success": true,
  "data": {
    "total_requests": 12450,
    "failed_requests": 23,
    "success_rate": 99.81,
    "average_duration": "145ms",
    "p50_latency": "120ms",
    "p95_latency": "380ms",
    "p99_latency": "650ms",
    "db_query_count": 8234,
    "db_avg_duration": "12ms",
    "streaming_requests": 340,
    "avg_ttft": "450ms"
  }
}

curl -X POST -H "Authorization: Bearer ADMIN_TOKEN" \
     http://localhost:3000/api/metrics/reset
```

####

：

```bash
X-Response-Time-Ms: 145.32
```

。

---


### 1.

####

```bash
# CPU
top -p $(pgrep one-api)

# 磁盘 I
iostat -x 1

ss -s
netstat -an | grep ESTABLISHED | wc -l
```

####

```bash
# SQLite
ls -lh one-api.db*

# WAL （）
ls -lh one-api.db-wal

sqlite3 one-api.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

### 2.

```bash
journalctl -u one-api | grep "slow request"

journalctl -u one-api --since "1 hour ago" | \
    grep "X-Response-Time" | \
    awk '{print $NF}' | \
    sort -n | \
    tail -100

journalctl -u one-api -p err --since "1 hour ago"
```

### 3.

```bash
ab -n 1000 -c 10 \
   -H "Authorization: Bearer YOUR_TOKEN" \
   -H "Content-Type: application/json" \
   -p request.json \
   http://localhost:3000/v1/chat/completions

# request.json:
{
  "model": "gpt-3.5-turbo",
  "messages": [{"role":"user","content":"Hello"}],
  "max_tokens": 50
}

for i in {1..100}; do
  curl -N http://localhost:3000/v1/chat/completions \
    -H "Authorization: Bearer YOUR_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hi"}],"stream":true}' \
    > /dev/null 2>&1 &
done
wait
```

### 4. Prometheus + Grafana ()

TODO:  Prometheus metrics 

---



#### 1. SQLite

**:** `SQLITE_BUSY`

**:**
```bash
SQLITE_BUSY_TIMEOUT=10000

SQLITE_MAX_OPEN_CONNS=3
```

#### 2. WAL

**:** `one-api.db-wal`

**:**
```bash
sqlite3 one-api.db "PRAGMA wal_checkpoint(TRUNCATE);"

SQLITE_WAL_AUTO_CHECKPOINT=500
```

#### 3.

**:**

**:**

**:**
```bash
#  TTL
LOCAL_CACHE_TTL=30

LOG_BATCH_SIZE=20
```

#### 4.

**:**

1. ：
```bash
curl -H "Authorization: Bearer TOKEN" \
     http://localhost:3000/api/metrics/
```

2.  pprof：
```bash
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8080 cpu.prof
```

3.  API：
```bash
time curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer UPSTREAM_KEY"
```

---



```bash
# 1.
systemctl stop one-api

# 2.
cat > .env << EOF
SQLITE_OPTIMIZE_ENABLED=false
LOCAL_CACHE_ENABLED=false
LOG_BATCH_ENABLED=false
METRICS_ENABLED=false
PPROF_ENABLED=false
EOF

# 3.  SQLite  DELETE
sqlite3 one-api.db << EOF
PRAGMA journal_mode=DELETE;
PRAGMA synchronous=FULL;
EOF

# 4.
systemctl start one-api
```


####  SQLite

```bash
SQLITE_OPTIMIZE_ENABLED=false
SQLITE_SYNCHRONOUS=FULL
SQLITE_JOURNAL_MODE=DELETE
```

####

```bash
LOG_BATCH_ENABLED=false
```

####

```bash
LOCAL_CACHE_ENABLED=false
```

---


### （）

```bash
# SQLite
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_CACHE_SIZE=-32000

# HTTP
HTTP_KEEPALIVE=60
HTTP_MAX_IDLE_CONNS=50

LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=60

LOG_BATCH_ENABLED=false

METRICS_ENABLED=true
PPROF_ENABLED=false  # 
```

### 激进配置（测试

```bash
# SQLite
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL  #  OFF
SQLITE_CACHE_SIZE=-128000
SQLITE_MMAP_SIZE=536870912

# HTTP
HTTP_KEEPALIVE=120
HTTP_MAX_IDLE_CONNS=200
HTTP_MAX_IDLE_CONNS_PER_HOST=50

LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=120
LOG_BATCH_ENABLED=true
LOG_BATCH_SIZE=100
LOG_CONSUME_ENABLED=false

METRICS_ENABLED=true
PPROF_ENABLED=true
```

---


：

✅ **Phase A**: SQLite PRAGMA + HTTP   
✅ **Phase B**:  +   
✅ **Phase C**: pprof  +   

**：**
-  40-70%
-  60-85%
- CPU  20-40%

**：**

---


- [快速开始指南](.
- [详细优化文档](.
- [代码片段集合](.

---

**需要帮助？** 提交 Issue 到原项目: https:
