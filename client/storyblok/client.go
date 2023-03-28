package storyblok

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/dryaf/headless_cms/cache"
	"github.com/dryaf/headless_cms/client/storyblok/models"
)

type Client struct {
	httpClient *http.Client

	empty_cache_token string

	cache cache.Cache

	token   string
	api_url string

	default_version          string // "draft", "published"
	ignore_cache_for_version string
}

// /https://github.com/storyblok/storyblok-ruby
func NewClient(token string, empty_cache_token string, cache cache.Cache) *Client {
	if token == "" || empty_cache_token == "" || cache == nil {
		panic(" NewClient(token string, empty_cache_token string, cache cache.Cache) *Client params need to be set")
	}
	return &Client{
		cache:                    cache,
		empty_cache_token:        empty_cache_token,
		ignore_cache_for_version: "draft",

		httpClient:      &http.Client{},
		api_url:         "https://api.storyblok.com/v2/cdn/stories",
		token:           token,
		default_version: "published",
	}
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) EmptyCache(user_input_token string) error {
	if user_input_token != c.empty_cache_token {
		return errors.New("token incorrect")
	}
	return c.cache.Empty()
}

// RequestJSON story for example /login or "" for getting all stories
func (c *Client) RequestJSON(page string, version string, language string) ([]byte, error) {
	url_params := c.getKey(page, version, language) + "_bytes"

	// Cache - Read
	if version != c.ignore_cache_for_version {
		obj, err := c.cache.Get(url_params)
		if err != nil {
			return nil, nil
		}
		if obj != nil {
			jsonResp, ok := obj.([]byte)
			if ok {
				return jsonResp, nil
			}
			slog.Warn("storyblok", "cache object not []byte type", "url_params", url_params, "obj", obj)
		}
	}

	// Remote CMS
	reqURL := c.api_url + url_params + "&token=" + c.token
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: %w", reqURL, err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: resp: %w", reqURL, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("headless_cms: %s: status: %d err: %w", reqURL, resp.StatusCode, errors.New(statusCodes[resp.StatusCode]))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: readBody: %w", reqURL, err)
	}

	// Cache - Write
	if c.cache != nil {
		err = c.cache.Set(url_params, body)
	}
	return body, err
}

func (c *Client) RequestTranslatableTexts(page string, version string, language string) (map[string]string, error) {
	url_params := c.getKey(page, version, language) + "_texts"

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, err := c.cache.Get(url_params)
		if err != nil {
			return nil, err
		}
		texts, ok := obj.(map[string]string)
		if ok {
			return texts, nil
		}
		slog.Warn("storyblok", "cache object not map[string]stringtype", "url_params", url_params, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", url_params, err)
	}
	storyblokPage := models.StoryWithTranslatableTextsOnly{}
	err = json.Unmarshal(jsonResp, &storyblokPage)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", url_params, err)
	}
	texts := map[string]string{}
	for _, tt := range storyblokPage.Story.Content.Body {
		if tt.Component == "_translatable_text" {
			texts[tt.ID] = tt.Value
			texts[tt.ID+"_editable"] = tt.Editable
		}
	}

	// Cache - Write
	if c.cache != nil {
		err = c.cache.Set(url_params, texts)
	}
	return texts, err
}

func (c *Client) Request(page string, version string, language string) (map[string]any, error) {
	url_params := c.getKey(page, version, language) + "_blk"

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, err := c.cache.Get(url_params)
		if err != nil {
			return nil, err
		}
		cmsData, ok := obj.(map[string]any)
		if ok {
			return cmsData, nil
		}
		slog.Warn("storyblok", "cache object not map[string]any", "url_params", url_params, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", url_params, err)
	}
	cmsData := map[string]any{}
	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", url_params, err)
	}

	// Cache - Write
	if c.cache != nil {
		err = c.cache.Set(url_params, cmsData)
	}
	return cmsData, err
}

func (c *Client) RequestSimpleBlocksWithID(page string, version string, language string) (map[string]any, error) {
	url_params := c.getKey(page, version, language) + "_blk"

	// Cache - Read
	if c.cache != nil && version != c.ignore_cache_for_version {
		obj, err := c.cache.Get(url_params)
		if err != nil {
			return nil, err
		}
		cmsData, ok := obj.(map[string]any)
		if ok {
			return cmsData, nil
		}
		slog.Warn("storyblok", "cache object not map[string]any", "url_params", url_params, "obj", obj)
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", url_params, err)
	}
	cmsData := models.SimpleBlocskWithID{}
	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", url_params, err)
	}
	resp := map[string]any{}
	for _, tt := range cmsData.Story.Content.Body {
		id, ok := tt["id"].(string)
		if ok && len(id) > 0 {
			resp[id] = tt
		}
	}

	// Cache - Write
	if c.cache != nil {
		err = c.cache.Set(url_params, cmsData)
	}
	return resp, err
}

func (c *Client) getKey(page, version, language string) string {
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
