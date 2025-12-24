package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
)

// Local in-memory cache for single-user workload
// This reduces DB queries for frequently accessed data like tokens, user info, and channel config

var (
	localCacheEnabled = env.Bool("LOCAL_CACHE_ENABLED", false)
	localCacheTTL     = env.Int("LOCAL_CACHE_TTL", 60) // seconds
)

// cacheEntry represents a cached item with expiration
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// localCache is a simple in-memory cache with TTL
type localCache struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
}

var (
	tokenCache   *localCache
	userCache    *localCache
	channelCache *localCache
	cacheOnce    sync.Once
)

// InitLocalCache initializes the local in-memory cache
func InitLocalCache() {
	if !localCacheEnabled {
		return
	}

	cacheOnce.Do(func() {
		tokenCache = &localCache{items: make(map[string]*cacheEntry)}
		userCache = &localCache{items: make(map[string]*cacheEntry)}
		channelCache = &localCache{items: make(map[string]*cacheEntry)}

		logger.SysLog("local in-memory cache enabled for single-user optimization")
		logger.SysLog(fmt.Sprintf("cache TTL: %d seconds", localCacheTTL))

		// Start cleanup goroutine
		go cleanupExpiredCache()
	})
}

// cleanupExpiredCache periodically removes expired cache entries
func cleanupExpiredCache() {
	ticker := time.NewTicker(time.Duration(localCacheTTL) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cleanupCache(tokenCache)
		cleanupCache(userCache)
		cleanupCache(channelCache)
	}
}

// cleanupCache removes expired entries from a cache
func cleanupCache(cache *localCache) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now()
	for key, entry := range cache.items {
		if now.After(entry.expiration) {
			delete(cache.items, key)
		}
	}
}

// get retrieves a value from cache
func (c *localCache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

// set stores a value in cache
func (c *localCache) set(key string, value interface{}, ttl int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(time.Duration(ttl) * time.Second),
	}
}

// delete removes a value from cache
func (c *localCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// LocalCacheGetToken retrieves a token from cache or DB
func LocalCacheGetToken(ctx context.Context, key string) (*Token, error) {
	if !localCacheEnabled {
		return CacheGetTokenByKey(key)
	}

	// Try cache first
	cacheKey := "token:" + key
	if cached, found := tokenCache.get(cacheKey); found {
		if token, ok := cached.(*Token); ok {
			logger.Debug(ctx, "token cache hit: "+key)
			return token, nil
		}
	}

	// Cache miss, fetch from DB/Redis
	token, err := CacheGetTokenByKey(key)
	if err != nil {
		return nil, err
	}

	// Store in cache
	tokenCache.set(cacheKey, token, localCacheTTL)
	logger.Debug(ctx, "token cache miss, fetched from DB: "+key)

	return token, nil
}

// LocalCacheInvalidateToken removes a token from cache
func LocalCacheInvalidateToken(key string) {
	if !localCacheEnabled {
		return
	}
	tokenCache.delete("token:" + key)
}

// LocalCacheGetUserQuota retrieves user quota from cache or DB
func LocalCacheGetUserQuota(ctx context.Context, userId int) (int64, error) {
	if !localCacheEnabled {
		return CacheGetUserQuota(ctx, userId)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("user_quota:%d", userId)
	if cached, found := userCache.get(cacheKey); found {
		if quota, ok := cached.(int64); ok {
			logger.Debug(ctx, "user quota cache hit")
			return quota, nil
		}
	}

	// Cache miss, fetch from DB/Redis
	quota, err := CacheGetUserQuota(ctx, userId)
	if err != nil {
		return 0, err
	}

	// Store in cache
	userCache.set(cacheKey, quota, localCacheTTL)

	return quota, nil
}

// LocalCacheInvalidateUserQuota removes user quota from cache
func LocalCacheInvalidateUserQuota(userId int) {
	if !localCacheEnabled {
		return
	}
	userCache.delete(fmt.Sprintf("user_quota:%d", userId))
}
