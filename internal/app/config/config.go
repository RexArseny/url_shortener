package config

import (
	"flag"
	"fmt"

	env "github.com/caarlos0/env/v11"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBasicPath     = "http://localhost:8080"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BasicPath     string `env:"BASE_URL"`
}

func Init() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", DefaultServerAddress, "server address")
	flag.StringVar(&cfg.BasicPath, "b", DefaultBasicPath, "basic path")

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
