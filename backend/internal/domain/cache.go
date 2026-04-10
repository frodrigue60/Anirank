package domain

import (
	"context"
	"time"
)

type PubSubMessage struct {
	Channel string
	Payload string
}

type Subscriber interface {
	Channel() <-chan PubSubMessage
	Close() error
}

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	IsAvailable() bool
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (Subscriber, error)
}
