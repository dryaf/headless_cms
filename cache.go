package headless_cms

import (
	"context"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, bytes []byte) error
	Del(ctx context.Context, key string) error
	Empty(ctx context.Context) error
}
