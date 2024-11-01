package controllers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortLink(t *testing.T) {
	type want struct {
		stastusCode int
		contenType  string
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
			interactor := usecases.NewInteractor(cfg.BasicPath)
			conntroller := NewController(interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))

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

func TestGetShortLink(t *testing.T) {
	type input struct {
		valid bool
		path  string
	}
	type want struct {
		stastusCode int
		location    string
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
			interactor := usecases.NewInteractor(cfg.BasicPath)
			conntroller := NewController(interactor)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))

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
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/:%s", ID), nil)
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
