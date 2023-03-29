package headless_cms

type Client interface {
	EmptyCache(user_input_token string) error

	Request(page string, version string, language string) (map[string]any, error)
	RequestJSON(page string, version string, language string) ([]byte, error)
	RequestTranslatableTexts(page string, version string, language string) (map[string]string, error)
	RequestSimpleBlocksWithID(page string, version string, language string) (map[string]any, error)
}
