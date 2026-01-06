package main

import (
	"net/http"
)

const ipSrvAddr = "localhost:8080"

func main() {
	var urls = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc("/", UrlShortenerHandler(&urls))
	mux.HandleFunc("/{id}", GetShortUrlHandler(&urls))

	err := http.ListenAndServe(ipSrvAddr, mux)
	if err != nil {
		panic(err)
	}
}
