package redis_cache

import (
	"context"
	"errors"

	"github.com/dryaf/headless_cms"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client redis.UniversalClient
}

func New(client redis.UniversalClient) headless_cms.Cache {
	return &Cache{client: client}
}

func (mc *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := mc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.New("key not found")
	} else if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func (mc *Cache) Set(ctx context.Context, key string, bytes []byte) error {
	err := mc.client.Set(ctx, key, bytes, 0).Err()
	return err
}

func (mc *Cache) Del(ctx context.Context, key string) error {
	err := mc.client.Del(ctx, key).Err()
	return err
}

func (mc *Cache) Empty(ctx context.Context) error {
	keys, err := mc.client.Keys(ctx, "*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		err = mc.client.Del(ctx, keys...).Err()
	}
	return err
}
