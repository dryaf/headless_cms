package storyblok_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/dryaf/headless_cms"
	"github.com/dryaf/headless_cms/cache/redis_cache"
	"github.com/dryaf/headless_cms/client/storyblok"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func httpResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func getRedisCacheForTests() headless_cms.Cache {
	useRealRedis, _ := strconv.ParseBool(os.Getenv("USE_REAL_REDIS"))

	if useRealRedis {
		// Replace the values below with your actual Redis settings
		redisAddrs := []string{os.Getenv("REDIS_ADDR")}
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisMasterName := os.Getenv("REDIS_MASTER_NAME")
		redisDB := 0
		return redis_cache.New(redis.NewUniversalClient(&redis.UniversalOptions{Addrs: redisAddrs, Password: redisPassword, DB: redisDB, MasterName: redisMasterName}))
	}

	return nil
}

func TestNewClient(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, &MockHTTPClient{})

	assert.NotNil(t, client)
	assert.Equal(t, token, client.AuthToken())
	assert.Equal(t, cache, client.Cache())
}

func TestEmptyCache(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, &MockHTTPClient{})

	cache.On("Empty").Return(nil)
	err := client.EmptyCache(ctx, emptyCacheToken)
	assert.Nil(t, err)

	cache.AssertExpectations(t)
}

func TestEmptyCacheWrongToken(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, &MockHTTPClient{})

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
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "published"
	language := "en"

	mockResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"test": "data"}`))),
	}
	cache.On("Get", "j:published:en:login").Return(nil, nil)
	mockHTTPClient.On("Do", mock.Anything).Return(mockResp, nil)
	cache.On("Set", "j:published:en:login", []byte(`{"test": "data"}`)).Return(nil)

	resp, err := client.GetPageAsJSON(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`{"test": "data"}`), resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}
func TestEmptyCacheToken(t *testing.T) {
	token := "cms_auth_token"
	expectedToken := "empty_cache_token"

	cache := &MockCache{}

	// Test with valid cache action token

	client := storyblok.NewClient(context.Background(), token, expectedToken, cache, &MockHTTPClient{})
	actualToken, err := client.EmptyCacheToken(context.Background())
	assert.Equal(t, expectedToken, actualToken)
	assert.Nil(t, err)
}

func TestRequestJSONCache(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "published"
	language := "en"

	cacheKey := "j:published:en:login"

	cache.On("Get", cacheKey).Return([]byte(`{"test": "data"}`), nil)

	resp, err := client.GetPageAsJSON(ctx, page, version, language)
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
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)
	client.HttpClient = mockHTTPClient
	httpResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"test": "data"}`))),
	}
	mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(httpResponse, nil)

	page := "login"
	version := "draft"
	language := "en"

	resp, err := client.GetPageAsJSON(ctx, page, version, language)
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
	ctx := context.Background()

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "draft"
	language := "en"

	mockResp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"error": "Server error"}`))),
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResp, nil)

	resp, err := client.GetPageAsJSON(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	mockHTTPClient.AssertExpectations(t)
}

// ... Add similar tests for RequestTranslatableTexts, Request, and RequestSimpleBlocksWithID ...

func TestGenerateKey(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)

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
	ctx := context.Background()
	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)

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

	resp, err := client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
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
	resp, err = client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
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

	resp, err = client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	cache.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestRequestSimpleBlocksWithIDRedis(t *testing.T) {
	ctx := context.Background()
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := getRedisCacheForTests()
	if cache == nil {
		t.Skip("Skipping Redis tests")
	}
	cache.Empty(ctx)
	mockHTTPClient := &MockHTTPClient{}

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, cache, mockHTTPClient)

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

	resp, err := client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))

	for _, block := range sampleBlocks {
		id, _ := block["id"].(string)
		assert.Equal(t, block, resp[id])
	}

	mockHTTPClient.AssertExpectations(t)

	// Test cache hit

	resp, err = client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
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
	resp, err = client.GetPageAsSimpleBlocksWithID(ctx, page, version, language)
	assert.NotNil(t, err)
	assert.Nil(t, resp)

	mockHTTPClient.AssertExpectations(t)
}

func TestClient_Request(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	mockCache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}

	client := storyblok.NewClient(context.Background(), token, emptyCacheToken, mockCache, mockHTTPClient)

	t.Run("successful request", func(t *testing.T) {
		ctx := context.Background()
		page := "login"
		version := "published"
		language := "en"
		mockCache.ExpectedCalls = nil
		mockHTTPClient.ExpectedCalls = nil

		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			httpResponse(http.StatusOK, []byte(`
			{
				"content": "test content", 
			 	"node": { 
					"name": "node 1", 
					"node": { 
						"name": "node 2", 
						"bool": true, 
						"age": 18
				    }
				}
			}
			`)), nil,
		)

		mockCache.On("Get", "r:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Set", "j:published:en:login", mock.Anything).Return(nil)
		mockCache.On("Set", "r:published:en:login", mock.Anything).Return(nil)

		resp, err := client.GetPage(ctx, page, version, language)
		require.NoError(t, err)

		expectedResp := map[string]any{
			"content": "test content",
			"node": map[string]any{
				"name": "node 1",
				"node": map[string]any{
					"name": "node 2",
					"bool": true,
					"age":  float64(18),
				},
			},
		}
		assert.Equal(t, expectedResp, resp)
	})

	t.Run("error from HTTP client", func(t *testing.T) {
		ctx := context.Background()
		page := "login"
		version := "published"
		language := "en"
		mockCache.ExpectedCalls = nil
		mockHTTPClient.ExpectedCalls = nil

		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			(*http.Response)(nil), errors.New("HTTP client error"),
		)

		mockCache.On("Get", "r:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))

		_, err := client.GetPage(ctx, page, version, language)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP client error")
	})

	t.Run("error status code", func(t *testing.T) {
		ctx := context.Background()
		page := "login"
		version := "published"
		language := "en"
		mockCache.ExpectedCalls = nil
		mockHTTPClient.ExpectedCalls = nil

		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			httpResponse(http.StatusInternalServerError, []byte("Internal Server Error")), nil,
		)

		mockCache.On("Get", "r:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))

		_, err := client.GetPage(ctx, page, version, language)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Server Errors")
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		ctx := context.Background()
		page := "login"
		version := "published"
		language := "en"
		mockCache.ExpectedCalls = nil
		mockHTTPClient.ExpectedCalls = nil

		mockHTTPClient.On("Do", mock.AnythingOfType("*http.Request")).Return(
			httpResponse(http.StatusOK, []byte("invalid JSON")), nil,
		)

		mockCache.On("Get", "r:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Get", "j:published:en:login").Return(nil, errors.New("cache miss"))
		mockCache.On("Set", "j:published:en:login", mock.Anything).Return(nil)

		_, err := client.GetPage(ctx, page, version, language)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})

	t.Run("data in cache", func(t *testing.T) {
		ctx := context.Background()
		page := "example"
		version := "published"
		language := "en"
		mockCache.ExpectedCalls = nil
		mockHTTPClient.ExpectedCalls = nil

		mockCache.On("Get", mock.Anything).Return([]byte(`{"content": "cached content"}`), nil)

		resp, err := client.GetPage(ctx, page, version, language)
		require.NoError(t, err)

		expectedResp := map[string]any{
			"content": "cached content",
		}
		assert.Equal(t, expectedResp, resp)
	})
}
