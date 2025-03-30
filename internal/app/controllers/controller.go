package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ID is name of URL parameter.
const ID = "id"

// Controller is responsible for managing the network interactions of the service.
type Controller struct {
	logger        *zap.Logger
	trustedSubnet *net.IPNet
	interactor    usecases.Interactor
}

// NewController create new Controller.
func NewController(
	logger *zap.Logger,
	interactor usecases.Interactor,
	trustedSubnet *net.IPNet,
) Controller {
	return Controller{
		logger:        logger,
		trustedSubnet: trustedSubnet,
		interactor:    interactor,
	}
}

// CreateShortLink create new short URL from original URL.
// Input and output are in plain text format.
// Generate new JWT and put it in cookie if it is not presented.
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

// CreateShortLinkJSON create new short URL from original URL.
// Input and output are in JSON format.
// Generate new JWT and put it in cookie if it is not presented.
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
		c.logger.Error("Can not create short link from json", zap.Error(err))
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

// CreateShortLinkJSONBatch create new short URLs from original URLs.
// Input and output are in JSON format.
// Generate new JWT and put it in cookie if it is not presented.
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

// GetShortLink return original URL from short URL.
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

// PingDB ping and return the status of database.
func (c *Controller) PingDB(ctx *gin.Context) {
	err := c.interactor.PingDB(ctx)
	if err != nil {
		c.logger.Error("Can not ping db", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusText(http.StatusOK)})
}

// GetShortLinksOfUser return all short and original URLs of user if such exist and JWT is presented.
func (c *Controller) GetShortLinksOfUser(ctx *gin.Context) {
	newToken := ctx.GetBool(middlewares.AuthorizationNew)
	if newToken {
		ctx.JSON(http.StatusNoContent, gin.H{"error": http.StatusText(http.StatusNoContent)})
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

// DeleteURLs delete short URLs of user if such exist and JWT is presented.
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

// Stats return statistic of shortened urls and users in service.
func (c *Controller) Stats(ctx *gin.Context) {
	if c.trustedSubnet == nil || !c.trustedSubnet.Contains(net.ParseIP(ctx.GetHeader("X-Real-IP"))) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": http.StatusText(http.StatusForbidden)})
		return
	}

	stats, err := c.interactor.Stats(ctx)
	if err != nil {
		c.logger.Error("Can not get stats", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
