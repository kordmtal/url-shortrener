package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	URL string `json:"result"`
}

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

		if !strings.HasPrefix(reqURL, "http://") && !strings.HasPrefix(reqURL, "https://") {
			reqURL = "http://" + reqURL
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

func URLShortenerJSONHandler(urls map[string]string, basicURLServerAdress string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request
		var res Response

		body := c.Request.Body
		if err := json.NewDecoder(body).Decode(&req); err != nil {
			c.String(http.StatusBadRequest, "Error parse body")
			return
		}

		sendJSON := func(status int, r Response) {
			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(&r)
			c.Data(status, "application/json; charset=utf-8", buf.Bytes())
		}

		if req.URL == "" {
			sendJSON(http.StatusBadRequest, res)
			return
		}

		if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
			req.URL = "http://" + req.URL
		}

		for k, v := range urls {
			if v == req.URL {
				res.URL = basicURLServerAdress + k
				sendJSON(http.StatusOK, res)
				return
			}
		}

		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			c.String(http.StatusBadRequest, "Error generate ID")
			return
		}
		resID := base64.RawURLEncoding.EncodeToString(b)

		urls[resID] = req.URL

		res.URL = basicURLServerAdress + resID

		sendJSON(http.StatusCreated, res)
	}
}
