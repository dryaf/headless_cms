package storyblok

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/dryaf/headless_cms"
	"github.com/dryaf/headless_cms/client/storyblok/models"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	HttpClient HTTPClient

	empty_cache_token string

	cache headless_cms.Cache

	token   string
	api_url string

	default_version          string // "draft", "published"
	ignore_cache_for_version string
}

// /https://github.com/storyblok/storyblok-ruby
func NewClient(token string, empty_cache_token string, cache headless_cms.Cache, httpClient HTTPClient) *Client {
	if token == "" || empty_cache_token == "" || cache == nil {
		panic(" NewClient(token string, empty_cache_token string, cache cache.Cache) *Client params need to be set")
	}
	return &Client{
		cache:                    cache,
		empty_cache_token:        empty_cache_token,
		ignore_cache_for_version: "draft",

		HttpClient:      httpClient,
		api_url:         "https://api.storyblok.com/v2/cdn/stories",
		token:           token,
		default_version: "published",
	}
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) Cache() headless_cms.Cache {
	return c.cache
}

func (c *Client) EmptyCache(user_input_token string) error {
	if user_input_token != c.empty_cache_token {
		return errors.New("token incorrect")
	}
	return c.cache.Empty()
}

// RequestJSON story for example /login or "" for getting all stories
func (c *Client) RequestJSON(page string, version string, language string) ([]byte, error) {
	cacheKey := c.CacheKey("j", page, version, language)

	// Cache read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, cacheErr := c.cache.Get(cacheKey)
		if cacheErr != nil {
			slog.Warn("storyblok", "cache.Get error", "url_params", cacheKey, "err", cacheErr)
		}
		if obj != nil {
			jsonResp, ok := obj.([]byte)
			if ok {
				return jsonResp, nil
			}
			slog.Warn("storyblok", "cache object not []byte type", "url_params", cacheKey, "obj", obj)
		}
	}

	// Remote CMS
	reqURL := c.api_url + c.URLParams(page, version, language) + "&token=" + c.token
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: %w", reqURL, err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: resp: %w", reqURL, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("headless_cms: %s: status: %d err: %w", reqURL, resp.StatusCode, errors.New(storyblokStatusDescriptions[resp.StatusCode]))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: readBody: %w", reqURL, err)
	}

	// Cache - Write
	if c.cache != nil && version != c.ignore_cache_for_version {
		err = c.cache.Set(cacheKey, body)
	}
	return body, err
}

func (c *Client) RequestTranslatableTexts(page string, version string, language string) (map[string]string, error) {
	cacheKey := c.CacheKey("t", page, version, language)

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, cacheErr := c.cache.Get(cacheKey)
		if cacheErr != nil {
			slog.Warn("storyblok", "cache.Get error", "url_params", cacheKey, "err", cacheErr)
		}
		texts, ok := obj.(map[string]string)
		if ok {
			return texts, nil
		}
		slog.Warn("storyblok", "cache object not map[string]string type", "url_params", cacheKey, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", cacheKey, err)
	}
	storyblokPage := models.StoryWithTranslatableTextsOnly{}
	err = json.Unmarshal(jsonResp, &storyblokPage)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", cacheKey, err)
	}
	texts := map[string]string{}
	for _, tt := range storyblokPage.Story.Content.Body {
		if tt.Component == "_translatable_text" {
			texts[tt.ID] = tt.Value
			texts[tt.ID+"_editable"] = tt.Editable
		}
	}

	// Cache - Write
	if c.cache != nil && version != c.ignore_cache_for_version {
		err = c.cache.Set(cacheKey, texts)
	}
	return texts, err
}

func (c *Client) Request(page string, version string, language string) (map[string]interface{}, error) {
	cacheKey := c.CacheKey("r", page, version, language)

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, cacheErr := c.cache.Get(cacheKey)
		if cacheErr != nil {
			slog.Warn("storyblok", "cache.Get error", "url_params", cacheKey, "err", cacheErr)
		}
		cmsData, ok := obj.(map[string]interface{})
		if ok {
			return cmsData, nil
		}
		slog.Warn("storyblok", "cache object not map[string]interface{}", "url_params", cacheKey, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", cacheKey, err)
	}
	cmsData := map[string]interface{}{}
	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", cacheKey, err)
	}

	// Cache - Write
	if c.cache != nil && version != c.ignore_cache_for_version {
		err = c.cache.Set(cacheKey, cmsData)
	}
	return cmsData, err
}

func (c *Client) RequestSimpleBlocksWithID(page string, version string, language string) (map[string]interface{}, error) {
	cacheKey := c.CacheKey("i", page, version, language)

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, cacheErr := c.cache.Get(cacheKey)
		if cacheErr != nil {
			slog.Warn("storyblok", "cache.Get error", "url_params", cacheKey, "err", cacheErr)
		}
		cmsData, ok := obj.(map[string]interface{})
		if ok {
			return cmsData, nil
		}
		slog.Warn("storyblok", "cache object not map[string]any", "url_params", cacheKey, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", cacheKey, err)
	}
	cmsData := models.SimpleBlockskWithID{}
	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", cacheKey, err)
	}
	resp := map[string]interface{}{}
	for _, tt := range cmsData.Story.Content.Body {
		id, ok := tt["id"].(string)
		if ok && len(id) > 0 {
			resp[id] = tt
		}
	}

	// Cache - Write
	if c.cache != nil && version != c.ignore_cache_for_version {
		err = c.cache.Set(cacheKey, resp)
	}
	return resp, err
}

func (c *Client) CacheKey(prefix, page, version, language string) string {
	return fmt.Sprint(prefix, ":", version, ":", language, ":", page)
}

func (c *Client) URLParams(page, version, language string) string {
	if page != "" {
		page = "/" + page
	}
	if version == "" {
		version = c.default_version
	}
	url_params := page + "?version=" + version
	if language != "" {
		url_params = url_params + "&language=" + language
	}
	return url_params
}

// https://www.storyblok.com/docs/api/content-delivery
var storyblokStatusDescriptions = map[int]string{
	200: "OK Everything worked as expected.",
	400: "Bad Request Wrong format was sent (eg. XML instead of JSON).",
	401: "Unauthorized No valid API key provided.",
	404: "Not Found	The requested resource doesn't exist (perhaps due to not yet published content entries).",
	422: "Unprocessable Entity The request was unacceptable, often due to missing a required parameter.",
	429: "Too Many Requests	Too many requests hit the API too quickly. We recommend an exponential backoff of your requests.",
	500: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	502: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	503: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
	504: "Server Errors	Something went wrong on Storyblok's end. (These are rare.)",
}
