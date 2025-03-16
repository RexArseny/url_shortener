package config

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	oldArgs := os.Args
	oldEnv := os.Environ()
	defer func() {
		os.Args = oldArgs
		os.Clearenv()
		for _, envVar := range oldEnv {
			keyVal := strings.SplitN(envVar, "=", 2)
			os.Setenv(keyVal[0], keyVal[1])
		}
	}()

	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		expectedConfig *Config
		expectedError  string
	}{
		{
			name:    "default values",
			args:    []string{"cmd"},
			envVars: map[string]string{},
			expectedConfig: &Config{
				ServerAddress:   DefaultServerAddress,
				BasicPath:       DefaultBasicPath,
				FileStoragePath: DefaultFileStoragePath,
				DatabaseDSN:     "",
				PublicKeyPath:   DefaultPublicKeyPath,
				PrivateKeyPath:  DefaultPrivateKeyPath,
				EnableHTTPS:     DefaultEnableHTTPS,
			},
			expectedError: "",
		},
		{
			name: "flag overrides",
			args: []string{
				"cmd",
				"-a", "localhost:9090",
				"-b", "http://localhost:9090",
				"-f", "custom.txt",
				"-d", "postgres://user:pass@localhost:5432/db",
				"-p", "custom_public.pem",
				"-k", "custom_private.pem",
				"-s",
			},
			envVars: map[string]string{},
			expectedConfig: &Config{
				ServerAddress:   "localhost:9090",
				BasicPath:       "http://localhost:9090",
				FileStoragePath: "custom.txt",
				DatabaseDSN:     "postgres://user:pass@localhost:5432/db",
				PublicKeyPath:   "custom_public.pem",
				PrivateKeyPath:  "custom_private.pem",
				EnableHTTPS:     true,
			},
			expectedError: "",
		},
		{
			name: "environment variable overrides",
			args: []string{"cmd"},
			envVars: map[string]string{
				"SERVER_ADDRESS":    "localhost:9090",
				"BASE_URL":          "http://localhost:9090",
				"FILE_STORAGE_PATH": "custom.txt",
				"DATABASE_DSN":      "postgres://user:pass@localhost:5432/db",
				"PUBLIC_KEY_PATH":   "custom_public.pem",
				"PRIVATE_KEY_PATH":  "custom_private.pem",
				"ENABLE_HTTPS":      "true",
			},
			expectedConfig: &Config{
				ServerAddress:   "localhost:9090",
				BasicPath:       "http://localhost:9090",
				FileStoragePath: "custom.txt",
				DatabaseDSN:     "postgres://user:pass@localhost:5432/db",
				PublicKeyPath:   "custom_public.pem",
				PrivateKeyPath:  "custom_private.pem",
				EnableHTTPS:     true,
			},
			expectedError: "",
		},
		{
			name: "environment variable parsing error",
			args: []string{"cmd"},
			envVars: map[string]string{
				"ENABLE_HTTPS": "invalid",
			},
			expectedConfig: nil,
			expectedError:  "can not parse env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Clearenv()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := Init()

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, *cfg, *tt.expectedConfig)
		})
	}
}
