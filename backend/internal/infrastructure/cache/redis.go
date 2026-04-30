package cache

import (
	"anirank/api/internal/domain"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client  *redis.Client
	healthy bool
}

func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	opts.DialTimeout = 1 * time.Second
	opts.ReadTimeout = 1 * time.Second
	opts.WriteTimeout = 1 * time.Second
	client := redis.NewClient(opts)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return &RedisCache{client: client, healthy: false}, nil // Return unhealthy instead of error
	}

	return &RedisCache{client: client, healthy: true}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		c.healthy = false
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = c.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		c.healthy = false
	}
	return err
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) IsAvailable() bool {
	if !c.healthy {
		// Attempt to recover if it was marked unhealthy
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		if err := c.client.Ping(ctx).Err(); err == nil {
			c.healthy = true
		}
	}
	return c.healthy
}

type RedisSubscriber struct {
	pubsub *redis.PubSub
	ch     chan domain.PubSubMessage
}

func (r *RedisSubscriber) Channel() <-chan domain.PubSubMessage {
	return r.ch
}

func (r *RedisSubscriber) Close() error {
	return r.pubsub.Close()
}

func (c *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.client.Publish(ctx, channel, data).Err()
}

func (c *RedisCache) Subscribe(ctx context.Context, channel string) (domain.Subscriber, error) {
	ps := c.client.Subscribe(ctx, channel)
	sub := &RedisSubscriber{
		pubsub: ps,
		ch:     make(chan domain.PubSubMessage),
	}

	go func() {
		defer close(sub.ch)
		for msg := range ps.Channel() {
			sub.ch <- domain.PubSubMessage{
				Channel: msg.Channel,
				Payload: msg.Payload,
			}
		}
	}()

	return sub, nil
}
