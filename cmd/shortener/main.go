package main

import (
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/RexArseny/url_shortener/internal/app/utils"
	"go.uber.org/zap"
)

func main() {
	utils.InitLogger()
	defer func() {
		if err := utils.Logger.Sync(); err != nil {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		utils.Logger.Fatal("Can not init config", zap.Error(err))
	}

	interactor := usecases.NewInteractor(cfg.BasicPath)
	controller := controllers.NewController(interactor)
	router, err := routers.NewRouter(cfg, controller)
	if err != nil {
		utils.Logger.Fatal("Can not init router", zap.Error(err))
	}

	s := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}
	err = s.ListenAndServe()
	if err != nil {
		utils.Logger.Fatal("Can not listen and serve", zap.Error(err))
	}
}
