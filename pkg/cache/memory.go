package cache

import (
	"errors"
	"sync"
	"time"
)

type item struct {
	value     interface{}
	createdAt int64
	ttl       int64
}

type InMemoryCache struct {
	mu         sync.RWMutex
	defaultTtl int64
	items      map[interface{}]*item
}

var (
	ErrItemNotFound = errors.New("item wasn't found")
	ErrItemExpired  = errors.New("time to live of item was over")
)

func NewMemoryCache(defaultTtl int64) *InMemoryCache {
	c := &InMemoryCache{
		items:      make(map[interface{}]*item),
		defaultTtl: defaultTtl,
	}
	go c.setTtlTimer()

	return c
}

func (c *InMemoryCache) setTtlTimer() {
	for {
		c.mu.Lock()
		for k, v := range c.items {
			if time.Now().Unix()-v.createdAt > v.ttl {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()

		<-time.After(time.Second)
	}
}

func (c *InMemoryCache) Set(key, value interface{}, ttl int64) {
	if ttl == 0 {
		ttl = c.defaultTtl
	}

	c.mu.Lock()
	c.items[key] = &item{
		value:     value,
		createdAt: time.Now().Unix(),
		ttl:       ttl,
	}
	c.mu.Unlock()
}

func (c *InMemoryCache) Get(key interface{}) (interface{}, error) {
	c.mu.RLock()
	i, ex := c.items[key]
	c.mu.RUnlock()

	if !ex {
		return nil, ErrItemNotFound
	}

	if time.Now().Unix()-i.createdAt > i.ttl {
		return nil, ErrItemExpired
	}

	return i.value, nil
}
