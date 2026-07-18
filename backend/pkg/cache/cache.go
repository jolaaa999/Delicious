package cache

import (
	"sync"
	"time"
)

// MemoryCache 通用内存缓存，支持 TTL 过期。
// 适用于 Vercel serverless 环境（单实例内有效，冷启动后重建）。
// 不再使用时需调用 Close() 停止后台协程。
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
	done  chan struct{}
}

type entry struct {
	value     string
	expiresAt time.Time
}

// New 创建缓存实例，并启动后台清理协程。
func New(ttl time.Duration) *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]entry),
		ttl:   ttl,
		done:  make(chan struct{}),
	}
	go c.cleanupLoop(5 * time.Minute)
	return c
}

// Close 停止后台清理协程。调用后缓存不可再使用。
func (c *MemoryCache) Close() {
	close(c.done)
}

// Get 获取缓存值，过期返回空字符串。
func (c *MemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.value, true
}

// Set 写入缓存。
func (c *MemoryCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}

// Delete 删除缓存键。
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len 返回当前缓存条目数。
func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// cleanupLoop 定期清理过期条目，done 关闭时退出。
func (c *MemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.items {
		if now.After(v.expiresAt) {
			delete(c.items, k)
		}
	}
}
