//nolint:dupl // tests of handlers
package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	type want struct {
		contenType  string
		stastusCode int
		response    bool
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "empty url",
			request: "",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
		{
			name:    "invalid url",
			request: "abc",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
		{
			name:    "valid url",
			request: "https://ya.ru",
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "text/plain",
				response:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			file, err := os.CreateTemp("./", "*.test")
			assert.NoError(t, err)
			urlRepository, err := repository.NewLinksWithFile(file.Name())
			assert.NoError(t, err)

			defer func() {
				err = urlRepository.Close()
				assert.NoError(t, err)
				err = file.Close()
				assert.NoError(t, err)
				err = os.Remove(file.Name())
				assert.NoError(t, err)
			}()

			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				urlRepository,
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.CreateShortLink(ctx)

			result := w.Result()

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.contenType, result.Header.Get("Content-Type"))

			resultBody, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			assert.NoError(t, err)

			if tt.want.response {
				assert.Contains(t, string(resultBody), "http")
				return
			}

			assert.NotContains(t, string(resultBody), "http")
		})
	}
}

func TestCreateShortLinkJSON(t *testing.T) {
	type want struct {
		response    map[string]interface{}
		contenType  string
		stastusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "empty body",
			request: "",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid body",
			request: "abc",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid url",
			request: `{"url":"abc"}`,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "valid url",
			request: `{"url":"https://ya.ru"}`,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"result": "http"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.request))

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.CreateShortLinkJSON(ctx)

			result := w.Result()

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.contenType, result.Header.Get("Content-Type"))

			resultBody, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			assert.NoError(t, err)

			var response map[string]interface{}
			err = json.Unmarshal(resultBody, &response)
			assert.NoError(t, err)

			for key, val := range tt.want.response {
				assert.Contains(t, response[key], val)
			}
		})
	}
}

func TestCreateShortLinkJSONBatch(t *testing.T) {
	type want struct {
		response    map[string]interface{}
		contenType  string
		stastusCode int
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "nil body",
			request: "",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "empty body",
			request: `[{}]`,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid url",
			request: `[{"correlation_id":"1","original_url":"abc"}]`,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "valid url",
			request: `[{"correlation_id":"1","original_url":"https://ya.ru"}]`,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"short_url": "http"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.request))

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.CreateShortLinkJSONBatch(ctx)

			result := w.Result()

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.contenType, result.Header.Get("Content-Type"))

			resultBody, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			assert.NoError(t, err)

			for _, val := range tt.want.response {
				assert.Contains(t, string(resultBody), val)
			}
		})
	}
}

func TestGetShortLink(t *testing.T) {
	type input struct {
		path  string
		valid bool
	}
	type want struct {
		location    string
		stastusCode int
	}
	tests := []struct {
		name    string
		request input
		want    want
	}{
		{
			name: "empty id",
			request: input{
				valid: false,
				path:  "",
			},
			want: want{
				stastusCode: http.StatusBadRequest,
				location:    "",
			},
		},
		{
			name: "invalid id",
			request: input{
				valid: false,
				path:  "abc",
			},
			want: want{
				stastusCode: http.StatusBadRequest,
				location:    "",
			},
		},
		{
			name: "valid id",
			request: input{
				valid: true,
				path:  "",
			},
			want: want{
				stastusCode: http.StatusTemporaryRedirect,
				location:    "https://ya.ru",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.CreateShortLink(ctx)

			result := w.Result()

			assert.Equal(t, http.StatusCreated, result.StatusCode)
			assert.Equal(t, "text/plain", result.Header.Get("Content-Type"))

			resultBody, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			assert.NoError(t, err)

			parsedURL, err := url.ParseRequestURI(string(resultBody))
			assert.NoError(t, err)
			assert.NotEmpty(t, parsedURL)

			var id string
			if tt.request.valid {
				id = path.Base(parsedURL.Path)
			} else {
				id = tt.request.path
			}

			w = httptest.NewRecorder()
			ctx, _ = gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/:"+ID, http.NoBody)
			ctx.Params = []gin.Param{
				{
					Key:   ID,
					Value: id,
				},
			}

			conntroller.GetShortLink(ctx)

			result = w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}

func TestPingDB(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "valid data",
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.PingDB(ctx)

			result := w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)
		})
	}
}

func TestGetShortLinksOfUser(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "valid data",
			want: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			privateKeyFile, err := os.ReadFile("../../../private.pem")
			assert.NoError(t, err)
			privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
			assert.NoError(t, err)
			userID := uuid.New()
			claims := &middlewares.JWT{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "url_shortener",
					Subject:   userID.String(),
					Audience:  jwt.ClaimStrings{},
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 900)),
					NotBefore: jwt.NewNumericDate(time.Now()),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ID:        uuid.New().String(),
				},
				UserID: userID,
			}
			token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
			tokenString, err := token.SignedString(privateKey)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
			ctx.Request.AddCookie(&http.Cookie{
				Name:   middlewares.Authorization,
				Value:  tokenString,
				Path:   "/",
				Domain: "",
			})

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.GetShortLinksOfUser(ctx)

			result := w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)
		})
	}
}

func TestDeleteURLs(t *testing.T) {
	tests := []struct {
		name string
		file bool
		want int
	}{
		{
			name: "valid url in memory",
			file: false,
			want: http.StatusAccepted,
		},
		{
			name: "valid url in file",
			file: true,
			want: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)

			var urlRepository repository.Repository
			var file *os.File
			if tt.file {
				file, err = os.CreateTemp("./", "*.test")
				assert.NoError(t, err)
				urlRepository, err = repository.NewLinksWithFile(file.Name())
				assert.NoError(t, err)
			} else {
				urlRepository = repository.NewLinks()
			}

			defer func() {
				if tt.file {
					linksWithFile, ok := urlRepository.(*repository.LinksWithFile)
					if !ok {
						return
					}
					err = linksWithFile.Close()
					assert.NoError(t, err)
					err = file.Close()
					assert.NoError(t, err)
					err = os.Remove(file.Name())
					assert.NoError(t, err)
				}
			}()

			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				urlRepository,
			)
			conntroller := NewController(testLogger.Named("controller"), interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/shorten/batch",
				strings.NewReader(`[{"correlation_id":"1","original_url":"https://ya.ru"}]`))

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.CreateShortLinkJSONBatch(ctx)

			result := w.Result()

			var tokenString string
			for _, cookie := range result.Cookies() {
				if cookie.Name == middlewares.Authorization {
					tokenString = cookie.Value
				}
			}

			resultBody, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			assert.NoError(t, err)

			var data []models.ShortenBatchResponse
			err = json.Unmarshal(resultBody, &data)
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			parsedURL, err := url.ParseRequestURI(data[0].ShortURL)
			assert.NoError(t, err)
			assert.NotEmpty(t, parsedURL)

			w = httptest.NewRecorder()
			ctx, _ = gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				http.MethodDelete,
				"/api/user/urls",
				strings.NewReader(`["`+path.Base(parsedURL.Path)+`"]`))
			ctx.Request.AddCookie(&http.Cookie{
				Name:   middlewares.Authorization,
				Value:  tokenString,
				Path:   "/",
				Domain: "",
			})

			auth(ctx)

			conntroller.DeleteURLs(ctx)

			result = w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)
		})
	}
}
