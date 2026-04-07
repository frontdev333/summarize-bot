package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time
}

type SummaryCache struct {
	mtx     sync.RWMutex
	entries map[string]entry
	ttl     time.Duration
}

func NewSummaryCache(ttl time.Duration) *SummaryCache {
	return &SummaryCache{
		mtx:     sync.RWMutex{},
		entries: make(map[string]entry),
		ttl:     ttl,
	}
}

func (c *SummaryCache) Get(key string) (string, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}

	return e.value, true
}

func (c *SummaryCache) Set(key, value string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.entries[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(5 * c.ttl),
	}
}
