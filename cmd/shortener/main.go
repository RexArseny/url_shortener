package main

import (
	"fmt"

	"github.com/RexArseny/url_shortener/internal/app/args"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
)

func main() {
	interactor := usecases.NewInteractor()
	controller := controllers.NewController(interactor)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("/", controller.CreateShortLink)
	router.GET(fmt.Sprintf("/:%s", controllers.ID), controller.GetShortLink)

	router.Run(fmt.Sprintf("%s:%d", args.DefaultDomain, args.DefaultPort))
}
