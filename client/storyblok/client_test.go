package storyblok_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/dryaf/headless_cms"
	"github.com/dryaf/headless_cms/cache/redis_cache"
	"github.com/dryaf/headless_cms/client/storyblok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(key)
	bb, ok := args.Get(0).([]byte)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return bb, args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, obj []byte) error {
	args := m.Called(key, obj)
	return args.Error(0)
}

func (m *MockCache) Del(ctx context.Context, key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Empty(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func getRedisCacheForTests() headless_cms.Cache {
	useRealRedis, _ := strconv.ParseBool(os.Getenv("USE_REAL_REDIS"))

	if useRealRedis {
		// Replace the values below with your actual Redis settings
		addr := "localhost:6379"
		password := ""
		db := 0
		return redis_cache.New(addr, password, db)
	}

	return nil
}

func TestNewClient(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	assert.NotNil(t, client)
	assert.Equal(t, token, client.Token())
	assert.Equal(t, cache, client.Cache())
}

func TestEmptyCache(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	cache.On("Empty").Return(nil)
	err := client.EmptyCache(ctx, emptyCacheToken)
	assert.Nil(t, err)

	cache.AssertExpectations(t)
}

func TestEmptyCacheWrongToken(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	err := client.EmptyCache(ctx, "wrong_token")
	assert.NotNil(t, err)
	assert.Equal(t, "token incorrect", err.Error())

	cache.AssertExpectations(t)
}

func TestRequestJSON(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "published"
	language := "en"

	mockResp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"test": "data"}`))),
	}
	cache.On("Get", "j:published:en:login").Return(nil, nil)
	mockHTTPClient.On("Do", mock.Anything).Return(mockResp, nil)
	cache.On("Set", "j:published:en:login", []byte(`{"test": "data"}`)).Return(nil)

	resp, err := client.RequestJSON(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"test": "data"}`), resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRequestJSONCache(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "published"
	language := "en"

	cacheKey := "j:published:en:login"

	cache.On("Get", cacheKey).Return([]byte(`{"test": "data"}`), nil)

	resp, err := client.RequestJSON(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"test": "data"}`), resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRequestJSONCachingInDraft(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})
	client.HttpClient = mockHTTPClient
	httpResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"test": "data"}`))),
	}
	mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResponse, nil)

	page := "login"
	version := "draft"
	language := "en"

	resp, err := client.RequestJSON(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"test": "data"}`), resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRequestJSONError(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "draft"
	language := "en"

	mockResp := &http.Response{
		StatusCode: 500,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"error": "Server error"}`))),
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResp, nil)

	resp, err := client.RequestJSON(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	mockHTTPClient.AssertExpectations(t)
}

// ... Add similar tests for RequestTranslatableTexts, Request, and RequestSimpleBlocksWithID ...

func TestGenerateKey(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	page := "login"
	version := "draft"
	language := "en"

	expectedKey := "k:draft:en:login"

	key := client.CacheKey("k", page, version, language)
	assert.Equal(t, expectedKey, key)
}

func TestRequestSimpleBlocksWithID(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.TODO()

	client := storyblok.NewClient(token, emptyCacheToken, cache, mockHTTPClient)

	page := "login"
	version := "published"
	language := "en"

	sampleBlocks := []map[string]any{
		{
			"id":   "1",
			"text": "comp1",
			"num":  float64(1),
			"bool": true,
		},
		{
			"id":   "2",
			"text": "comp2",
			"num":  float64(2),
			"bool": false,
		},
	}

	sampleCMSData := storyblok.SimpleBlockskWithID{
		Story: storyblok.Story{
			Content: storyblok.Content{
				Body: sampleBlocks,
			},
		},
	}

	jsonData, _ := json.Marshal(sampleCMSData)

	// Test cache miss and HTTP request
	cache.On("Get", "i:published:en:login").Return(nil, errors.New("cache miss"))
	cache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))
	cache.On("Set", "j:published:en:login", mock.Anything).Return(nil)
	cache.On("Set", "i:published:en:login", mock.Anything).Return(nil)
	mockHTTPClient.On("Do", mock.Anything).Return(httpResponse(http.StatusOK, jsonData), nil)

	resp, err := client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))

	for _, block := range sampleBlocks {
		id, _ := block["id"].(string)
		assert.Equal(t, block, resp[id])
	}

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)

	// Test cache hit
	cache.ExpectedCalls = nil
	respJson, _ := json.Marshal(resp)
	cache.On("Get", "i:published:en:login").Return(respJson, nil)
	resp, err = client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))

	for _, block := range sampleBlocks {
		id, _ := block["id"].(string)
		assert.Equal(t, block, resp[id])
	}

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)

	// Test error case
	cache.ExpectedCalls = nil
	cache.On("Get", "i:published:en:login").Return(nil, errors.New("cache miss"))
	cache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))
	cache.On("Set", "j:published:en:login", mock.Anything).Return(nil)
	cache.On("Set", "i:published:en:login", mock.Anything).Return(nil)
	mockHTTPClient.On("Do", mock.Anything).Return(httpResponse(http.StatusInternalServerError, nil), nil)

	resp, err = client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRequestSimpleBlocksWithIDRedis(t *testing.T) {
	ctx := context.TODO()
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := getRedisCacheForTests()
	if cache == nil {
		t.Skip("Skipping Redis tests")
	}
	cache.Empty(ctx)
	mockHTTPClient := &MockHTTPClient{}

	client := storyblok.NewClient(token, emptyCacheToken, cache, mockHTTPClient)

	page := "login"
	version := "published"
	language := "en"

	sampleBlocks := []map[string]any{
		{
			"id":   "1",
			"text": "comp1",
			"num":  float64(1),
			"bool": true,
		},
		{
			"id":   "2",
			"text": "comp2",
			"num":  float64(2),
			"bool": false,
		},
	}

	sampleCMSData := storyblok.SimpleBlockskWithID{
		Story: storyblok.Story{
			Content: storyblok.Content{
				Body: sampleBlocks,
			},
		},
	}

	jsonData, _ := json.Marshal(sampleCMSData)

	// Test cache miss and HTTP request
	mockHTTPClient.On("Do", mock.Anything).Return(httpResponse(http.StatusOK, jsonData), nil)

	resp, err := client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))

	for _, block := range sampleBlocks {
		id, _ := block["id"].(string)
		assert.Equal(t, block, resp[id])
	}

	mockHTTPClient.AssertExpectations(t)

	// Test cache hit

	resp, err = client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))

	for _, block := range sampleBlocks {
		id, _ := block["id"].(string)
		assert.Equal(t, block, resp[id])
	}

	mockHTTPClient.AssertExpectations(t)

	// Test error case
	mockHTTPClient.On("Do", mock.Anything).Return(httpResponse(http.StatusInternalServerError, nil), nil)
	cache.Empty(ctx)
	resp, err = client.RequestSimpleBlocksWithID(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	mockHTTPClient.AssertExpectations(t)
}

func httpResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}
}
