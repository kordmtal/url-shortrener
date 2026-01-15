package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerHandler(t *testing.T) {
	type Want struct {
		url         url.URL
		contentType string
		statusCode  int
	}

	// Имитация хранилища (In-memory storage)
	urls := make(map[string]string)

	tests := []struct {
		name string
		url  string
		want Want
	}{
		{
			name: "Test case 1 - new URL",
			url:  "http://example.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: ipSrvAddr},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 2 - existing URL",
			url:  "http://example.com", // Тот же URL
			want: Want{
				url:         url.URL{Scheme: "http", Host: ipSrvAddr},
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK, // Ожидаем 200, так как уже есть
			},
		},
		{
			name: "Test case 3 - another new URL",
			url:  "https://openai.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: ipSrvAddr},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/", URLShortenerHandler(urls))

			body := strings.NewReader(tt.url)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)

			result := res.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.url == "" {
				return
			}

			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			respBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			respURL, err := url.Parse(string(respBody))
			require.NoError(t, err, "Response body should be a valid URL")

			assert.Equal(t, tt.want.url.Scheme, respURL.Scheme)
			assert.Equal(t, tt.want.url.Host, respURL.Host)

			shortID := respURL.Path[1:]

			val, exists := urls[shortID]
			assert.True(t, exists, "Key should exist in map")
			assert.Equal(t, tt.url, val, "Map value should match original URL")
		})
	}
}

func TestGetShortURLHandler(t *testing.T) {
	urls := make(map[string]string)

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
			urls[tt.key] = tt.url
		}

		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/:id", GetShortURLHandler(urls))

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.key), nil)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)

			result := res.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.statusCode, result.StatusCode)
			assert.Equal(t, tt.expectedLocation, result.Header.Get("Location"))
		})
	}
}
