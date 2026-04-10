package cache

import (
	"anirank/api/internal/domain"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) IsAvailable() bool {
	return true
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
