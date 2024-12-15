package routers

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/gin-gonic/gin"
)

func NewRouter(
	cfg *config.Config,
	controller *controllers.Controller,
	middleware middlewares.Middleware) (*gin.Engine, error) {
	prefix, err := getURLPrefix(cfg)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.Logger(), middleware.Compressor())

	router.POST("/", controller.CreateShortLink)
	router.POST("/api/shorten", controller.CreateShortLinkJSON)
	router.POST("/api/shorten/batch", controller.CreateShortLinkJSONBatch)
	router.GET(fmt.Sprintf("%s/:%s", *prefix, controllers.ID), controller.GetShortLink)
	router.GET("/api/user/urls", controller.GetShortLinksOfUser)
	router.GET("/ping", controller.PingDB)

	return router, nil
}

func getURLPrefix(cfg *config.Config) (*string, error) {
	serverAddress, err := url.ParseRequestURI(cfg.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid server address: %w", err)
	}
	basicPath, err := url.ParseRequestURI(cfg.BasicPath)
	if err != nil {
		return nil, fmt.Errorf("invalid basic path: %w", err)
	}
	if serverAddress.String() != basicPath.Host {
		return nil, errors.New("server address does not correspond with basic path")
	}
	urlPrefix := path.Base(basicPath.Path)
	return &urlPrefix, nil
}
