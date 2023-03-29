package storyblok_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/dryaf/headless_cms/client/storyblok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (any, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Set(key string, obj any) error {
	args := m.Called(key, obj)
	return args.Error(0)
}

func (m *MockCache) Del(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Empty() error {
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

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	cache.On("Empty").Return(nil)
	err := client.EmptyCache(emptyCacheToken)
	assert.Nil(t, err)

	cache.AssertExpectations(t)
}

func TestEmptyCacheWrongToken(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})

	err := client.EmptyCache("wrong_token")
	assert.NotNil(t, err)
	assert.Equal(t, "token incorrect", err.Error())

	cache.AssertExpectations(t)
}

func TestRequestJSON(t *testing.T) {
	token := "test_token"
	emptyCacheToken := "empty_cache_token"
	cache := &MockCache{}
	mockHTTPClient := &MockHTTPClient{}

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

	resp, err := client.RequestJSON(page, version, language)
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

	client := storyblok.NewClient(token, emptyCacheToken, cache, &MockHTTPClient{})
	client.HttpClient = mockHTTPClient

	page := "login"
	version := "published"
	language := "en"

	cacheKey := "j:published:en:login"

	cache.On("Get", cacheKey).Return([]byte(`{"test": "data"}`), nil)

	resp, err := client.RequestJSON(page, version, language)
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

	resp, err := client.RequestJSON(page, version, language)
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

	resp, err := client.RequestJSON(page, version, language)
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
