// ZCache.zcache.cache.go
package zcache

import (
	"ZCache/lru"
	"sync"
)

/* 对LRU的封装，支持并发安全 */
type cache struct {
	mu       sync.Mutex // 锁
	lru      *lru.Cache // 封装lru结构
	maxBytes int64      // 用于设置缓存最大内存容量(单位：字节)
}

/* 对添加和更新操作上锁 */
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// 延迟初始化方法(Lazy Initialization)
		c.lru = lru.New(c.maxBytes, nil)
	}
	c.lru.Add(key, value)
}

/* 对获取操作上锁 */
func (c *cache) get(key string) (b ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}
