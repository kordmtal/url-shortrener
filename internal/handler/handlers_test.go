package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kordmtal/url-shortrener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testBaseHost = "localhost:8080"
	testBaseURL  = "http://" + testBaseHost + "/"
)

// performRequest - helper функция для выполнения HTTP запросов в тестах
func performRequest(r *gin.Engine, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// performJSONRequest - helper функция для выполнения JSON запросов в тестах
func performJSONRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestURLShortenerHandler(t *testing.T) {
	type Want struct {
		url         url.URL
		contentType string
		statusCode  int
	}

	// Имитация хранилища (In-memory storage)
	urls := storage.NewMapRepository()

	tests := []struct {
		name string
		url  string
		want Want
	}{
		{
			name: "Test case 1 - new URL",
			url:  "http://example.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: testBaseHost},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 2 - existing URL",
			url:  "http://example.com", // Тот же URL
			want: Want{
				url:         url.URL{Scheme: "http", Host: testBaseHost},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK, // Ожидаем 200, так как уже есть
			},
		},
		{
			name: "Test case 3 - another new URL",
			url:  "https://openai.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: testBaseHost},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 4 - empty URL",
			url:  "",
			want: Want{
				url:         url.URL{},
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "Test case 5 - localhost without scheme",
			url:  "localhost:8080",
			want: Want{
				url:         url.URL{Scheme: "http", Host: testBaseHost},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/", URLShortenerHandler(urls, testBaseURL))

			w := performRequest(r, http.MethodPost, "/", strings.NewReader(tt.url))

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.url == "" {
				return
			}

			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))

			respBody, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			respURL, err := url.Parse(string(respBody))
			require.NoError(t, err, "Response body should be a valid URL")

			assert.Equal(t, tt.want.url.Scheme, respURL.Scheme)
			assert.Equal(t, tt.want.url.Host, respURL.Host)

			shortID := respURL.Path[1:]

			val, err := urls.Get(shortID)
			require.NoError(t, err)
			assert.NotEmpty(t, val, "URL should exist in repository")
			expectedValue := tt.url
			if tt.name == "Test case 5 - localhost without scheme" {
				expectedValue = "http://" + tt.url
			}
			assert.Equal(t, expectedValue, val, "Map value should match expected URL")
		})
	}
}

func TestGetShortURLHandler(t *testing.T) {
	urls := storage.NewMapRepository()

	tests := []struct {
		name             string
		key              string
		url              string // URL для предварительного заполнения мапы
		expectedLocation string // Ожидаемый заголовок Location
		statusCode       int
	}{
		{
			name:             "Test case 1 - Valid ID",
			key:              "abc12345",
			url:              "http://example.com",
			expectedLocation: "http://example.com",
			statusCode:       http.StatusTemporaryRedirect,
		},
		{
			name:             "Test case 2 - Non-existing key",
			key:              "nonexistent",
			url:              "", // Не добавляем в мапу
			expectedLocation: "",
			statusCode:       http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		if tt.url != "" {
			urls.Set(tt.url, tt.key)
		}

		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/:id", GetShortURLHandler(urls))

			w := performRequest(r, http.MethodGet, fmt.Sprintf("/%s", tt.key), nil)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, tt.expectedLocation, w.Header().Get("Location"))
		})
	}
}

func TestURLShortenerJSONHandler(t *testing.T) {
	type Want struct {
		res         Response
		contentType string
		statusCode  int
	}

	// Имитация хранилища (In-memory storage)
	urls := storage.NewMapRepository()

	tests := []struct {
		name string
		req  Request
		want Want
	}{
		{
			name: "Test case 1 - new URL",
			req:  Request{"http://example.com"},
			want: Want{
				res:         Response{testBaseHost},
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 2 - existing URL",
			req:  Request{"http://example.com"},
			want: Want{
				res:         Response{testBaseHost},
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "Test case 3 - another new URL",
			req:  Request{"https://openai.com"},
			want: Want{
				res:         Response{testBaseHost},
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 4 - empty URL",
			req:  Request{""},
			want: Want{
				res:         Response{""},
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "Test case 5 - localhost without scheme",
			req:  Request{"localhost:8080"},
			want: Want{
				res:         Response{testBaseHost},
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/api/shorten", URLShortenerJSONHandler(urls, testBaseURL))

			w := performJSONRequest(r, http.MethodPost, "/api/shorten", tt.req)

			assert.Equal(t, tt.want.statusCode, w.Code)

			var val Response
			err := json.NewDecoder(w.Body).Decode(&val)
			require.NoError(t, err)

			if val.URL == "" {
				return
			}

			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))

			respURL, err := url.Parse(val.URL)
			require.NoError(t, err)
			assert.Equal(t, tt.want.res.URL, respURL.Host)

			shortID := respURL.Path[1:]

			addr, err := urls.Get(shortID)
			require.NoError(t, err)
			assert.NotEmpty(t, addr, "URL should exist in repository")

			expectedValue := tt.req.URL
			if !strings.HasPrefix(tt.req.URL, "http://") && !strings.HasPrefix(tt.req.URL, "https://") {
				expectedValue = "http://" + tt.req.URL
			}
			assert.Equal(t, expectedValue, addr, "Map value should match expected URL")
		})
	}
}
