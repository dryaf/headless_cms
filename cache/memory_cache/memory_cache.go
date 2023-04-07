package memory_cache

import (
	"context"
	"errors"
	"sync"
)

type Cache struct {
	mp   map[string]interface{}
	lock sync.RWMutex
}

func New() *Cache {
	return &Cache{mp: make(map[string]interface{})}
}

func (mc *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	value, ok := mc.mp[key]
	if !ok {
		return nil, nil
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		return nil, errors.New("value is not []byte")
	}
	return valueBytes, nil
}

func (mc *Cache) Set(ctx context.Context, key string, obj []byte) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.mp[key] = obj
	return nil
}

func (mc *Cache) Del(ctx context.Context, key string) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.mp, key)
	return nil
}

func (mc *Cache) Empty(ctx context.Context) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.mp = make(map[string]interface{})
	return nil
}
