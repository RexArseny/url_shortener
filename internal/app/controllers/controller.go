package controllers

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const ID = "id"

type Controller struct {
	serverAddress string
	publicKey     crypto.PublicKey
	privateKey    crypto.PrivateKey
	logger        *zap.Logger
	interactor    usecases.Interactor
}

func NewController(
	serverAddress string,
	publicKeyPath string,
	privateKeyPath string,
	logger *zap.Logger,
	interactor usecases.Interactor,
) (*Controller, error) {
	publicKeyFile, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("can not open public.pem file: %w", err)
	}
	publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("can not parse public key: %w", err)
	}

	privateKeyFile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("can not open private.pem file: %w", err)
	}
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("can not parse private key: %w", err)
	}

	return &Controller{
		serverAddress: serverAddress,
		publicKey:     publicKey,
		privateKey:    privateKey,
		logger:        logger,
		interactor:    interactor,
	}, nil
}

func (c *Controller) CreateShortLink(ctx *gin.Context) {
	token, err := c.getJWT(ctx)
	if err != nil && !errors.Is(err, ErrNoJWT) {
		ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	err = c.setJWT(ctx, token)
	if err != nil {
		c.logger.Error("Can not set jwt", zap.Error(err))
		ctx.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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
	token, err := c.getJWT(ctx)
	if err != nil && !errors.Is(err, ErrNoJWT) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}
	err = c.setJWT(ctx, token)
	if err != nil {
		c.logger.Error("Can not set jwt", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
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
	token, err := c.getJWT(ctx)
	if err != nil && !errors.Is(err, ErrNoJWT) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}
	err = c.setJWT(ctx, token)
	if err != nil {
		c.logger.Error("Can not set jwt", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
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
	token, err := c.getJWT(ctx)
	if err != nil {
		if errors.Is(err, ErrNoJWT) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
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
