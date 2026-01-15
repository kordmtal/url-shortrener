package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

const ipSrvAddr = "localhost:8080"

func main() {
	urls := make(map[string]string)

	r := gin.Default()
	r.POST("/", URLShortenerHandler(urls))
	r.GET("/:id", GetShortURLHandler(urls))

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
