package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")

type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

func (c *NoOpCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) IsAvailable() bool {
	return false
}
