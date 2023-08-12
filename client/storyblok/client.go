package storyblok

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/dryaf/headless_cms"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ headless_cms.Client = &Client{}

type Client struct {
	HttpClient HTTPClient

	cache                 headless_cms.Cache
	cacheEmptyActionToken string

	cmsAuthToken string
	cmsAPIUrl    string

	versionDefault           string // "published"
	versionWhereCacheIgnored string // "draft"
}

func NewClient(ctx context.Context, token string, emptyCacheToken string, cache headless_cms.Cache, httpClient HTTPClient) *Client {
	if token == "" {
		slog.ErrorContext(ctx, "storyblok - AuthToken is empty")
		os.Exit(1)
	}
	if emptyCacheToken == "" {
		slog.ErrorContext(ctx, "storyblok - EmptyCacheToken is empty")
		os.Exit(1)
	}
	if cache == nil {
		slog.ErrorContext(ctx, "storyblok - Cache is nil")
		os.Exit(1)
	}
	return &Client{
		cache:                    cache,
		cacheEmptyActionToken:    emptyCacheToken,
		versionWhereCacheIgnored: "draft",

		HttpClient:     httpClient,
		cmsAPIUrl:      "https://api.storyblok.com/v2/cdn/stories",
		cmsAuthToken:   token,
		versionDefault: "published",
	}
}

func (c *Client) AuthToken() string {
	return c.cmsAuthToken
}

func (c *Client) Cache() headless_cms.Cache {
	return c.cache
}

func (c *Client) EmptyCache(ctx context.Context, token string) error {
	if token != c.cacheEmptyActionToken {
		return errors.New("token incorrect")
	}
	return c.cache.Empty(ctx)
}

func (c *Client) EmptyCacheToken(ctx context.Context) (string, error) {
	if c.cacheEmptyActionToken == "" {
		return "", errors.New("token not set")
	}
	return c.cacheEmptyActionToken, nil
}

// GetPageAsJSON story for example /login or "" for getting all stories
func (c *Client) GetPageAsJSON(ctx context.Context, page string, version string, language string) ([]byte, error) {
	cacheKey := c.CacheKey("j", page, version, language)

	// Cache read
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		obj, cacheErr := c.cache.Get(ctx, cacheKey)
		if cacheErr != nil || obj == nil {
			slog.WarnContext(ctx, "storyblok - cache.Get",
				slog.String("url_params", cacheKey),
				slog.Any("err", cacheErr),
				slog.Bool("not_found", obj == nil))
		} else {
			return obj, nil
		}
	}

	// Remote CMS
	reqURL := c.cmsAPIUrl + c.cmsURLParams(page, version, language) + "&token=" + c.cmsAuthToken
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("headless_cms: %s: status: %d err: %w", reqURL, resp.StatusCode, errors.New(storyblokStatusDescriptions[resp.StatusCode]))
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: readBody: %w", reqURL, err)
	}

	// Cache - Write
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		cacheErr := c.cache.Set(ctx, cacheKey, body)
		if cacheErr != nil {
			slog.WarnContext(ctx, "storyblok - cache.Set error", slog.String("url_params", cacheKey), slog.Any("err", cacheErr))
		}
	}
	return body, nil
}

func (c *Client) GetPage(ctx context.Context, page string, version string, language string) (map[string]any, error) {
	cacheKey := c.CacheKey("r", page, version, language)
	cmsData := map[string]any{}

	// Cache - Read
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		obj, cacheErr := c.cache.Get(ctx, cacheKey)
		if cacheErr != nil || obj == nil {
			slog.WarnContext(ctx, "storyblok - cache.Get",
				slog.String("url_params", cacheKey),
				slog.Any("err", cacheErr),
				slog.Bool("not_found", obj == nil))
		} else {
			err := json.Unmarshal(obj, &cmsData)
			if err != nil {
				slog.ErrorContext(ctx, "storyblok - cache object not map[string]interface{}",
					slog.String("url_params", cacheKey), slog.Any("obj", obj))
			} else {
				return cmsData, nil
			}
		}
	}

	jsonResp, err := c.GetPageAsJSON(ctx, page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", cacheKey, err)
	}

	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", cacheKey, err)
	}

	// Cache - Write
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		jsonData, err := json.Marshal(cmsData)
		if err != nil {
			slog.ErrorContext(ctx, "storyblok - json marshal error", slog.Any("err", err), slog.String("key", cacheKey), slog.Any("data", cmsData))
		} else {
			err = c.cache.Set(ctx, cacheKey, jsonData)
			if err != nil {
				slog.ErrorContext(ctx, "storyblok - cache set error", slog.Any("err", err), slog.String("key", cacheKey), slog.Any("data", cmsData))
			}
		}
	}
	return cmsData, err
}

func (c *Client) GetPageAsSimpleBlocksWithID(ctx context.Context, page string, version string, language string) (map[string]map[string]any, error) {
	cacheKey := c.CacheKey("i", page, version, language)

	// Cache - Read
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		obj, cacheErr := c.cache.Get(ctx, cacheKey)
		if cacheErr != nil || obj == nil {
			slog.WarnContext(ctx, "storyblok - cache.Get",
				slog.String("url_params", cacheKey),
				slog.Any("err", cacheErr),
				slog.Bool("not_found", obj == nil))
		} else {
			resp := map[string]map[string]any{}
			err := json.Unmarshal(obj, &resp)
			if err != nil {
				slog.ErrorContext(ctx, "storyblok - cache object not map[string]map[string]any{}",
					slog.String("url_params", cacheKey), slog.Any("obj", obj))
			} else {
				return resp, nil
			}
		}
	}

	// Remote CMS
	cmsData := &SimpleBlockskWithID{
		Story: Story{
			Content:    Content{},
			TagList:    []interface{}{},
			Alternates: []interface{}{},
		},
		Rels:  []interface{}{},
		Links: []interface{}{},
	}

	jsonResp, err := c.GetPageAsJSON(ctx, page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", cacheKey, err)
	}

	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", cacheKey, err)
	}

	resp := map[string]map[string]any{}
	for _, tt := range cmsData.Story.Content.Body {
		id, ok := tt["id"].(string)
		if ok && len(id) > 0 {
			resp[id] = tt
		}
	}

	// Cache - Write
	if c.cache != nil && version != c.versionWhereCacheIgnored {
		jsonData, err := json.Marshal(resp)
		if err != nil {
			slog.ErrorContext(ctx, "storyblok - json marshal error", slog.Any("err", err), slog.String("cache_key", cacheKey), slog.Any("data", cmsData))
		} else {
			err = c.cache.Set(ctx, cacheKey, jsonData)
			if err != nil {
				slog.ErrorContext(ctx, "storyblok - cache set error", slog.Any("err", err), slog.String("cache_key", cacheKey), slog.Any("data", cmsData))
			}
		}
	}
	return resp, nil
}

func (c *Client) CacheKey(prefix, page, version, language string) string {
	return fmt.Sprint(prefix, ":", version, ":", language, ":", page)
}

func (c *Client) cmsURLParams(page, version, language string) string {
	if page != "" {
		page = "/" + page
	}
	if version == "" {
		version = c.versionDefault
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

type SimpleBlockskWithID struct {
	Story Story         `json:"story"`
	Cv    int           `json:"cv"`
	Rels  []interface{} `json:"rels"`
	Links []interface{} `json:"links"`
}

type Story struct {
	Name             string        `json:"name"`
	CreatedAt        time.Time     `json:"created_at"`
	PublishedAt      time.Time     `json:"published_at"`
	ID               int           `json:"id"`
	UUID             string        `json:"uuid"`
	Content          Content       `json:"content"`
	Slug             string        `json:"slug"`
	FullSlug         string        `json:"full_slug"`
	SortByDate       interface{}   `json:"sort_by_date"`
	Position         int           `json:"position"`
	TagList          []interface{} `json:"tag_list"`
	IsStartpage      bool          `json:"is_startpage"`
	ParentID         interface{}   `json:"parent_id"`
	MetaData         interface{}   `json:"meta_data"`
	GroupID          string        `json:"group_id"`
	FirstPublishedAt time.Time     `json:"first_published_at"`
	ReleaseID        interface{}   `json:"release_id"`
	Lang             string        `json:"lang"`
	Path             interface{}   `json:"path"`
	Alternates       []interface{} `json:"alternates"`
	DefaultFullSlug  interface{}   `json:"default_full_slug"`
	TranslatedSlugs  interface{}   `json:"translated_slugs"`
}

type Content struct {
	UID       string           `json:"_uid"`
	Body      []map[string]any `json:"body"`
	Component string           `json:"component"`
}
