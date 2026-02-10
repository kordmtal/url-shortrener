package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kordmtal/url-shortrener/internal/compress"
	"github.com/kordmtal/url-shortrener/internal/config"
	"github.com/kordmtal/url-shortrener/internal/handler"
	"github.com/kordmtal/url-shortrener/internal/logger"
)

func main() {
	urls := make(map[string]string)

	cfg := config.Parse()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GetLogger())
	r.Use(compress.GzipMiddleware())

	r.POST("/", handler.URLShortenerHandler(urls, cfg.BasicURLServerAdress))
	r.GET("/:id", handler.GetShortURLHandler(urls))
	r.POST("/api/shorten", handler.URLShortenerJSONHandler(urls, cfg.BasicURLServerAdress))

	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
