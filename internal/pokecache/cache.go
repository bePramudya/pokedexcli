package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries  map[string]cacheEntry
	mu       *sync.RWMutex
	interval time.Duration
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return []byte{}, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if entry.createdAt.Add(interval).Before(time.Now()) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

func NewCache(interval time.Duration) Cache {
	caches := Cache{
		entries:  map[string]cacheEntry{},
		mu:       &sync.RWMutex{},
		interval: interval,
	}
	go caches.reapLoop(interval)

	return caches
}
