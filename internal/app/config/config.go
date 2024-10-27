package config

import (
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/pflag"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBasicPath     = "http://localhost:8080"
)

type Config struct {
	ServerAddress string
	BasicPath     string
}

func Init() *Config {
	cfg := &Config{
		ServerAddress: DefaultServerAddress,
		BasicPath:     DefaultBasicPath,
	}
	pflag.StringVar(&cfg.ServerAddress, "a", DefaultServerAddress, "server address")
	pflag.StringVar(&cfg.BasicPath, "b", DefaultBasicPath, "basic path")
	return cfg
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
