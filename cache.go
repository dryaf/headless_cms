package headless_cms

type Cache interface {
	Get(key string) (any, error)
	Set(key string, obj any) error
	Del(key string) error
	Empty() error
}
