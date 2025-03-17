package main

import (
	"context"
	"fmt"
	"log"

	"github.com/RexArseny/url_shortener/internal/app"
	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	ctx := context.Background()

	mainLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Can not init logger: %s", err)
	}
	defer func() {
		if err = mainLogger.Sync(); err != nil {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		mainLogger.Fatal("Can not init config", zap.Error(err))
	}

	urlRepository, repositoryClose, err := repository.NewRepository(
		ctx,
		mainLogger.Named("repository"),
		cfg.FileStoragePath,
		cfg.DatabaseDSN,
	)
	if err != nil {
		mainLogger.Fatal("Can not init repository", zap.Error(err))
	}
	defer func() {
		if repositoryClose != nil {
			err = repositoryClose()
			if err != nil {
				mainLogger.Fatal("Can not close repository", zap.Error(err))
			}
		}
	}()

	s, err := app.NewServer(ctx, mainLogger, cfg, urlRepository)
	if err != nil {
		mainLogger.Fatal("Can not init server", zap.Error(err))
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if cfg.EnableHTTPS {
		err = s.ListenAndServeTLS("", "")
		if err != nil {
			mainLogger.Fatal("Can not listen and serve", zap.Error(err))
		}

		return
	}

	err = s.ListenAndServe()
	if err != nil {
		mainLogger.Fatal("Can not listen and serve", zap.Error(err))
	}
}
