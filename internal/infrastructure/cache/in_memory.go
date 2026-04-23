package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

type inMemoryCache struct {
	client *cache.Cache
}

// NewInMemoryCache creates a new instance of an in-memory cache.
// defaultExpiration: how long items stay in the cache by default.
// cleanupInterval: how often the cache should be purged of expired items.
func NewInMemoryCache(defaultExpiration, cleanupInterval time.Duration) Cache {
	return &inMemoryCache{
		client: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *inMemoryCache) Get(_ context.Context, key string) (interface{}, bool) {
	return c.client.Get(key)
}

func (c *inMemoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) {
	c.client.Set(key, value, ttl)
}

func (c *inMemoryCache) Delete(_ context.Context, key string) {
	c.client.Delete(key)
}
