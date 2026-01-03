package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
)

const ipSrvAddr = "localhost:8080"

var urlMap = make(map[string]string)

func urlShortener(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
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

		for k, v := range urlMap {
			if v == string(body) {
				io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, k))
				return
			}
		}

		res.WriteHeader(http.StatusCreated)

		id := rand.Text()
		urlMap[id] = string(body)
		io.WriteString(res, fmt.Sprintf("http://%s/%s", ipSrvAddr, id))
	case http.MethodGet:
		if req.Header.Get("Content-Type") != "text/plain" {
			http.Error(res, "Unsupported Media Type", http.StatusBadRequest)
			return
		}
		id := req.URL.Path[1:]
		originalURL, exists := urlMap[id]
		if !exists {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	default:
		http.Error(res, "Method not allowed", http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", urlShortener)

	err := http.ListenAndServe(ipSrvAddr, mux)
	if err != nil {
		panic(err)
	}
}
