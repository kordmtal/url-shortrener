package config

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress      = "localhost:8080"
	defaultFileRepositoryPath = "/tmp/short-url-db.json"
)

type Config struct {
	ServerAddress        string `env:"SERVER_ADDRESS"`
	BasicURLServerAdress string `env:"BASE_URL"`
	FileRepositoryPath   string `env:"FILE_STORAGE_PATH"`
}

func Parse() *Config {
	cfg := Config{}

	// Сначала регистрируем флаги с дефолтными значениями
	var serverAddr string
	var baseURL string
	var fileRepositoryPath string

	flag.StringVar(&serverAddr, "a", defaultServerAddress, "Server address and port to listen on")
	flag.StringVar(&baseURL, "b", defaultServerAddress, "Basic URL server address")
	flag.StringVar(&fileRepositoryPath, "f", defaultFileRepositoryPath, "Path to file repository")
	flag.Parse()

	// Затем парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Применяем приоритет: env -> flag -> default
	// Если env переменная не установлена, используем значение из флага
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = serverAddr
	}

	if cfg.BasicURLServerAdress == "" {
		cfg.BasicURLServerAdress = baseURL
	}

	if cfg.FileRepositoryPath == "" {
		cfg.FileRepositoryPath = fileRepositoryPath
	}

	cfg.BasicURLServerAdress = normalizeBaseURL(cfg.BasicURLServerAdress)

	return &cfg
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
