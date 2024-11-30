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

	var urlRepository repository.Repository
	switch {
	case cfg.DatabaseDSN != "":
		dbRepository, err := repository.NewDBRepository(ctx, mainLogger.Named("repository"), cfg.DatabaseDSN)
		if err != nil {
			mainLogger.Fatal("Can not init db repository", zap.Error(err))
		}
		defer dbRepository.Close()
		urlRepository = dbRepository
	case cfg.FileStoragePath != "":
		linksWithFile, err := repository.NewLinksWithFile(cfg.FileStoragePath)
		if err != nil {
			mainLogger.Fatal("Can not init file repository", zap.Error(err))
		}
		defer func() {
			if err := linksWithFile.Close(); err != nil {
				mainLogger.Fatal("Can not close file", zap.Error(err))
			}
		}()
		urlRepository = linksWithFile
	default:
		urlRepository = repository.NewLinks()
	}

	interactor := usecases.NewInteractor(cfg.BasicPath, urlRepository)
	controller := controllers.NewController(mainLogger.Named("controller"), interactor)
	middleware := middlewares.NewMiddleware(mainLogger.Named("middleware"))
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
