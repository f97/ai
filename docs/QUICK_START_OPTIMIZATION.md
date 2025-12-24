# One-API
# Quick Start Guide for Single-User Performance Optimization

## 🚀 快速开始

### 方案选择

：

#### 1.  (Conservative) -
```bash
cp .env.performance .env

#  .env，：
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_JOURNAL_MODE=WAL
LOCAL_CACHE_ENABLED=true
MEMORY_CACHE_ENABLED=true
```

**:**
- ✅ 20-30%

---

#### 2.  (Balanced) -
```bash
#  .env ：
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_JOURNAL_MODE=WAL
SQLITE_CACHE_SIZE=-64000
SQLITE_MMAP_SIZE=268435456

LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=60

BATCH_UPDATE_ENABLED=true
BATCH_UPDATE_INTERVAL=5

HTTP_MAX_IDLE_CONNS=100
HTTP_MAX_IDLE_CONNS_PER_HOST=20
```

**:**
- ✅ 40-60%
- ✅ 50%
- ⚠️

---

#### 3. 激进方案 (Aggressive) - 适合开发
```bash
#  .env ：
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_SYNCHRONOUS=NORMAL  #  OFF ()
SQLITE_JOURNAL_MODE=WAL
SQLITE_CACHE_SIZE=-128000  # 128MB
SQLITE_MMAP_SIZE=536870912  # 512MB

LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=120

LOG_BATCH_ENABLED=true
LOG_BATCH_SIZE=100
LOG_BATCH_FLUSH_INTERVAL=10

# LOG_CONSUME_ENABLED=false

BATCH_UPDATE_ENABLED=true
BATCH_UPDATE_INTERVAL=3

HTTP_MAX_IDLE_CONNS=150
HTTP_MAX_IDLE_CONNS_PER_HOST=30
```

**:**
- ✅ 60-80%
- ✅ 70%
- ⚠️  10
- ⚠️

---

## 📋 逐步启用指南

### Step 1:
```bash
systemctl stop one-api  #  docker stop one-api

cp one-api.db one-api.db.backup
cp one-api.db-wal one-api.db-wal.backup  # 
```

### Step 2:  SQLite
```bash
#  .env  docker-compose.yml
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL
```

### Step 3:
```bash
systemctl restart one-api
docker-compose restart
```

### Step 4:
```bash
journalctl -u one-api -n 50
docker logs one-api | tail -50

# [SYS] applying SQLite optimizations for single-user workload
# [SYS] SQLite PRAGMA: journal_mode = WAL
# [SYS] SQLite PRAGMA: synchronous = NORMAL
```

### Step 5:
```bash
#  ab  curl
time curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"Hello"}]
  }'
```

### Step 6:
```bash
#  Step 5 ，：
LOCAL_CACHE_ENABLED=true
BATCH_UPDATE_ENABLED=true

systemctl restart one-api

```

---

## 🔍 验证优化效果

###  SQLite
```bash
#  SQLite shell
sqlite3 one-api.db

PRAGMA journal_mode;
PRAGMA synchronous;
PRAGMA cache_size;
PRAGMA mmap_size;

# journal_mode = wal
# synchronous = 1 (NORMAL)
# cache_size = -64000
# mmap_size = 268435456
```

```bash
ls -lh one-api.db*

#  WAL
# one-api.db-wal  <-  WAL

top -p $(pgrep one-api)
```

---

## ⚠️ 常见问题

### Q:
**A:** ：
```bash
#  systemd
systemctl show one-api | grep Environment

#  Docker
docker inspect one-api | grep -A 20 Env
```

### Q:  (SQLITE_BUSY)
**A:** ：
```bash
SQLITE_MAX_OPEN_CONNS=3
SQLITE_BUSY_TIMEOUT=10000  # 10
```

### Q: WAL
**A:** 。：
```bash
sqlite3 one-api.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

### Q:
**A:** 
```bash
# 1.
systemctl stop one-api

# 2.
SQLITE_OPTIMIZE_ENABLED=false

# 3. ： DELETE
sqlite3 one-api.db "PRAGMA journal_mode=DELETE;"

# 4.
systemctl start one-api
```

---

## 📊 性能对比

- CPU: 2 cores
- RAM: 2GB
- Storage: SSD
- Load: Single user, moderate traffic

### 结果Example

| 配置 | p50 延迟 | p95 延迟 | 数据库写入
|------|----------|----------|---------------|------------|
|  | 250ms | 800ms | 20 | 25% |
|  | 200ms | 600ms | 15 | 22% |
|  | 150ms | 400ms | 8 | 18% |
|  | 100ms | 250ms | 2 | 15% |


---

## 🎯 下一步

1. ****
-  Prometheus + Grafana
-  pprof ( PERFORMANCE_OPTIMIZATION.md)

2. ****

3. ****
-  ab, wrk  vegeta

---

## 📚 更多资源

- [完整优化文档](.
- [SQLite WAL 模式](https:
- [Go 性能调优](https:

---

## 💡 Tips

1. ****: ，
2. ****: 
3. ****:  NORMAL ，
4. ****: ，

---

需要Help？提交 Issue: https:
