package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

func URLShortenerHandler(urls *map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqURL, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusBadRequest, "Error parse body")
			return
		}

		for k, v := range *urls {
			if v == string(reqURL) {
				c.String(http.StatusOK, "http://%s/%s", ipSrvAddr, k)
				return
			}
		}

		b := make([]byte, 8)
		_, err = rand.Read(b)
		if err != nil {
			c.String(http.StatusBadRequest, "Error generate ID")
			return
		}
		resId := base64.RawURLEncoding.EncodeToString(b)

		(*urls)[resId] = string(reqURL)
		c.String(http.StatusOK, "http://%s/%s", ipSrvAddr, resId)
	}
}

func GetShortURLHandler(urls *map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.Param("id")

		if reqId == "" {
			c.String(http.StatusBadRequest, "Invalid ID")
			return
		}

		originalURL, exists := (*urls)[reqId]
		if !exists {
			c.String(http.StatusBadRequest, "URL not found")
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, originalURL)
	}
}
