package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_RequestDecompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		body            string
		contentEncoding string
		wantBody        string
		wantStatus      int
	}{
		{
			name:            "Decompress gzipped request",
			body:            "test data",
			contentEncoding: "gzip",
			wantBody:        "test data",
			wantStatus:      http.StatusOK,
		},
		{
			name:            "Pass through non-gzipped request",
			body:            "plain text",
			contentEncoding: "",
			wantBody:        "plain text",
			wantStatus:      http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(GzipMiddleware())
			r.POST("/test", func(c *gin.Context) {
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				c.String(http.StatusOK, string(body))
			})

			var reqBody io.Reader
			if tt.contentEncoding == "gzip" {
				var buf bytes.Buffer
				gzWriter := gzip.NewWriter(&buf)
				_, err := gzWriter.Write([]byte(tt.body))
				require.NoError(t, err)
				gzWriter.Close()
				reqBody = &buf
			} else {
				reqBody = bytes.NewBufferString(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/test", reqBody)
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestGzipMiddleware_ResponseCompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		acceptEncoding string
		contentType    string
		responseBody   string
		wantCompressed bool
	}{
		{
			name:           "Compress JSON response when client accepts gzip",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			responseBody:   `{"message":"test"}`,
			wantCompressed: true,
		},
		{
			name:           "Compress HTML response when client accepts gzip",
			acceptEncoding: "gzip",
			contentType:    "text/html",
			responseBody:   "<html><body>test</body></html>",
			wantCompressed: true,
		},
		{
			name:           "Don't compress when client doesn't accept gzip",
			acceptEncoding: "",
			contentType:    "application/json",
			responseBody:   `{"message":"test"}`,
			wantCompressed: false,
		},
		{
			name:           "Don't compress non-JSON/HTML content",
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			responseBody:   "plain text",
			wantCompressed: false,
		},
		{
			name:           "Compress with multiple encodings",
			acceptEncoding: "deflate, gzip, br",
			contentType:    "application/json",
			responseBody:   `{"message":"test"}`,
			wantCompressed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(GzipMiddleware())
			r.GET("/test", func(c *gin.Context) {
				c.Data(http.StatusOK, tt.contentType, []byte(tt.responseBody))
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.wantCompressed {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

				// Decompress and verify
				gzReader, err := gzip.NewReader(w.Body)
				require.NoError(t, err)
				defer gzReader.Close()

				decompressed, err := io.ReadAll(gzReader)
				require.NoError(t, err)
				assert.Equal(t, tt.responseBody, string(decompressed))
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"))
				assert.Equal(t, tt.responseBody, w.Body.String())
			}
		})
	}
}

func TestGzipMiddleware_JSONResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(GzipMiddleware())
	r.POST("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodPost, "/json", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Decompress and verify JSON
	gzReader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	require.NoError(t, err)
	assert.Contains(t, string(decompressed), "success")
}

func TestGzipMiddleware_BothDirections(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(GzipMiddleware())
	r.POST("/echo", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		c.Data(http.StatusOK, "application/json", body)
	})

	// Compress request body
	var reqBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&reqBuf)
	_, err := gzWriter.Write([]byte("compressed request"))
	require.NoError(t, err)
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/echo", &reqBuf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	// Decompress response
	gzReader, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	require.NoError(t, err)
	assert.Equal(t, "compressed request", string(decompressed))
}

func TestGzipMiddleware_InvalidGzipRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(GzipMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach here")
	})

	// Send invalid gzip data
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("not gzipped"))
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Error decompressing request")
}
