package memory

func New() *Cache {
	return &Cache{mp: map[string]any{}}
}

type Cache struct {
	mp map[string]any
}

func (mc *Cache) Get(key string) (any, error) {
	return mc.mp[key], nil
}

func (mc *Cache) Set(key string, obj any) error {
	mc.mp[key] = obj
	return nil
}

func (mc *Cache) Del(key string) error {
	delete(mc.mp, key)
	return nil
}

func (mc *Cache) Empty() error {
	mc.mp = map[string]any{}
	return nil
}
