package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kordmtal/url-shortrener/internal/compress"
	"github.com/kordmtal/url-shortrener/internal/config"
	"github.com/kordmtal/url-shortrener/internal/handler"
	"github.com/kordmtal/url-shortrener/internal/logger"
	"github.com/kordmtal/url-shortrener/internal/storage"
)

func main() {
	cfg := config.Parse()

	var rep storage.URLsStorage
	if cfg.FileRepositoryPath == "" {
		rep = storage.NewMapRepository()
	} else {
		var err error
		rep, err = storage.NewFileRepository(cfg.FileRepositoryPath)
		if err != nil {
			log.Fatalf("failed to initialize file repository: %v", err)
		}
	}
	defer rep.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GetLogger())
	r.Use(compress.GzipMiddleware())

	r.POST("/", handler.URLShortenerHandler(rep, cfg.BasicURLServerAdress))
	r.GET("/:id", handler.GetShortURLHandler(rep))
	r.POST("/api/shorten", handler.URLShortenerJSONHandler(rep, cfg.BasicURLServerAdress))

	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
