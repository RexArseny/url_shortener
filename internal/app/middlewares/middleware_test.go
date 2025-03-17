package middlewares

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewMiddleware(t *testing.T) {
	t.Run("successful middleware creation", func(t *testing.T) {
		publicKeyPath := "../../../public.pem"
		privateKeyPath := "../../../private.pem"

		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		middleware, err := NewMiddleware(publicKeyPath, privateKeyPath, testLogger)
		assert.NoError(t, err)
		assert.NotNil(t, middleware)
		assert.NotNil(t, middleware.publicKey)
		assert.NotNil(t, middleware.privateKey)
		assert.Equal(t, testLogger, middleware.logger)
	})

	t.Run("failed to read public key file", func(t *testing.T) {
		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		_, err = NewMiddleware("nonexistent_public.pem", "nonexistent_private.pem", testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not open public.pem file")
	})

	t.Run("failed to parse public key", func(t *testing.T) {
		publicKeyPath := "test_public.pem"
		defer func() {
			err := os.Remove(publicKeyPath)
			assert.NoError(t, err)
		}()

		err := os.WriteFile(publicKeyPath, []byte("invalid public key"), 0644)
		assert.NoError(t, err)

		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		_, err = NewMiddleware(publicKeyPath, "nonexistent_private.pem", testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not parse public key")
	})

	t.Run("failed to read private key file", func(t *testing.T) {
		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		_, err = NewMiddleware("../../../public.pem", "nonexistent_private.pem", testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not open private.pem file")
	})

	t.Run("failed to parse private key", func(t *testing.T) {
		privateKeyPath := "test_private.pem"
		defer func() {
			err := os.Remove(privateKeyPath)
			assert.NoError(t, err)
		}()

		err := os.WriteFile(privateKeyPath, []byte("invalid private key"), 0644)
		assert.NoError(t, err)

		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		_, err = NewMiddleware("../../../public.pem", privateKeyPath, testLogger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not parse private key")
	})
}

func TestLogger(t *testing.T) {
	t.Run("successful logging", func(t *testing.T) {
		core, recordedLogs := observer.New(zap.InfoLevel)
		logger := zap.New(core)

		middleware := &Middleware{
			logger: logger,
		}

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/test?query=param", nil)

		middleware.Logger()(ctx)

		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Write([]byte("test response"))

		assert.Equal(t, 1, recordedLogs.Len())
		logEntry := recordedLogs.All()[0]
		assert.Equal(t, "Request", logEntry.Message)
		assert.Equal(t, int64(http.StatusOK), logEntry.ContextMap()["code"].(int64))
		assert.Equal(t, http.MethodGet, logEntry.ContextMap()["method"].(string))
		assert.Equal(t, "/test?query=param", logEntry.ContextMap()["path"].(string))
	})
}

func TestGzipWriter(t *testing.T) {
	t.Run("write string success", func(t *testing.T) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		gw := &gzipWriter{
			ResponseWriter: ctx.Writer,
			writer:         gzWriter,
		}

		n, err := gw.WriteString("test")
		assert.NoError(t, err)
		assert.Equal(t, 4, n)

		err = gzWriter.Close()
		assert.NoError(t, err)
	})

	t.Run("write string failure", func(t *testing.T) {
		gzWriter := gzip.NewWriter(&bytes.Buffer{})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		gw := &gzipWriter{
			ResponseWriter: ctx.Writer,
			writer:         gzWriter,
		}

		err := gzWriter.Close()
		assert.NoError(t, err)

		_, err = gw.WriteString("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not write string")
	})

	t.Run("write success", func(t *testing.T) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		gw := &gzipWriter{
			ResponseWriter: ctx.Writer,
			writer:         gzWriter,
		}

		n, err := gw.Write([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, 4, n)

		err = gzWriter.Close()
		assert.NoError(t, err)
	})

	t.Run("write failure", func(t *testing.T) {
		gzWriter := gzip.NewWriter(&bytes.Buffer{})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		gw := &gzipWriter{
			ResponseWriter: ctx.Writer,
			writer:         gzWriter,
		}

		err := gzWriter.Close()
		assert.NoError(t, err)

		_, err = gw.Write([]byte("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not write string")
	})
}

func TestCompressor(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := &Middleware{
		logger: logger,
	}

	t.Run("decompress request body", func(t *testing.T) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err := gzWriter.Write([]byte(`{"key":"value"}`))
		assert.NoError(t, err)
		gzWriter.Close()

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		resp := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(resp)
		ctx.Request = req

		middleware.Compressor()(ctx)

		body, err := ctx.GetRawData()
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, string(body))
	})

	t.Run("compress response body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(resp)
		ctx.Request = req

		middleware.Compressor()(ctx)

		ctx.String(http.StatusOK, "test")

		assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))

		gzReader, err := gzip.NewReader(resp.Body)
		assert.NoError(t, err)
		defer gzReader.Close()

		var decompressed bytes.Buffer
		_, err = decompressed.ReadFrom(gzReader)
		assert.NoError(t, err)
	})

	t.Run("invalid gzip request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid gzip data"))
		req.Header.Set("Content-Encoding", "gzip")

		resp := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(resp)
		ctx.Request = req

		middleware.Compressor()(ctx)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), "Internal Server Error")
	})
}

func TestAuth(t *testing.T) {
	core, recordedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	publicKeyFile, err := os.ReadFile("../../../public.pem")
	assert.NoError(t, err)
	publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyFile)
	assert.NoError(t, err)

	privateKeyFile, err := os.ReadFile("../../../private.pem")
	assert.NoError(t, err)
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
	assert.NoError(t, err)

	middleware := &Middleware{
		publicKey:  publicKey,
		privateKey: privateKey,
		logger:     logger,
	}

	t.Run("no JWT in cookie", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		middleware.Auth()(ctx)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		cookies := ctx.Writer.Header().Get("Set-Cookie")
		assert.Contains(t, cookies, Authorization)

		claims, exists := ctx.Get(Authorization)
		assert.True(t, exists)
		assert.IsType(t, &JWT{}, claims)

		isNew, exists := ctx.Get(AuthorizationNew)
		assert.True(t, exists)
		assert.True(t, isNew.(bool))
	})

	t.Run("valid JWT in cookie", func(t *testing.T) {
		userID := uuid.New()
		claims := &JWT{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "url_shortener",
				Subject:   userID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.New().String(),
			},
			UserID: userID,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		tokenString, err := token.SignedString(middleware.privateKey)
		assert.NoError(t, err)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx.Request.AddCookie(&http.Cookie{
			Name:  Authorization,
			Value: tokenString,
		})

		middleware.Auth()(ctx)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		claimsCtx, exists := ctx.Get(Authorization)
		assert.True(t, exists)
		assert.IsType(t, &JWT{}, claimsCtx)

		isNew, exists := ctx.Get(AuthorizationNew)
		assert.True(t, exists)
		assert.False(t, isNew.(bool))
	})

	t.Run("invalid JWT in cookie", func(t *testing.T) {
		invalidTokenString := "invalid.token.string"

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx.Request.AddCookie(&http.Cookie{
			Name:  Authorization,
			Value: invalidTokenString,
		})

		middleware.Auth()(ctx)

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())

		assert.Equal(t, 1, recordedLogs.Len())
		logEntry := recordedLogs.All()[0]
		assert.Equal(t, "Can not parse jwt", logEntry.Message)
	})

	t.Run("JWT signature mismatch", func(t *testing.T) {
		userID := uuid.New()
		claims := &JWT{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "url_shortener",
				Subject:   userID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.New().String(),
			},
			UserID: userID,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("invalid-key"))
		assert.NoError(t, err)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx.Request.AddCookie(&http.Cookie{
			Name:  Authorization,
			Value: tokenString,
		})

		middleware.Auth()(ctx)

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())

		logEntry := recordedLogs.All()[0]
		assert.Equal(t, "Can not parse jwt", logEntry.Message)
	})
}
