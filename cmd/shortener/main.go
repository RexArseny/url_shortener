package main

import (
	"context"
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	mainLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Can not init logger: %s", err)
	}
	defer func() {
		if err := mainLogger.Sync(); err != nil {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		mainLogger.Fatal("Can not init config", zap.Error(err))
	}

	urlRepository, repositoryClose, err := repository.NewRepository(
		ctx,
		mainLogger.Named("repository"),
		cfg.FileStoragePath,
		cfg.DatabaseDSN,
	)
	if err != nil {
		mainLogger.Fatal("Can not init repository", zap.Error(err))
	}
	defer func() {
		if repositoryClose != nil {
			err = repositoryClose()
			if err != nil {
				mainLogger.Fatal("Can not close repository", zap.Error(err))
			}
		}
	}()

	interactor := usecases.NewInteractor(ctx, mainLogger.Named("interactor"), cfg.BasicPath, urlRepository)
	controller := controllers.NewController(mainLogger.Named("controller"), interactor)
	middleware, err := middlewares.NewMiddleware(
		cfg.PublicKeyPath,
		cfg.PrivateKeyPath,
		mainLogger.Named("middleware"),
	)
	if err != nil {
		mainLogger.Fatal("Can not init middleware", zap.Error(err))
	}
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		mainLogger.Fatal("Can not init router", zap.Error(err))
	}

	s := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}
	err = s.ListenAndServe()
	if err != nil {
		mainLogger.Fatal("Can not listen and serve", zap.Error(err))
	}
}
