package controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const ID = "id"

var errorService = fmt.Errorf("service error")

type Controller struct {
	interactor usecases.Interactor
}

func NewController(interactor usecases.Interactor) Controller {
	return Controller{
		interactor: interactor,
	}
}

func (c *Controller) CreateShortLink(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logrus.Errorf("can not read body %v; request: %v", ctx.Request.Body, ctx.Request)
		ctx.String(http.StatusBadRequest, errorService.Error())
		return
	}

	result, err := c.interactor.CreateShortLink(string(data))
	if err != nil {
		logrus.Errorf("can not create short link %s; request: %v", err, ctx.Request)
		ctx.String(http.StatusBadRequest, errorService.Error())
		return
	}

	if result == nil || *result == "" {
		logrus.Errorf("short link is empty; request: %v", ctx.Request)
		ctx.String(http.StatusBadRequest, errorService.Error())
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/plain")
	ctx.String(http.StatusCreated, *result)
}

func (c *Controller) GetShortLink(ctx *gin.Context) {
	data := ctx.Param(ID)

	result, err := c.interactor.GetShortLink(data)
	if err != nil {
		logrus.Errorf("can not get short link %s; request: %v", err, ctx.Request)
		ctx.String(http.StatusBadRequest, errorService.Error())
		return
	}

	if result == nil || *result == "" {
		logrus.Errorf("short link is empty; request: %v", ctx.Request)
		ctx.String(http.StatusBadRequest, errorService.Error())
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, *result)
}
