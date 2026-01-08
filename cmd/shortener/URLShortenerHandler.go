package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
)

func URLShortenerHandler(urls *map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Method not allowed", http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Error reading body", http.StatusBadRequest)
			return
		}

		if string(body) == "" {
			http.Error(res, "Empty URL", http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "text/plain")

		for k, v := range *urls {
			if v == string(body) {
				io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, k))
				return
			}
		}

		res.WriteHeader(http.StatusCreated)

		id := rand.Text()
		(*urls)[id] = string(body)
		io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, id))
	}
}

func GetShortURLHandler(urls *map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusBadRequest)
			return
		}

		id := req.URL.Path[1:]
		originalURL, exists := (*urls)[id]
		if !exists {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
