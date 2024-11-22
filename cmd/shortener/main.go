package main

import (
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"go.uber.org/zap"
)

func main() {
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

	var repository models.Repository
	if cfg.FileStoragePath != "" {
		linksWithFile, err := models.NewLinksWithFile(cfg.FileStoragePath)
		if err != nil {
			mainLogger.Fatal("Can not init repository", zap.Error(err))
		}
		defer func() {
			if err := linksWithFile.Close(); err != nil {
				mainLogger.Fatal("Can not close file", zap.Error(err))
			}
		}()
		repository = linksWithFile
	} else {
		repository = models.NewLinks()
	}

	interactor := usecases.NewInteractor(cfg.BasicPath, repository)
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
