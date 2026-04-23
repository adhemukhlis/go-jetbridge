package cache

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type inMemoryCache[V any] struct {
	client *ttlcache.Cache[string, V]
}

// NewInMemoryCache creates a new instance of an in-memory cache.
// defaultExpiration: how long items stay in the cache by default.
func NewInMemoryCache[V any](defaultExpiration time.Duration) Cache[V] {
	client := ttlcache.New[string, V](
		ttlcache.WithTTL[string, V](defaultExpiration),
	)

	// Start the background worker for automatic cache eviction
	go client.Start()

	return &inMemoryCache[V]{
		client: client,
	}
}

func (c *inMemoryCache[V]) Get(_ context.Context, key string) (V, bool) {
	item := c.client.Get(key)
	if item != nil && !item.IsExpired() {
		return item.Value(), true
	}
	var zero V
	return zero, false
}

func (c *inMemoryCache[V]) Set(_ context.Context, key string, value V, ttl time.Duration) {
	c.client.Set(key, value, ttl)
}

func (c *inMemoryCache[V]) Delete(_ context.Context, key string) {
	c.client.Delete(key)
}
