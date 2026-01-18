package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func URLShortenerHandler(urls map[string]string, basicURLServerAdress string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusBadRequest, "Error parse body")
			return
		}

		reqURL := string(body)

		if reqURL == "" {
			c.String(http.StatusBadRequest, "Error parse body")
			return
		}

		if strings.Split(basicURLServerAdress, ":")[0] == "localhost" {
			basicURLServerAdress = "http://" + basicURLServerAdress + "/"
		}

		for k, v := range urls {
			if v == reqURL {
				c.String(http.StatusOK, "%s%s", basicURLServerAdress, k)
				return
			}
		}

		b := make([]byte, 8)
		_, err = rand.Read(b)
		if err != nil {
			c.String(http.StatusBadRequest, "Error generate ID")
			return
		}
		resID := base64.RawURLEncoding.EncodeToString(b)

		urls[resID] = reqURL

		c.String(http.StatusCreated, "%s%s", basicURLServerAdress, resID)
	}
}

func GetShortURLHandler(urls map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.Param("id")

		if reqID == "" {
			c.String(http.StatusBadRequest, "Invalid ID")
			return
		}

		originalURL, exists := urls[reqID]
		if !exists {
			c.String(http.StatusBadRequest, "URL not found")
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, originalURL)
	}
}
