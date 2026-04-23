package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations.
// This abstraction allows for easy swapping of cache providers.
type Cache interface {
	// Get retrieves a value from the cache by key.
	// It returns the value and a boolean indicating if the key was found.
	Get(ctx context.Context, key string) (interface{}, bool)

	// Set stores a value in the cache with a specified TTL (Time-To-Live).
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)

	// Delete removes a value from the cache by key.
	Delete(ctx context.Context, key string)
}
