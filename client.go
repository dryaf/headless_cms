package headless_cms

import "context"

type Client interface {
	AuthToken() string

	GetPage(ctx context.Context, pageSlug string, version string, language string) (map[string]any, error)
	GetPageAsJSON(ctx context.Context, pageSlug string, version string, language string) ([]byte, error)
	GetPageAsSimpleBlocksWithID(ctx context.Context, pageSlug string, version string, language string) (map[string]map[string]any, error)

	Cache() Cache
	CacheKey(prefix, page, version, language string) string

	EmptyCache(ctx context.Context, token string) error
	EmptyCacheToken(ctx context.Context) (string, error)
}
