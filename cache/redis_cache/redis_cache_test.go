package redis_cache

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestCache_Get(t *testing.T) {
	redisAddrs := []string{os.Getenv("REDIS_ADDR")}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisMasterName := os.Getenv("REDIS_MASTER_NAME")
	redisDB := 0

	ctx := context.Background()
	c := New(redis.NewUniversalClient(&redis.UniversalOptions{Addrs: redisAddrs, Password: redisPassword, DB: redisDB, MasterName: redisMasterName}))

	// Set key-value pairs in the cache
	err := c.Set(ctx, "1", []byte("1"))
	if err != nil {
		t.Error(err)
	}
	err = c.Set(ctx, "2", []byte("zwei"))
	if err != nil {
		t.Error(err)
	}

	// Test getting existing keys
	a1, err := c.Get(ctx, "1")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a1, []byte("1")) {
		t.Error("should equal")
	}
	a2, err := c.Get(ctx, "2")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a2, []byte("zwei")) {
		t.Error("should equal")
	}

	// Test getting non-existent key
	a3, err := c.Get(ctx, "not found")
	if err == nil || a3 != nil {
		t.Error("error:", err, "a3", a3)
	}

	// Test deleting a key
	err = c.Del(ctx, "2")
	if err != nil {
		t.Error(err)
	}
	a2, err = c.Get(ctx, "2")
	if err == nil || a2 != nil {
		t.Error(err)
	}

	// Test emptying the cache
	err = c.Empty(ctx)
	if err != nil {
		t.Error(err)
	}
	a1, err = c.Get(ctx, "1")
	if err == nil || a1 != nil {
		t.Error(err)
	}
}
