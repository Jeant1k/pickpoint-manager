package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	ttl    time.Duration
	client *redis.Client
}

func New(ctx context.Context, url, password string, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	status := client.Ping(ctx)
	if status.Err() != nil {
		log.Fatalf("failed to connect to redis: %v", status.Err())
	}

	return &RedisCache{
		ttl:    ttl,
		client: client,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) bool {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false
		}
		log.Printf("failed to fetch key %s: %v", key, err)
		return false
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		log.Printf("failed to unmarshal key %s: %v", key, err)
		return false
	}

	return true
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	err = r.client.Set(ctx, key, b, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to write value to redis: %w", err)
	}
	return nil
}
