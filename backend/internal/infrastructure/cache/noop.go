package cache

import (
	"anirank/api/internal/domain"
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

type NoOpSubscriber struct {
	ch chan domain.PubSubMessage
}

func (n *NoOpSubscriber) Channel() <-chan domain.PubSubMessage {
	return n.ch
}

func (n *NoOpSubscriber) Close() error {
	return nil
}

func (c *NoOpCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return nil
}

func (c *NoOpCache) Subscribe(ctx context.Context, channel string) (domain.Subscriber, error) {
	sub := &NoOpSubscriber{
		ch: make(chan domain.PubSubMessage),
	}
	close(sub.ch)
	return sub, nil
}
