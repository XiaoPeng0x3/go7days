package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	// 互斥锁
	mu sync.Mutex
	// cache
	lru *lru.Cache
	// 最大字节数
	cacheBytes int64
}

// 向cache里面新增元素
func (c *cache) add(key string, value ByteView) {
	// lock
	c.mu.Lock()
	// unlock
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value) // ByteView实现了Len接口，所以可以传递这个类型的参数
}

func (c *cache) get(key string)(bv ByteView, ok bool) {
	// lock
	c.mu.Lock()
	// unlock
	defer c.mu.Unlock()
	if value, ok := c.lru.Get(key); ok {
		return value.(ByteView), ok
	}
	return
}