package headless_cms

type Client interface {
	Token() string

	Request(page string, version string, language string) (cmsData map[string]any, err error)
	RequestJSON(page string, version string, language string) (jsonResp []byte, err error)
	RequestTranslatableTexts(page string, version string, language string) (texts map[string]string, err error)

	EmptyCache(empty_cache_token string) error
}
