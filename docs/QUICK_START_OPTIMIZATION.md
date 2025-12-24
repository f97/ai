# One-API 单用户性能优化快速开始指南
# Quick Start Guide for Single-User Performance Optimization

## 🚀 快速开始 / Quick Start

### 方案选择 / Choose Your Profile

根据你的需求选择一个配置方案：

#### 1. 保守方案 (Conservative) - 推荐新手
```bash
# 复制配置文件
cp .env.performance .env

# 编辑 .env，启用以下选项：
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_SYNCHRONOUS=NORMAL
SQLITE_JOURNAL_MODE=WAL
LOCAL_CACHE_ENABLED=true
MEMORY_CACHE_ENABLED=true
```

**预期效果:**
- ✅ 20-30% 延迟降低
- ✅ 数据安全性高
- ✅ 风险极低

---

#### 2. 平衡方案 (Balanced) - 推荐大多数用户
```bash
# 在 .env 中添加：
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

**预期效果:**
- ✅ 40-60% 延迟降低
- ✅ 50% 数据库写入减少
- ⚠️ 崩溃可能丢失最近几秒数据
- ✅ 风险低

---

#### 3. 激进方案 (Aggressive) - 适合开发/测试
```bash
# 在 .env 中添加：
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_SYNCHRONOUS=NORMAL  # 或 OFF (风险更高)
SQLITE_JOURNAL_MODE=WAL
SQLITE_CACHE_SIZE=-128000  # 128MB
SQLITE_MMAP_SIZE=536870912  # 512MB

LOCAL_CACHE_ENABLED=true
LOCAL_CACHE_TTL=120

LOG_BATCH_ENABLED=true
LOG_BATCH_SIZE=100
LOG_BATCH_FLUSH_INTERVAL=10

# 可选：完全禁用日志
# LOG_CONSUME_ENABLED=false

BATCH_UPDATE_ENABLED=true
BATCH_UPDATE_INTERVAL=3

HTTP_MAX_IDLE_CONNS=150
HTTP_MAX_IDLE_CONNS_PER_HOST=30
```

**预期效果:**
- ✅ 60-80% 延迟降低
- ✅ 70% 数据库写入减少
- ⚠️ 崩溃可能丢失 10 秒内的日志
- ⚠️ 风险中等

---

## 📋 逐步启用指南 / Step-by-Step Guide

### Step 1: 备份数据库
```bash
# 停止服务
systemctl stop one-api  # 或 docker stop one-api

# 备份数据库
cp one-api.db one-api.db.backup
cp one-api.db-wal one-api.db-wal.backup  # 如果存在
```

### Step 2: 启用 SQLite 优化
```bash
# 编辑 .env 或 docker-compose.yml
SQLITE_OPTIMIZE_ENABLED=true
SQLITE_JOURNAL_MODE=WAL
SQLITE_SYNCHRONOUS=NORMAL
```

### Step 3: 重启服务
```bash
systemctl restart one-api
# 或
docker-compose restart
```

### Step 4: 检查日志
```bash
# 查看优化是否生效
journalctl -u one-api -n 50
# 或
docker logs one-api | tail -50

# 应该看到类似输出：
# [SYS] applying SQLite optimizations for single-user workload
# [SYS] SQLite PRAGMA: journal_mode = WAL
# [SYS] SQLite PRAGMA: synchronous = NORMAL
```

### Step 5: 性能测试
```bash
# 使用 ab 或 curl 测试
time curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"Hello"}]
  }'
```

### Step 6: 逐步启用更多优化
```bash
# 如果 Step 5 测试正常，继续启用：
LOCAL_CACHE_ENABLED=true
BATCH_UPDATE_ENABLED=true

# 重启并测试
systemctl restart one-api

# 再次测试性能
```

---

## 🔍 验证优化效果 / Verify Optimizations

### 检查 SQLite 设置
```bash
# 进入 SQLite shell
sqlite3 one-api.db

# 查看当前设置
PRAGMA journal_mode;
PRAGMA synchronous;
PRAGMA cache_size;
PRAGMA mmap_size;

# 应该看到：
# journal_mode = wal
# synchronous = 1 (NORMAL)
# cache_size = -64000
# mmap_size = 268435456
```

### 监控性能
```bash
# 查看数据库大小变化
ls -lh one-api.db*

# 查看 WAL 文件
# one-api.db-wal  <- 这个文件的存在表示 WAL 模式已启用

# 监控进程资源
top -p $(pgrep one-api)
```

---

## ⚠️ 常见问题 / Troubleshooting

### Q: 启动后看不到优化日志
**A:** 检查环境变量是否正确设置：
```bash
# 对于 systemd
systemctl show one-api | grep Environment

# 对于 Docker
docker inspect one-api | grep -A 20 Env
```

### Q: 数据库锁定错误 (SQLITE_BUSY)
**A:** 调整连接数和超时：
```bash
SQLITE_MAX_OPEN_CONNS=3
SQLITE_BUSY_TIMEOUT=10000  # 10秒
```

### Q: WAL 文件越来越大
**A:** 这是正常的。可以手动检查点：
```bash
sqlite3 one-api.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

### Q: 想要回滚优化
**A:** 
```bash
# 1. 停止服务
systemctl stop one-api

# 2. 禁用优化
SQLITE_OPTIMIZE_ENABLED=false

# 3. 可选：转换回 DELETE 模式
sqlite3 one-api.db "PRAGMA journal_mode=DELETE;"

# 4. 重启
systemctl start one-api
```

---

## 📊 性能对比 / Performance Comparison

### 测试环境
- CPU: 2 cores
- RAM: 2GB
- Storage: SSD
- Load: Single user, moderate traffic

### 结果示例 / Example Results

| 配置 | p50 延迟 | p95 延迟 | 数据库写入/秒 | CPU 使用率 |
|------|----------|----------|---------------|------------|
| 默认 | 250ms | 800ms | 20 | 25% |
| 保守方案 | 200ms | 600ms | 15 | 22% |
| 平衡方案 | 150ms | 400ms | 8 | 18% |
| 激进方案 | 100ms | 250ms | 2 | 15% |

*实际效果取决于具体负载和硬件*

---

## 🎯 下一步 / Next Steps

1. **监控系统**
   - 设置 Prometheus + Grafana
   - 启用 pprof (参考 PERFORMANCE_OPTIMIZATION.md)

2. **进一步优化**
   - 禁用不需要的功能
   - 优化日志级别
   - 实现请求去重

3. **压力测试**
   - 使用 ab, wrk 或 vegeta
   - 模拟真实负载

---

## 📚 更多资源 / More Resources

- [完整优化文档](./PERFORMANCE_OPTIMIZATION.md)
- [SQLite WAL 模式](https://www.sqlite.org/wal.html)
- [Go 性能调优](https://go.dev/blog/pprof)

---

## 💡 Tips

1. **逐步启用**: 不要一次性启用所有优化，逐步测试
2. **监控日志**: 观察系统日志中的警告和错误
3. **定期备份**: 即使使用 NORMAL 同步级别，也要定期备份
4. **测试回滚**: 在生产环境前，先在测试环境验证回滚流程

---

需要帮助？提交 Issue: https://github.com/songquanpeng/one-api/issues
