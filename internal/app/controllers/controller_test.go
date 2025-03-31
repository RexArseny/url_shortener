//nolint:dupl // tests of handlers
package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net"
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
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)
	auth := middleware.Auth()

	type want struct {
		contenType  string
		stastusCode int
		response    bool
	}
	tests := []struct {
		name      string
		request   string
		auth      gin.HandlerFunc
		duplicate bool
		want      want
	}{
		{
			name:      "empty url",
			request:   "",
			auth:      auth,
			duplicate: false,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
		{
			name:      "invalid url",
			request:   "abc",
			auth:      auth,
			duplicate: false,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
		{
			name:      "valid url",
			request:   "https://ya.ru",
			auth:      auth,
			duplicate: false,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "text/plain",
				response:    true,
			},
		},
		{
			name:      "duplicate url",
			request:   "https://ya.ru",
			auth:      auth,
			duplicate: true,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "text/plain",
				response:    true,
			},
		},
		{
			name:      "no auth",
			request:   "https://ya.ru",
			auth:      func(ctx *gin.Context) {},
			duplicate: false,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
		{
			name:      "invlaid token",
			request:   "https://ya.ru",
			auth:      func(ctx *gin.Context) { ctx.Set(middlewares.Authorization, "abc") },
			duplicate: false,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "text/plain; charset=utf-8",
				response:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
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
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))

			tt.auth(ctx)

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

			if tt.duplicate {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))

				tt.auth(ctx)

				conntroller.CreateShortLink(ctx)

				result := w.Result()

				assert.Equal(t, http.StatusConflict, result.StatusCode)
				assert.Equal(t, tt.want.contenType, result.Header.Get("Content-Type"))

				err = result.Body.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateShortLinkJSON(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)
	auth := middleware.Auth()

	type want struct {
		response    map[string]interface{}
		contenType  string
		stastusCode int
	}
	tests := []struct {
		name    string
		request string
		auth    gin.HandlerFunc
		want    want
	}{
		{
			name:    "empty body",
			request: "",
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid body",
			request: "abc",
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid url",
			request: `{"url":"abc"}`,
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "valid url",
			request: `{"url":"https://ya.ru"}`,
			auth:    auth,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"result": "http"},
			},
		},
		{
			name:    "no auth",
			request: `{"url":"https://ya.ru"}`,
			auth:    func(ctx *gin.Context) {},
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid token",
			request: `{"url":"https://ya.ru"}`,
			auth:    func(ctx *gin.Context) { ctx.Set(middlewares.Authorization, "abc") },
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.request))

			tt.auth(ctx)

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
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)
	auth := middleware.Auth()

	type want struct {
		response    map[string]interface{}
		contenType  string
		stastusCode int
	}
	tests := []struct {
		name    string
		request string
		auth    gin.HandlerFunc
		want    want
	}{
		{
			name:    "nil body",
			request: "",
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "empty body",
			request: `[{}]`,
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "invalid url",
			request: `[{"correlation_id":"1","original_url":"abc"}]`,
			auth:    auth,
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "valid url",
			request: `[{"correlation_id":"1","original_url":"https://ya.ru"}]`,
			auth:    auth,
			want: want{
				stastusCode: http.StatusCreated,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"short_url": "http"},
			},
		},
		{
			name:    "no auth",
			request: `[{"correlation_id":"1","original_url":"https://ya.ru"}]`,
			auth:    func(ctx *gin.Context) {},
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
		{
			name:    "no token",
			request: `[{"correlation_id":"1","original_url":"https://ya.ru"}]`,
			auth:    func(ctx *gin.Context) { ctx.Set(middlewares.Authorization, "abc") },
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "application/json; charset=utf-8",
				response:    map[string]interface{}{"error": "Bad Request"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.request))

			tt.auth(ctx)

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
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

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
			want: http.StatusOK,
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
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

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
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)
	auth := middleware.Auth()

	tests := []struct {
		name  string
		token bool
		data  bool
		auth  gin.HandlerFunc
		want  int
	}{
		{
			name:  "valid data",
			token: true,
			data:  false,
			auth:  auth,
			want:  http.StatusNoContent,
		},
		{
			name:  "no token",
			token: false,
			data:  false,
			auth:  auth,
			want:  http.StatusNoContent,
		},
		{
			name:  "with data",
			token: true,
			data:  true,
			auth:  auth,
			want:  http.StatusOK,
		},
		{
			name:  "no auth",
			token: true,
			data:  true,
			auth:  func(ctx *gin.Context) {},
			want:  http.StatusNoContent,
		},
		{
			name:  "invalid token",
			token: true,
			data:  true,
			auth:  func(ctx *gin.Context) { ctx.Set(middlewares.Authorization, "abc") },
			want:  http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

			var tokenString string
			if tt.data {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(
					http.MethodPost,
					"/api/shorten/batch",
					strings.NewReader(`[{"correlation_id":"1","original_url":"https://ya.ru"}]`))

				tt.auth(ctx)

				conntroller.CreateShortLinkJSONBatch(ctx)

				result := w.Result()

				for _, cookie := range result.Cookies() {
					if cookie.Name == middlewares.Authorization {
						tokenString = cookie.Value
					}
				}

				err = result.Body.Close()
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
			if tt.token {
				ctx.Request.AddCookie(&http.Cookie{
					Name:   middlewares.Authorization,
					Value:  tokenString,
					Path:   "/",
					Domain: "",
				})
			}

			tt.auth(ctx)

			conntroller.GetShortLinksOfUser(ctx)

			result := w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)
		})
	}
}

func TestDeleteURLs(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)
	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	assert.NoError(t, err)
	auth := middleware.Auth()

	tests := []struct {
		name    string
		file    bool
		token   bool
		request string
		auth    gin.HandlerFunc
		want    int
	}{
		{
			name:    "valid url in memory",
			file:    false,
			token:   true,
			request: "",
			auth:    auth,
			want:    http.StatusAccepted,
		},
		{
			name:    "valid url in file",
			file:    true,
			token:   true,
			request: "",
			auth:    auth,
			want:    http.StatusAccepted,
		},
		{
			name:    "invalid data",
			file:    false,
			token:   true,
			request: `{"id":"test"}`,
			auth:    auth,
			want:    http.StatusBadRequest,
		},
		{
			name:    "no token",
			file:    false,
			token:   false,
			request: "",
			auth:    auth,
			want:    http.StatusUnauthorized,
		},
		{
			name:    "no auth",
			file:    false,
			token:   false,
			request: "",
			auth:    func(ctx *gin.Context) {},
			want:    http.StatusUnauthorized,
		},
		{
			name:    "invalid token",
			file:    false,
			token:   false,
			request: "",
			auth:    func(ctx *gin.Context) { ctx.Set(middlewares.Authorization, "abc") },
			want:    http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath: config.DefaultBasicPath,
			}

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
			conntroller := NewController(testLogger.Named("controller"), interactor, nil)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/shorten/batch",
				strings.NewReader(`[{"correlation_id":"1","original_url":"https://ya.ru"}]`))

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

			requestData := `["` + path.Base(parsedURL.Path) + `"]`
			if tt.request != "" {
				requestData = tt.request
			}
			w = httptest.NewRecorder()
			ctx, _ = gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(
				http.MethodDelete,
				"/api/user/urls",
				strings.NewReader(requestData))
			if tt.token {
				ctx.Request.AddCookie(&http.Cookie{
					Name:   middlewares.Authorization,
					Value:  tokenString,
					Path:   "/",
					Domain: "",
				})
			}

			tt.auth(ctx)

			conntroller.DeleteURLs(ctx)

			result = w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)

			if result.StatusCode == http.StatusAccepted {
				time.Sleep(time.Second)

				w = httptest.NewRecorder()
				ctx, _ = gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/:"+ID, http.NoBody)
				ctx.Params = []gin.Param{
					{
						Key:   ID,
						Value: path.Base(parsedURL.Path),
					},
				}

				conntroller.GetShortLink(ctx)

				result = w.Result()

				err = result.Body.Close()
				assert.NoError(t, err)

				assert.Equal(t, http.StatusGone, result.StatusCode)
			}
		})
	}
}

func TestStats(t *testing.T) {
	tests := []struct {
		name    string
		request string
		want    int
	}{
		{
			name:    "valid data",
			request: "127.0.0.1",
			want:    http.StatusOK,
		},
		{
			name:    "invalid data",
			request: "128.0.0.1",
			want:    http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				BasicPath:     config.DefaultBasicPath,
				TrustedSubnet: "127.0.0.0/24",
			}
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				cfg.BasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			assert.NoError(t, err)
			conntroller := NewController(testLogger.Named("controller"), interactor, trustedSubnet)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/internal/stats", http.NoBody)
			ctx.Request.Header.Add("X-Real-IP", tt.request)

			middleware, err := middlewares.NewMiddleware(
				"../../../public.pem",
				"../../../private.pem",
				testLogger.Named("middleware"),
			)
			assert.NoError(t, err)
			auth := middleware.Auth()
			auth(ctx)

			conntroller.Stats(ctx)

			result := w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want, result.StatusCode)
		})
	}
}
