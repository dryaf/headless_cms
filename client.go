package headless_cms

import "context"

type Client interface {
	EmptyCache(ctx context.Context, user_input_token string) error

	Request(ctx context.Context, page string, version string, language string) (map[string]any, error)
	RequestJSON(ctx context.Context, page string, version string, language string) ([]byte, error)
	RequestTranslatableTexts(ctx context.Context, page string, version string, language string) (map[string]string, error)
	RequestSimpleBlocksWithID(ctx context.Context, page string, version string, language string) (map[string]any, error)
}
