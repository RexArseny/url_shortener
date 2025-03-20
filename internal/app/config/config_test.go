//nolint:reassign // reassign for tests
package config

import (
	"encoding/json"
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
			err := os.Setenv(keyVal[0], keyVal[1])
			assert.NoError(t, err)
		}
	}()

	tests := []struct {
		name            string
		args            []string
		envVars         map[string]string
		validConfigFile bool
		expectedConfig  *Config
		expectedError   string
	}{
		{
			name:            "default values",
			args:            []string{"cmd"},
			envVars:         map[string]string{},
			validConfigFile: true,
			expectedConfig: &Config{
				ServerAddress:      DefaultServerAddress,
				BasicPath:          DefaultBasicPath,
				FileStoragePath:    DefaultFileStoragePath,
				DatabaseDSN:        "",
				PublicKeyPath:      DefaultPublicKeyPath,
				PrivateKeyPath:     DefaultPrivateKeyPath,
				EnableHTTPS:        DefaultEnableHTTPS,
				CertificatePath:    DefaultCertificatePath,
				CertificateKeyPath: DefaultCertificateKeyPath,
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
				"-cert", "custom_cert.pem",
				"-key", "custom_key.pem",
				"-c", "config.json",
			},
			envVars:         map[string]string{},
			validConfigFile: true,
			expectedConfig: &Config{
				ServerAddress:      "localhost:9090",
				BasicPath:          "http://localhost:9090",
				FileStoragePath:    "custom.txt",
				DatabaseDSN:        "postgres://user:pass@localhost:5432/db",
				PublicKeyPath:      "custom_public.pem",
				PrivateKeyPath:     "custom_private.pem",
				EnableHTTPS:        true,
				CertificatePath:    "custom_cert.pem",
				CertificateKeyPath: "custom_key.pem",
				Config:             "config.json",
			},
			expectedError: "",
		},
		{
			name: "environment variable overrides",
			args: []string{"cmd"},
			envVars: map[string]string{
				"SERVER_ADDRESS":       "localhost:9090",
				"BASE_URL":             "http://localhost:9090",
				"FILE_STORAGE_PATH":    "custom.txt",
				"DATABASE_DSN":         "postgres://user:pass@localhost:5432/db",
				"PUBLIC_KEY_PATH":      "custom_public.pem",
				"PRIVATE_KEY_PATH":     "custom_private.pem",
				"ENABLE_HTTPS":         "true",
				"CERTIFICATE_PATH":     "custom_cert.pem",
				"CERTIFICATE_KEY_PATH": "custom_key.pem",
				"CONFIG":               "config.json",
			},
			validConfigFile: true,
			expectedConfig: &Config{
				ServerAddress:      "localhost:9090",
				BasicPath:          "http://localhost:9090",
				FileStoragePath:    "custom.txt",
				DatabaseDSN:        "postgres://user:pass@localhost:5432/db",
				PublicKeyPath:      "custom_public.pem",
				PrivateKeyPath:     "custom_private.pem",
				EnableHTTPS:        true,
				CertificatePath:    "custom_cert.pem",
				CertificateKeyPath: "custom_key.pem",
				Config:             "config.json",
			},
			expectedError: "",
		},
		{
			name: "config file overrides",
			args: []string{"cmd"},
			envVars: map[string]string{
				"CONFIG": "config.json",
			},
			validConfigFile: true,
			expectedConfig: &Config{
				ServerAddress:      "localhost:9090",
				BasicPath:          "http://localhost:9090",
				FileStoragePath:    "custom.txt",
				DatabaseDSN:        "postgres://user:pass@localhost:5432/db",
				PublicKeyPath:      "custom_public.pem",
				PrivateKeyPath:     "custom_private.pem",
				EnableHTTPS:        true,
				CertificatePath:    "custom_cert.pem",
				CertificateKeyPath: "custom_key.pem",
				Config:             "config.json",
			},
			expectedError: "",
		},
		{
			name: "environment variable parsing error",
			args: []string{"cmd"},
			envVars: map[string]string{
				"ENABLE_HTTPS": "invalid",
			},
			validConfigFile: true,
			expectedConfig:  nil,
			expectedError:   "can not parse env",
		},
		{
			name: "config file parsing error",
			args: []string{"cmd"},
			envVars: map[string]string{
				"CONFIG": "custom_config.json",
			},
			validConfigFile: true,
			expectedConfig:  nil,
			expectedError:   "can not read config file",
		},
		{
			name: "config file parsing error",
			args: []string{"cmd"},
			envVars: map[string]string{
				"CONFIG": "invalid.json",
			},
			validConfigFile: false,
			expectedConfig:  nil,
			expectedError:   "can not unmarshal config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Clearenv()
			for key, value := range tt.envVars {
				err := os.Setenv(key, value)
				assert.NoError(t, err)
			}

			configFileName := "config.json"
			data, err := json.Marshal(tt.expectedConfig)
			assert.NoError(t, err)
			if !tt.validConfigFile {
				configFileName = "invalid.json"
				data = []byte("abc")
			}
			file, err := os.Create(configFileName)
			assert.NoError(t, err)
			_, err = file.Write(data)
			assert.NoError(t, err)
			err = file.Close()
			assert.NoError(t, err)
			defer func() {
				err := os.Remove(configFileName)
				assert.NoError(t, err)
			}()

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
