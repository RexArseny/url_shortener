package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-contrib/pprof"
	"go.uber.org/zap"
)

// NewServer create new server with new interactor, controller, middleware and router.
func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	urlRepository repository.Repository,
) (*http.Server, error) {
	interactor := usecases.NewInteractor(ctx, logger.Named("interactor"), cfg.BasicPath, urlRepository)
	controller := controllers.NewController(logger.Named("controller"), interactor)
	middleware, err := middlewares.NewMiddleware(
		cfg.PublicKeyPath,
		cfg.PrivateKeyPath,
		logger.Named("middleware"),
	)
	if err != nil {
		return nil, fmt.Errorf("can not init middleware: %w", err)
	}
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		return nil, fmt.Errorf("can not init router: %w", err)
	}

	pprof.Register(router)

	return &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}, nil
}
