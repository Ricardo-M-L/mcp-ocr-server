package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Value      interface{}
	ExpireTime time.Time
}

// Cache OCR 结果缓存
type Cache struct {
	store   map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
	enabled bool
}

// NewCache 创建缓存实例
func NewCache(maxSize int, ttl time.Duration, enabled bool) *Cache {
	cache := &Cache{
		store:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		enabled: enabled,
	}

	if enabled {
		// 启动清理协程
		go cache.cleanupRoutine()
	}

	return cache
}

// Get 获取缓存值
func (c *Cache) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.store[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.ExpireTime) {
		return nil, false
	}

	logger.Debug("Cache hit", zap.String("key", key))
	return entry.Value, true
}

// Set 设置缓存值
func (c *Cache) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存是否已满
	if len(c.store) >= c.maxSize {
		c.evictOldest()
	}

	c.store[key] = &CacheEntry{
		Value:      value,
		ExpireTime: time.Now().Add(c.ttl),
	}

	logger.Debug("Cache set", zap.String("key", key))
}

// Delete 删除缓存值
func (c *Cache) Delete(key string) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
	logger.Debug("Cache deleted", zap.String("key", key))
}

// Clear 清空缓存
func (c *Cache) Clear() {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*CacheEntry)
	logger.Info("Cache cleared")
}

// Size 获取缓存大小
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// Stats 获取缓存统计信息
func (c *Cache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"size":     len(c.store),
		"max_size": c.maxSize,
		"ttl_seconds": c.ttl.Seconds(),
		"enabled": c.enabled,
	}
}

// evictOldest 淘汰最旧的条目
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.store {
		if oldestKey == "" || entry.ExpireTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpireTime
		}
	}

	if oldestKey != "" {
		delete(c.store, oldestKey)
		logger.Debug("Cache evicted oldest", zap.String("key", oldestKey))
	}
}

// cleanupRoutine 定期清理过期条目
func (c *Cache) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期条目
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range c.store {
		if now.After(entry.ExpireTime) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.store, key)
	}

	if len(expiredKeys) > 0 {
		logger.Debug("Cache cleanup", zap.Int("expired_count", len(expiredKeys)))
	}
}

// GenerateKey 生成缓存键 (基于图像数据的 SHA256 哈希)
func GenerateKey(data []byte, options ...string) string {
	hasher := sha256.New()
	hasher.Write(data)

	// 添加选项到哈希
	for _, opt := range options {
		hasher.Write([]byte(opt))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}