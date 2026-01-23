package config

import (
	"flag"
	"os"
	"strings"
)

const (
	defaultServerAddress = "localhost:8080"
)

type Config struct {
	ServerAddress        string
	BasicURLServerAdress string
}

func Parse() *Config {
	cfg := &Config{
		ServerAddress:        defaultServerAddress,
		BasicURLServerAdress: defaultServerAddress,
	}

	flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "Server address and port to listen on")
	flag.StringVar(&cfg.BasicURLServerAdress, "b", defaultServerAddress, "Basic URL server address")

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BasicURLServerAdress = envBaseURL
	}

	cfg.BasicURLServerAdress = normalizeBaseURL(cfg.BasicURLServerAdress)

	return cfg
}

func normalizeBaseURL(u string) string {
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	return u
}
