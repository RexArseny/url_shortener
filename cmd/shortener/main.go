package main

import (
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Can not init logger: %s", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		logger.Fatal("Can not init config", zap.Error(err))
	}

	interactor := usecases.NewInteractor(cfg.BasicPath)
	controller := controllers.NewController(logger.Named("controller"), interactor)
	middleware := middlewares.NewMiddleware(logger.Named("middleware"))
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		logger.Fatal("Can not init router", zap.Error(err))
	}

	s := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}
	err = s.ListenAndServe()
	if err != nil {
		logger.Fatal("Can not listen and serve", zap.Error(err))
	}
}
