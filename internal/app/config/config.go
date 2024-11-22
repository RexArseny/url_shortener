package config

import (
	"flag"
	"fmt"

	env "github.com/caarlos0/env/v11"
)

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBasicPath       = "http://localhost:8080"
	DefaultFileStoragePath = "shorturls.txt"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BasicPath       string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func Init() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", DefaultServerAddress, "server address")
	flag.StringVar(&cfg.BasicPath, "b", DefaultBasicPath, "basic path")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "file storage path")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can not parse env: %w", err)
	}

	if cfg.BasicPath[len(cfg.BasicPath)-1] == '/' {
		cfg.BasicPath = cfg.BasicPath[:len(cfg.BasicPath)-1]
	}

	return &cfg, nil
}
