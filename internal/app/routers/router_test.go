package routers

import (
	"context"
	"os"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	cfg, err := config.Init()
	assert.NoError(t, err)
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	file, err := os.CreateTemp("./", "*.test")
	assert.NoError(t, err)
	urlRepository, err := repository.NewLinksWithFile(file.Name())
	assert.NoError(t, err)

	defer func() {
		err = urlRepository.Close()
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)
		err = os.Remove(file.Name())
		assert.NoError(t, err)
	}()

	interactor := usecases.NewInteractor(
		context.Background(),
		testLogger.Named("interactor"),
		cfg.BasicPath,
		urlRepository,
	)
	conntroller := controllers.NewController(testLogger.Named("controller"), interactor, nil)

	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)

	router, err := NewRouter(cfg, conntroller, middleware)
	assert.NoError(t, err)
	assert.NotEmpty(t, router)

	cfg.ServerAddress = "abc"
	router, err = NewRouter(cfg, conntroller, middleware)
	assert.Error(t, err)
	assert.Empty(t, router)

	cfg.ServerAddress = config.DefaultServerAddress
	cfg.BasicPath = "abc"
	router, err = NewRouter(cfg, conntroller, middleware)
	assert.Error(t, err)
	assert.Empty(t, router)

	cfg.BasicPath = "http://localhost:8081"
	router, err = NewRouter(cfg, conntroller, middleware)
	assert.Error(t, err)
	assert.Empty(t, router)
}
