package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
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
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
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
