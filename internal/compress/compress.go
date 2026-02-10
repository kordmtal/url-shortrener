package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipResponseWriter wraps gin.ResponseWriter to compress response data
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer         *gzip.Writer
	wroteHeader    bool
	shouldCompress bool
}

// Write compresses data before writing to the underlying ResponseWriter
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	// Check content type on first write
	if !g.wroteHeader {
		g.wroteHeader = true
		contentType := g.Header().Get("Content-Type")

		// Only compress application/json and text/html
		g.shouldCompress = strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")

		if g.shouldCompress {
			// Set Content-Encoding header only when we compress
			g.Header().Set("Content-Encoding", "gzip")
		}
	}

	if !g.shouldCompress {
		return g.ResponseWriter.Write(data)
	}

	return g.writer.Write(data)
}

// WriteString compresses string data before writing
func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

// GzipMiddleware handles both request decompression and response compression
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Handle request decompression if Content-Encoding: gzip
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.String(http.StatusBadRequest, "Error decompressing request")
				c.Abort()
				return
			}
			defer reader.Close()
			c.Request.Body = io.NopCloser(reader)
		}

		// Handle response compression if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Create gzip writer for response
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestCompression)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error initializing compression")
			c.Abort()
			return
		}

		// Wrap the response writer (don't set Content-Encoding yet, will be set in Write if needed)
		gzWriter := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}
		c.Writer = gzWriter

		c.Next()

		// Only close and flush if we actually compressed
		if gzWriter.shouldCompress {
			gz.Flush()
			gz.Close()
		}
	}
}
