package compress

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	gzip *gzip.Writer
}

func ToGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestCompression)
		if err != nil {
			c.String(http.StatusBadRequest, "Error compress")
		}
		defer gz.Close()

		c.Header("Accept-Encoding", "gzip")
		c.Writer = gzipResponseWriter{c.Writer, gz}
		c.Next()
	}
}
