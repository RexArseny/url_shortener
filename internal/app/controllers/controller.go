package controllers

import (
	"io"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/RexArseny/url_shortener/internal/app/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ID = "id"

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
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	result, err := c.interactor.CreateShortLink(string(data))
	if err != nil {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if result == nil || *result == "" {
		utils.Logger.Error("Short link is empty", zap.Any("request", ctx.Request))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/plain")
	ctx.String(http.StatusCreated, *result)
}

func (c *Controller) GetShortLink(ctx *gin.Context) {
	data := ctx.Param(ID)

	result, err := c.interactor.GetShortLink(data)
	if err != nil {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if result == nil || *result == "" {
		utils.Logger.Error("Short link is empty", zap.Any("request", ctx.Request))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, *result)
}
