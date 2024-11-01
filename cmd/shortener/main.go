package main

import (
	"net/http"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		logrus.Fatalf("can not init config: %s", err)
	}

	interactor := usecases.NewInteractor(cfg.BasicPath)
	controller := controllers.NewController(interactor)
	router, err := routers.NewRouter(cfg, controller)
	if err != nil {
		logrus.Fatalf("can not init router: %s", err)
	}

	s := &http.Server{
		Addr:           cfg.ServerAddress,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = s.ListenAndServe()
	if err != nil {
		logrus.Fatalf("can not listen and serve: %s", err)
	}
}
