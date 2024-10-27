package main

import (
	"fmt"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	env "github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	cfg := config.Init()
	pflag.Parse()
	err := env.Parse(cfg)
	if err != nil {
		logrus.Fatalf("can not parse env: %s", err)
	}
	prefix, err := cfg.GetURLPrefix()
	if err != nil {
		logrus.Fatalf("invallid arguments: %s", err)
	}

	interactor := usecases.NewInteractor(cfg.BasicPath)
	controller := controllers.NewController(interactor)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("/", controller.CreateShortLink)
	router.GET(fmt.Sprintf("%s/:%s", *prefix, controllers.ID), controller.GetShortLink)

	router.Run(cfg.ServerAddress)
}
