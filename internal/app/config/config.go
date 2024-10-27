package config

import (
	"fmt"
	"net/url"
	"path"

	env "github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
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
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg.ServerAddress == "" {
		pflag.StringVarP(&cfg.ServerAddress, "a", "a", DefaultServerAddress, "server address")
	}
	if cfg.BasicPath == "" {
		pflag.StringVarP(&cfg.BasicPath, "b", "b", DefaultBasicPath, "basic path")
	}
	return &cfg, nil
}

func (c *Config) GetURLPrefix() (*string, error) {
	serverAddress, err := url.ParseRequestURI(c.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid server address: %s", err)
	}
	basicPath, err := url.ParseRequestURI(c.BasicPath)
	if err != nil {
		return nil, fmt.Errorf("invalid basic path: %s", err)
	}
	if serverAddress.String() != basicPath.Host {
		return nil, fmt.Errorf("server address does not correspond with basic path")
	}
	if c.BasicPath[len(c.BasicPath)-1] == '/' {
		c.BasicPath = c.BasicPath[:len(c.BasicPath)-1]
	}
	urlPrefix := path.Base(basicPath.Path)
	return &urlPrefix, nil
}
