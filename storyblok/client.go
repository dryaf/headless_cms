package storyblok

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// https://www.storyblok.com/docs/api/content-delivery
var statusCodes = map[int]string{
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

type Client struct {
	httpClient *http.Client

	empty_cache_token string
	cache             map[string]map[string]any
	cache_json        map[string][]byte
	cache_texts       map[string]map[string]string

	token   string
	api_url string

	default_version          string // "draft", "published"
	ignore_cache_for_version string
}

// /https://github.com/storyblok/storyblok-ruby
func NewClient(token string, empty_cache_token string) *Client {

	return &Client{
		default_version: "published",

		// Cache stuff
		ignore_cache_for_version: "draft",
		empty_cache_token:        empty_cache_token,
		cache:                    map[string]map[string]any{},
		cache_json:               map[string][]byte{},
		cache_texts:              map[string]map[string]string{},

		api_url:    "https://api.storyblok.com/v2/cdn/stories",
		token:      token,
		httpClient: &http.Client{},
	}
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) EmptyCache(user_input_token string) error {
	if user_input_token != c.empty_cache_token {
		return errors.New("token incorrect")
	}
	c.cache = map[string]map[string]any{}
	c.cache_json = map[string][]byte{}
	c.cache_texts = map[string]map[string]string{}

	return nil
}

// RequestJSON story for example /login or "" for getting all stories
func (c *Client) RequestJSON(page string, version string, language string) (jsonResp []byte, err error) {
	url_params := c.generateQuery(page, version, language)

	// Cache - Read
	var ok bool
	if version != c.ignore_cache_for_version {
		jsonResp, ok = c.cache_json[url_params]
		if ok && jsonResp != nil {
			return jsonResp, nil
		}
	}

	// Remote CMS
	//log.Default().Println("--- storyblok-call: " + c.api_url + url_params + "&token=" + c.token)
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
	if resp.StatusCode == 404 {
		c.cache_json[url_params] = nil
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
	c.cache_json[url_params] = body

	return body, err
}

func (c *Client) RequestTranslatableTexts(page string, version string, language string) (texts map[string]string, err error) {
	url_params := c.generateQuery(page, version, language)

	// Cache - Read
	var ok bool
	if version != c.ignore_cache_for_version {
		texts, ok = c.cache_texts[url_params]
		if ok && texts != nil {
			return texts, nil
		}
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", url_params, err)
	}
	storyblokPage := &StoryWithTranslatableTextsOnly{}
	err = json.Unmarshal(jsonResp, storyblokPage)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", url_params, err)
	}
	texts = map[string]string{}
	for _, tt := range storyblokPage.Story.Content.Body {
		if tt.Component == "_translatable_text" {
			texts[tt.ID] = tt.Value
			texts[tt.ID+"_editable"] = tt.Editable
		}
	}

	// Cache - Write
	c.cache_texts[url_params] = texts

	return texts, nil
}

func (c *Client) Request(page string, version string, language string) (cmsData map[string]any, err error) {
	url_params := c.generateQuery(page, version, language)

	// Cache - Read
	var ok bool
	if version != c.ignore_cache_for_version {
		cmsData, ok = c.cache[url_params]
		if ok && cmsData != nil {
			return cmsData, nil
		}
	}

	jsonResp, err := c.RequestJSON(page, version, language)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: request_json: %w", url_params, err)
	}
	cmsData = map[string]any{}
	err = json.Unmarshal(jsonResp, &cmsData)
	if err != nil {
		return nil, fmt.Errorf("headless_cms: %s: json_unmarshal: %w", url_params, err)
	}

	// Cache - Write
	c.cache[url_params] = cmsData
	return cmsData, nil
}

func (c *Client) generateQuery(page, version, language string) string {
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
	//log.Default().Println("url_params", url_params)
	return url_params
}
