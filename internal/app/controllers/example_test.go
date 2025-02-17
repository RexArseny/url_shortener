package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-gonic/gin"
)

func ExampleController_CreateShortLink() {
	cfg := config.Config{
		BasicPath: config.DefaultBasicPath,
	}
	testLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("can not init logger: %s", err)
	}
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
	if err != nil {
		log.Fatalf("can not init middleware: %s", err)
	}
	auth := middleware.Auth()
	auth(ctx)

	conntroller.CreateShortLink(ctx)

	result := w.Result()

	fmt.Println(result.StatusCode)

	// Output:
	// 201
}

func ExampleController_CreateShortLinkJSON() {
	cfg := config.Config{
		BasicPath: config.DefaultBasicPath,
	}
	testLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("can not init logger: %s", err)
	}
	interactor := usecases.NewInteractor(
		context.Background(),
		testLogger.Named("interactor"),
		cfg.BasicPath,
		repository.NewLinks(),
	)
	conntroller := NewController(testLogger.Named("controller"), interactor)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"url":"https://ya.ru"}`))

	middleware, err := middlewares.NewMiddleware(
		"../../../public.pem",
		"../../../private.pem",
		testLogger.Named("middleware"),
	)
	if err != nil {
		log.Fatalf("can not init middleware: %s", err)
	}
	auth := middleware.Auth()
	auth(ctx)

	conntroller.CreateShortLinkJSON(ctx)

	result := w.Result()

	fmt.Println(result.StatusCode)

	// Output:
	// 201
}
