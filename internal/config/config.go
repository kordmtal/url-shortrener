package config

import (
	"flag"
	"os"
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

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}

	return cfg
}
