package memory_cache

import (
	"sync"
)

type Cache struct {
	mp   map[string]interface{}
	lock sync.RWMutex
}

func New() *Cache {
	return &Cache{mp: make(map[string]interface{})}
}

func (mc *Cache) Get(key string) (interface{}, error) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	value, ok := mc.mp[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (mc *Cache) Set(key string, obj interface{}) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.mp[key] = obj
	return nil
}

func (mc *Cache) Del(key string) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.mp, key)
	return nil
}

func (mc *Cache) Empty() error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.mp = make(map[string]interface{})
	return nil
}
