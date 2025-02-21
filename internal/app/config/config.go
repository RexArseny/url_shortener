package config

import (
	"flag"
	"fmt"

	env "github.com/caarlos0/env/v11"
)

// Default Config values.
const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBasicPath       = "http://localhost:8080"
	DefaultFileStoragePath = "shorturls.txt"
	DefaultPublicKeyPath   = "public.pem"
	DefaultPrivateKeyPath  = "private.pem"
)

// Config is a set of service configurable variables.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BasicPath       string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	PublicKeyPath   string `env:"PUBLIC_KEY_PATH"`
	PrivateKeyPath  string `env:"PRIVATE_KEY_PATH"`
}

// Init parse values for Config from environment and flags.
func Init() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", DefaultServerAddress, "server address")
	flag.StringVar(&cfg.BasicPath, "b", DefaultBasicPath, "basic path")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.PublicKeyPath, "p", DefaultPublicKeyPath, "public key path")
	flag.StringVar(&cfg.PrivateKeyPath, "s", DefaultPrivateKeyPath, "private key path")

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
