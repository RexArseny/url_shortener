package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	env "github.com/caarlos0/env/v11"
)

// Default Config values.
const (
	DefaultServerAddress      = "localhost:8080"
	DefaultBasicPath          = "http://localhost:8080"
	DefaultFileStoragePath    = "shorturls.txt"
	DefaultPublicKeyPath      = "public.pem"
	DefaultPrivateKeyPath     = "private.pem"
	DefaultEnableHTTPS        = false
	DefaultCertificatePath    = "cert.pem"
	DefaultCertificateKeyPath = "key.pem"
)

// Config is a set of service configurable variables.
type Config struct {
	ServerAddress      string `env:"SERVER_ADDRESS" json:"server_address"`
	BasicPath          string `env:"BASE_URL" json:"basic_url"`
	FileStoragePath    string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN        string `env:"DATABASE_DSN" json:"database_dsn"`
	PublicKeyPath      string `env:"PUBLIC_KEY_PATH" json:"public_key_path"`
	PrivateKeyPath     string `env:"PRIVATE_KEY_PATH" json:"private_key_path"`
	Config             string `env:"CONFIG" json:"config"`
	TrustedSubnet      string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	CertificatePath    string `env:"CERTIFICATE_PATH" json:"certificate_path"`
	CertificateKeyPath string `env:"CERTIFICATE_KEY_PATH" json:"certificate_key_path"`
	EnableHTTPS        bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

// Init parse values for Config from environment and flags.
func Init() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", DefaultServerAddress, "server address")
	flag.StringVar(&cfg.BasicPath, "b", DefaultBasicPath, "basic path")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.PublicKeyPath, "p", DefaultPublicKeyPath, "public key path")
	flag.StringVar(&cfg.PrivateKeyPath, "k", DefaultPrivateKeyPath, "private key path")
	flag.BoolVar(&cfg.EnableHTTPS, "s", DefaultEnableHTTPS, "enable https")
	flag.StringVar(&cfg.CertificatePath, "cert", DefaultCertificatePath, "certificate path")
	flag.StringVar(&cfg.CertificateKeyPath, "key", DefaultCertificateKeyPath, "certificate key path")
	flag.StringVar(&cfg.Config, "c", "", "config")
	flag.StringVar(&cfg.Config, "config", "", "config")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can not parse env: %w", err)
	}

	if cfg.Config != "" {
		configFile, err := os.ReadFile(cfg.Config)
		if err != nil {
			return nil, fmt.Errorf("can not read config file: %w", err)
		}
		var configFileData Config
		err = json.Unmarshal(configFile, &configFileData)
		if err != nil {
			return nil, fmt.Errorf("can not unmarshal config file: %w", err)
		}

		if cfg.ServerAddress == DefaultServerAddress {
			cfg.ServerAddress = configFileData.ServerAddress
		}
		if cfg.BasicPath == DefaultBasicPath {
			cfg.BasicPath = configFileData.BasicPath
		}
		if cfg.FileStoragePath == DefaultFileStoragePath {
			cfg.FileStoragePath = configFileData.FileStoragePath
		}
		if cfg.DatabaseDSN == "" {
			cfg.DatabaseDSN = configFileData.DatabaseDSN
		}
		if cfg.PublicKeyPath == DefaultPublicKeyPath {
			cfg.PublicKeyPath = configFileData.PublicKeyPath
		}
		if cfg.PrivateKeyPath == DefaultPrivateKeyPath {
			cfg.PrivateKeyPath = configFileData.PrivateKeyPath
		}
		if !cfg.EnableHTTPS {
			cfg.EnableHTTPS = configFileData.EnableHTTPS
		}
		if cfg.CertificatePath == DefaultCertificatePath {
			cfg.CertificatePath = configFileData.CertificatePath
		}
		if cfg.CertificateKeyPath == DefaultCertificateKeyPath {
			cfg.CertificateKeyPath = configFileData.CertificateKeyPath
		}
		if cfg.TrustedSubnet == "" {
			cfg.TrustedSubnet = configFileData.TrustedSubnet
		}
	}

	if cfg.BasicPath[len(cfg.BasicPath)-1] == '/' {
		cfg.BasicPath = cfg.BasicPath[:len(cfg.BasicPath)-1]
	}

	if cfg.TrustedSubnet != "" {
		_, _, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("can not parse cidr from trusted subnet: %w", err)
		}
	}

	return &cfg, nil
}
