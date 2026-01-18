package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kordmtal/url-shortrener/internal/config"
)

func main() {
	urls := make(map[string]string)

	cfg := config.Parse()

	r := gin.Default()
	r.POST("/", URLShortenerHandler(urls, cfg.BasicURLServerAdress))
	r.GET("/:id", GetShortURLHandler(urls))

	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
