package main

import (
	"net/http"
)

const ipSrvAddr = "localhost:8080"

func main() {
	var urls = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc("/", URLShortenerHandler(&urls))
	mux.HandleFunc("/{id}", GetShortURLHandler(&urls))

	err := http.ListenAndServe(ipSrvAddr, mux)
	if err != nil {
		panic(err)
	}
}
