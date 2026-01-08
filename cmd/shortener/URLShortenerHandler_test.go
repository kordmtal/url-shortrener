package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLShortenerHandler(t *testing.T) {
	type Want struct {
		url         url.URL
		contentType string
		statusCode  int
	}

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
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name: "Test case 2 - existing URL",
			url:  "http://example.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: ipSrvAddr},
				contentType: "text/plain",
				statusCode:  http.StatusOK},
		},
		{
			name: "Test case 3 - another new URL",
			url:  "https://openai.com",
			want: Want{
				url:         url.URL{Scheme: "http", Host: ipSrvAddr},
				contentType: "text/plain",
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
			body := io.NopCloser(strings.NewReader(tt.url))
			r := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()

			defer r.Body.Close()

			handler := URLShortenerHandler(&urls)
			handler(w, r)

			defer w.Result().Body.Close()

			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)

			if tt.url == "" {
				return
			}

			assert.Equal(t, tt.want.contentType, w.Result().Header.Get("Content-Type"))
			respBody, err := io.ReadAll(w.Result().Body)
			require.NoError(t, err)
			respURL, err := url.Parse(string(respBody))
			require.NoError(t, err)
			assert.Equal(t, tt.want.url.Scheme, respURL.Scheme)
			assert.Equal(t, tt.want.url.Host, respURL.Host)

			for k, v := range urls {
				if v == tt.url {
					assert.Equal(t, respURL.Path[1:], k)
				}
			}
		})
	}
}

func TestGetShortURLHandler(t *testing.T) {
	urls := make(map[string]string)

	tests := []struct {
		name       string
		url        string
		key        string
		statusCode int
	}{
		{
			name:       "Test case 1",
			key:        "abc123",
			url:        "http://example.com",
			statusCode: http.StatusTemporaryRedirect,
		},
		{
			name:       "Test case 2 - non-existing key",
			key:        "nonexistent",
			url:        "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		if tt.url != "" {
			urls[tt.key] = tt.url
		}

		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.key), nil)
			w := httptest.NewRecorder()

			defer r.Body.Close()

			handler := GetShortURLHandler(&urls)
			handler(w, r)

			defer w.Result().Body.Close()

			assert.Equal(t, tt.statusCode, w.Result().StatusCode)
			assert.Equal(t, tt.url, w.Result().Header.Get("Location"))
		})
	}
}
