package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ID = "id"

type Controller struct {
	logger     *zap.Logger
	interactor usecases.Interactor
}

func NewController(logger *zap.Logger, interactor usecases.Interactor) Controller {
	return Controller{
		logger:     logger,
		interactor: interactor,
	}
}

func (c *Controller) CreateShortLink(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	result, err := c.interactor.CreateShortLink(ctx, string(data), token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLUniqueViolation) && result != nil {
			ctx.Writer.Header().Set("Content-Type", "text/plain")
			ctx.String(http.StatusConflict, *result)
			return
		}
		if errors.Is(err, repository.ErrInvalidURL) {
			ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		c.logger.Error("Can not create short link", zap.Error(err))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if result == nil || *result == "" {
		c.logger.Error("Short link is empty", zap.Any("request", ctx.Request))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Writer.Header().Set("Content-Type", "text/plain")
	ctx.String(http.StatusCreated, *result)
}

func (c *Controller) CreateShortLinkJSON(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request models.ShortenRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	result, err := c.interactor.CreateShortLink(ctx, request.URL, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLUniqueViolation) && result != nil {
			ctx.JSON(http.StatusConflict, models.ShortenResponse{
				Result: *result,
			})
			return
		}
		if errors.Is(err, repository.ErrInvalidURL) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
			return
		}
		c.logger.Error("Can not create short link", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if result == nil || *result == "" {
		c.logger.Error("Short link is empty", zap.Any("request", ctx.Request))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusCreated, models.ShortenResponse{
		Result: *result,
	})
}

func (c *Controller) CreateShortLinkJSONBatch(ctx *gin.Context) {
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request []models.ShortenBatchRequest
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	result, err := c.interactor.CreateShortLinks(ctx, request, token.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLUniqueViolation) && result != nil {
			ctx.JSON(http.StatusConflict, result)
			return
		}
		if errors.Is(err, repository.ErrInvalidURL) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
			return
		}
		c.logger.Error("Can not create short links", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (c *Controller) GetShortLink(ctx *gin.Context) {
	data := ctx.Param(ID)

	result, err := c.interactor.GetShortLink(ctx, data)
	if err != nil {
		if errors.Is(err, repository.ErrURLIsDeleted) {
			ctx.String(http.StatusGone, http.StatusText(http.StatusGone))
			return
		}
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if result == nil || *result == "" {
		c.logger.Error("Short link is empty", zap.Any("request", ctx.Request))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, *result)
}

func (c *Controller) PingDB(ctx *gin.Context) {
	err := c.interactor.PingDB(ctx)
	if err != nil {
		c.logger.Error("Can not ping db", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusText(http.StatusOK)})
}

func (c *Controller) GetShortLinksOfUser(ctx *gin.Context) {
	newToken := ctx.GetBool(middlewares.AuthorizationNew)
	if newToken {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	result, err := c.interactor.GetShortLinksOfUser(ctx, token.UserID)
	if err != nil {
		c.logger.Error("Can not get short links of user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if len(result) == 0 {
		ctx.JSON(http.StatusNoContent, gin.H{"error": http.StatusText(http.StatusNoContent)})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controller) DeleteURLs(ctx *gin.Context) {
	newToken := ctx.GetBool(middlewares.AuthorizationNew)
	if newToken {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	tokenValue, ok := ctx.Get(middlewares.Authorization)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	token, ok := tokenValue.(*middlewares.JWT)
	if !ok {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	var request []string
	err = json.Unmarshal(data, &request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}

	err = c.interactor.DeleteURLs(ctx, request, token.UserID)
	if err != nil {
		c.logger.Error("Can not delete urls", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"status": http.StatusText(http.StatusAccepted)})
}
