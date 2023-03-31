package redis_cache

import (
	"context"
	"errors"

	"github.com/dryaf/headless_cms"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

func New(addr, password string, db int) headless_cms.Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx := context.Background()

	return &Cache{client: client, ctx: ctx}
}

func NewFailover(sentinelAddrs []string, masterName, password string, db int) headless_cms.Cache {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
		DB:            db,
	})
	ctx := context.Background()
	return &Cache{client: client, ctx: ctx}
}

func (mc *Cache) Get(key string) (interface{}, error) {
	value, err := mc.client.Get(mc.ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.New("key not found")
	} else if err != nil {
		return nil, err
	}
	return value, nil
}

func (mc *Cache) Set(key string, obj interface{}) error {
	err := mc.client.Set(mc.ctx, key, obj, 0).Err()
	return err
}

func (mc *Cache) Del(key string) error {
	err := mc.client.Del(mc.ctx, key).Err()
	return err
}

func (mc *Cache) Empty() error {
	keys, err := mc.client.Keys(mc.ctx, "*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		err = mc.client.Del(mc.ctx, keys...).Err()
	}
	return err
}
