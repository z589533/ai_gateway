package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type CachedKey struct {
	TenantID     uint64     `json:"tenant_id"`
	TenantStatus int8       `json:"tenant_status"`
	APIKeyID     uint64     `json:"api_key_id"`
	Status       int8       `json:"status"`
	Scopes       []string   `json:"scopes"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type RedisKeyCache struct {
	client *redis.Client
}

func NewRedisKeyCache(client *redis.Client) *RedisKeyCache {
	return &RedisKeyCache{client: client}
}

func (c *RedisKeyCache) Get(ctx context.Context, hash string) (*CachedKey, error) {
	value, err := c.client.Get(ctx, cacheKey(hash)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	var key CachedKey
	if err := json.Unmarshal([]byte(value), &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *RedisKeyCache) Set(ctx context.Context, hash string, key CachedKey, ttl time.Duration) error {
	data, err := json.Marshal(key)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cacheKey(hash), data, ttl).Err()
}

func (c *RedisKeyCache) Del(ctx context.Context, hash string) error {
	return c.client.Del(ctx, cacheKey(hash)).Err()
}

func cacheKey(hash string) string {
	return "ag:key:" + hash
}
