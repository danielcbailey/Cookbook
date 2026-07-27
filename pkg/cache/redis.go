package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(connStr string) (*RedisCache, error) {
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, err
	}
	return &RedisCache{
		client: redis.NewClient(opts),
	}, nil
}

// Get returns a not-found error (detectable via ErrIsNotFound) for missing
// keys, per the Cache interface contract.
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) GetInterface(ctx context.Context, key string, ptr any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), ptr)
}

func (r *RedisCache) Set(ctx context.Context, key, value string, expiry time.Duration) error {
	return r.client.Set(ctx, key, value, expiry).Err()
}

func (r *RedisCache) SetInterface(ctx context.Context, key string, ptr any, expiry time.Duration) error {
	data, err := json.Marshal(ptr)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, string(data), expiry).Err()
}

func (r *RedisCache) Lock(ctx context.Context, key string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := r.client.SetNX(ctx, key+":lock", "1", timeout).Result()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("lock acquisition timed out")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (r *RedisCache) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key+":lock").Err()
}

func (r *RedisCache) Add(ctx context.Context, key string, delta int64, expiry time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.IncrBy(ctx, key, delta)
	if expiry > 0 {
		pipe.Expire(ctx, key, expiry)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (_ *RedisCache) ErrIsNotFound(err error) bool {
	return errors.Is(err, redis.Nil)
}
