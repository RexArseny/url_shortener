package middlewares

import (
	"compress/gzip"
	"crypto"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	Authorization    = "Authorization"
	AuthorizationNew = "AuthorizationNew"
	maxAge           = 900
)

var ErrNoJWT = errors.New("no jwt in cookie")

type Middleware struct {
	publicKey  crypto.PublicKey
	privateKey crypto.PrivateKey
	logger     *zap.Logger
}

func NewMiddleware(publicKeyPath string, privateKeyPath string, logger *zap.Logger) (*Middleware, error) {
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

	return &Middleware{
		publicKey:  publicKey,
		privateKey: privateKey,
		logger:     logger,
	}, nil
}

func (m *Middleware) Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		m.logger.Info("Request",
			zap.Int("code", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("size", ctx.Writer.Size()))
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	n, err := g.writer.Write([]byte(s))
	if err != nil {
		return 0, fmt.Errorf("can not write string: %w", err)
	}
	return n, nil
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	n, err := g.writer.Write(data)
	if err != nil {
		return 0, fmt.Errorf("can not write string: %w", err)
	}
	return n, nil
}

func (m *Middleware) Compressor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				m.logger.Error("Can not create gzip reader", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
				ctx.Abort()
				return
			}
			defer func() {
				err = reader.Close()
				if err != nil {
					m.logger.Error("Can not close gzip reader", zap.Error(err))
				}
			}()
			ctx.Request.Body = reader
		}

		if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") &&
			(strings.Contains(ctx.GetHeader("Content-Type"), "application/json") ||
				strings.Contains(ctx.GetHeader("Content-Type"), "text/html")) {
			writer := gzip.NewWriter(ctx.Writer)
			defer func() {
				err := writer.Close()
				if err != nil {
					m.logger.Error("Can not close gzip writer", zap.Error(err))
				}
			}()
			ctx.Writer.Header().Set("Content-Encoding", "gzip")
			ctx.Writer = &gzipWriter{ctx.Writer, writer}
		}

		ctx.Next()
	}
}

type JWT struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func (m *Middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie(Authorization)

		if err != nil || tokenString == "" {
			userID := uuid.New()
			claims := &JWT{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "url_shortener",
					Subject:   userID.String(),
					Audience:  jwt.ClaimStrings{},
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
					NotBefore: jwt.NewNumericDate(time.Now()),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ID:        uuid.New().String(),
				},
				UserID: userID,
			}

			token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

			tokenString, err = token.SignedString(m.privateKey)
			if err != nil {
				m.logger.Error("Can not sign token", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
				ctx.Abort()
				return
			}

			ctx.SetCookie(
				Authorization,
				tokenString,
				maxAge,
				"/",
				"",
				false,
				false,
			)
			ctx.Set(Authorization, claims)
			ctx.Set(AuthorizationNew, true)

			ctx.Next()

			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&JWT{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodEdDSA {
					return nil, errors.New("jwt signature mismatch")
				}
				return m.publicKey, nil
			},
		)
		if err != nil {
			m.logger.Error("Can not parse jwt", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(*JWT)
		if !ok {
			m.logger.Error("Token is not jwt format")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			ctx.Abort()
			return
		}

		ctx.SetCookie(
			Authorization,
			tokenString,
			maxAge,
			"/",
			"",
			false,
			false,
		)
		ctx.Set(Authorization, claims)
		ctx.Set(AuthorizationNew, false)

		ctx.Next()
	}
}
