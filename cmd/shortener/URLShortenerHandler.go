package main

import (
	"crypto/rand"
	"encoding/base64"
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
				res.WriteHeader(http.StatusOK)
				io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, k))
				return
			}
		}

		res.WriteHeader(http.StatusCreated)

		b := make([]byte, 8)
		_, err = rand.Read(b)
		if err != nil {
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}
		id := base64.RawURLEncoding.EncodeToString(b)

		(*urls)[id] = string(body)
		io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, id))
	}
}

func GetShortURLHandler(urls *map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusBadRequest)
			return
		}

		if len(req.URL.Path) < 2 {
			http.Error(res, "Invalid ID", http.StatusBadRequest)
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
