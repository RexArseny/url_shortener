package routers

import (
	"fmt"
	"net/url"
	"path"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, controller controllers.Controller) (*gin.Engine, error) {
	prefix, err := getURLPrefix(cfg)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("/", controller.CreateShortLink)
	router.GET(fmt.Sprintf("%s/:%s", *prefix, controllers.ID), controller.GetShortLink)

	return router, nil
}

func getURLPrefix(cfg *config.Config) (*string, error) {
	serverAddress, err := url.ParseRequestURI(cfg.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid server address: %s", err)
	}
	basicPath, err := url.ParseRequestURI(cfg.BasicPath)
	if err != nil {
		return nil, fmt.Errorf("invalid basic path: %s", err)
	}
	if serverAddress.String() != basicPath.Host {
		return nil, fmt.Errorf("server address does not correspond with basic path")
	}
	urlPrefix := path.Base(basicPath.Path)
	return &urlPrefix, nil
}
