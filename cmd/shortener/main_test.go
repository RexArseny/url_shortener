package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gojek/heimdall/v7/httpclient"
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
				ServerAddress: config.DefaultServerAddress,
				BasicPath:     config.DefaultBasicPath,
			}
			interactor := usecases.NewInteractor(cfg.BasicPath)
			conntroller := controllers.NewController(interactor)
			router, err := routers.NewRouter(&cfg, conntroller)
			assert.NoError(t, err)

			server := httptest.NewServer(router)
			defer server.Close()

			client := httpclient.NewClient(httpclient.WithHTTPTimeout(15 * time.Second))

			result, err := client.Post(fmt.Sprintf("%s/", server.URL), strings.NewReader(tt.request), nil)
			assert.NoError(t, err)

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
				stastusCode: http.StatusNotFound,
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
				ServerAddress: config.DefaultServerAddress,
				BasicPath:     config.DefaultBasicPath,
			}
			interactor := usecases.NewInteractor(cfg.BasicPath)
			conntroller := controllers.NewController(interactor)
			router, err := routers.NewRouter(&cfg, conntroller)
			assert.NoError(t, err)

			server := httptest.NewServer(router)
			defer server.Close()

			client := httpclient.NewClient(httpclient.WithHTTPClient(&http.Client{
				Timeout: 15 * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}))

			result, err := client.Post(fmt.Sprintf("%s/", server.URL), strings.NewReader("https://ya.ru"), nil)
			assert.NoError(t, err)

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

			result, err = client.Get(fmt.Sprintf("%s/%s", server.URL, id), nil)
			assert.NoError(t, err)

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
