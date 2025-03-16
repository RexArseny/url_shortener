package app

import (
	"context"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		ServerAddress:  config.DefaultServerAddress,
		BasicPath:      config.DefaultBasicPath,
		PublicKeyPath:  "../../public.pem",
		PrivateKeyPath: "../../private.pem",
		EnableHTTPS:    false,
	}

	repo := repository.NewLinks()

	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	t.Run("successful server creation without HTTPS", func(t *testing.T) {
		server, err := NewServer(context.Background(), testLogger.Named("logger"), cfg, repo)
		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, server.Addr, cfg.ServerAddress)
	})

	t.Run("successful server creation with HTTPS", func(t *testing.T) {
		cfg.EnableHTTPS = true
		server, err := NewServer(context.Background(), testLogger.Named("logger"), cfg, repo)
		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.NotNil(t, server.TLSConfig)
	})
}
