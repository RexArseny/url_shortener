package controllers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/usecases"
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
				contenType:  "",
				response:    false,
			},
		},
		{
			name:    "invalid url",
			request: "abc",
			want: want{
				stastusCode: http.StatusBadRequest,
				contenType:  "",
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

	interactor := usecases.NewInteractor()
	conntroller := NewController(interactor)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))
			w := httptest.NewRecorder()
			conntroller.CreateShortLink(w, request)

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
	type want struct {
		stastusCode int
		location    string
	}
	tests := []struct {
		name    string
		request bool
		want    want
	}{
		{
			name:    "empty id",
			request: false,
			want: want{
				stastusCode: http.StatusBadRequest,
				location:    "",
			},
		},
		{
			name:    "valid id",
			request: true,
			want: want{
				stastusCode: http.StatusTemporaryRedirect,
				location:    "https://ya.ru",
			},
		},
	}

	interactor := usecases.NewInteractor()
	conntroller := NewController(interactor)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ya.ru"))
	w := httptest.NewRecorder()
	conntroller.CreateShortLink(w, request)

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id string
			if tt.request {
				id = path.Base(parsedURL.Path)
			}

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue(ID, id)
			w := httptest.NewRecorder()
			conntroller.GetShortLink(w, request)

			result := w.Result()

			err = result.Body.Close()
			assert.NoError(t, err)

			assert.Equal(t, tt.want.stastusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
