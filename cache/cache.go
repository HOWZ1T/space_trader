// Provides a caching mechanism to reduce the amount of API calls.
package cache

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	data        interface{}
	lastUpdated time.Time
}

// Caches objects to reduce the number of API calls.
type Cache struct {
	mu           sync.Mutex
	expiresAfter time.Duration
	store        map[string]cacheEntry
}

func New(expiresAfter time.Duration) Cache {
	return Cache{
		expiresAfter: expiresAfter,
		store:        make(map[string]cacheEntry),
	}
}

func processKey(key string) string {
	return strings.ToLower(strings.Trim(key, "\n\r "))
}

func (c *Cache) Store(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key = processKey(key)
	c.store[key] = cacheEntry{
		data:        data,
		lastUpdated: time.Now(),
	}
}

func (c *Cache) Fetch(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	key = processKey(key)
	if v, ok := c.store[key]; ok {
		if os.Getenv("ST_LOG") == "verbose" {
			fmt.Println(key + " -> cache hit")
		}
		fmt.Println(key + " -> cache hit")

		return v.data
	}
	return nil
}

func (c *Cache) IsOld(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	key = processKey(key)
	if v, ok := c.store[key]; ok {
		return time.Now().Sub(v.lastUpdated) >= c.expiresAfter
	}

	return false
}
