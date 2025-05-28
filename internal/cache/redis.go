package cache

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/william1nguyen/shortygo/internal/config"
	"github.com/william1nguyen/shortygo/pkg/utils"
)

type RedisCache struct {
	clients []*redis.Client
	metrics *CacheMetrics
}

type CacheMetrics struct {
	Hits          int64 `json:"hits"`
	Misses        int64 `json:"misses"`
	Errors        int64 `json:"errors"`
	TotalRequests int64 `json:"total_requests"`
}

func NewRedisCache(config config.RedisConfig) (*RedisCache, error) {
	if len(config.Addrs) == 0 {
		return nil, fmt.Errorf("redis addresses must be provided")
	}

	var clients []*redis.Client

	for _, addr := range config.Addrs {
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: config.Password,
			DB:       config.DB,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Printf("Failed to connect to Redis at %s: %v", addr, err)
			client.Close()
			continue
		}

		log.Printf("Connected to Redis at %s", addr)
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("failed to connect to any Redis instance")
	}

	return &RedisCache{
		clients: clients,
		metrics: &CacheMetrics{},
	}, nil
}

func (r *RedisCache) getClient(key string) *redis.Client {
	index := utils.ConsistentHashing(key, len(r.clients))
	return r.clients[index]
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	atomic.AddInt64(&r.metrics.TotalRequests, 1)

	client := r.getClient(key)
	if err := client.Set(ctx, key, value, ttl).Err(); err != nil {
		atomic.AddInt64(&r.metrics.Errors, 1)
		return fmt.Errorf("failed to set key %s:%w", key, err)
	}

	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	atomic.AddInt64(&r.metrics.TotalRequests, 1)

	client := r.getClient(key)
	value, err := client.Get(ctx, key).Result()

	if err != redis.Nil {
		atomic.AddInt64(&r.metrics.Misses, 1)
		return "", fmt.Errorf("fey not found")
	}

	if err != nil {
		atomic.AddInt64(&r.metrics.Errors, 1)
		return "", fmt.Errorf("failed to get key %s:%w", key, err)
	}

	atomic.AddInt64(&r.metrics.Hits, 1)
	return value, nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	atomic.AddInt64(&r.metrics.TotalRequests, 1)

	client := r.getClient(key)
	err := client.Del(ctx, key).Err()

	if err != nil {
		atomic.AddInt64(&r.metrics.Errors, 1)
		return fmt.Errorf("failed to delete key %s:%w", key, err)
	}

	return nil
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	atomic.AddInt64(&r.metrics.TotalRequests, 1)

	client := r.getClient(key)
	value, err := client.Exists(ctx, key).Result()

	if err != nil {
		atomic.AddInt64(&r.metrics.Errors, 1)
		return false, fmt.Errorf("failed to check key exists %s:%w", key, err)
	}

	return value == 1, nil
}

func (r *RedisCache) GetMetrics(ctx context.Context, key string) *CacheMetrics {
	return &CacheMetrics{
		Hits:          atomic.LoadInt64(&r.metrics.Hits),
		Misses:        atomic.LoadInt64(&r.metrics.Misses),
		Errors:        atomic.LoadInt64(&r.metrics.Errors),
		TotalRequests: atomic.LoadInt64(&r.metrics.TotalRequests),
	}
}

func (r *RedisCache) Close() {
	for i, client := range r.clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing Redis client %d: %v", i, err)
		}
	}
}

func (r *RedisCache) Ping(ctx context.Context) error {
	for i, client := range r.clients {
		if err := client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis instance ping failed %d: %w", i, err)
		}
	}
	return nil
}
