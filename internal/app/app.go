package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-contrib/pprof"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// NewServer create new server with new interactor, controller, middleware and router.
func NewServer() error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)
	defer cancel()

	mainLogger, err := logger.InitLogger()
	if err != nil {
		return fmt.Errorf("can not init logger: %w", err)
	}
	defer func() {
		var pathErr *fs.PathError
		if err = mainLogger.Sync(); err != nil && !errors.As(err, &pathErr) {
			log.Fatalf("Logger sync failed: %s", err)
		}
	}()

	cfg, err := config.Init()
	if err != nil {
		return fmt.Errorf("can not init config: %w", err)
	}

	urlRepository, repositoryClose, err := repository.NewRepository(
		ctx,
		mainLogger.Named("repository"),
		cfg.FileStoragePath,
		cfg.DatabaseDSN,
	)
	if err != nil {
		return fmt.Errorf("can not init repository: %w", err)
	}
	defer func() {
		if repositoryClose != nil {
			err = repositoryClose()
			if err != nil {
				mainLogger.Fatal("Can not close repository", zap.Error(err))
			}
		}
	}()

	interactor := usecases.NewInteractor(ctx, mainLogger.Named("interactor"), cfg.BasicPath, urlRepository)
	controller := controllers.NewController(mainLogger.Named("controller"), interactor)
	middleware, err := middlewares.NewMiddleware(
		cfg.PublicKeyPath,
		cfg.PrivateKeyPath,
		mainLogger.Named("middleware"),
	)
	if err != nil {
		return fmt.Errorf("can not init middleware: %w", err)
	}
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		return fmt.Errorf("can not init router: %w", err)
	}

	pprof.Register(router)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	if cfg.EnableHTTPS && cfg.CertificatePath != "" && cfg.CertificateKeyPath != "" {
		certBytes, err := os.ReadFile(cfg.CertificatePath)
		if err != nil {
			return fmt.Errorf("can not read certificate file: %w", err)
		}

		keyBytes, err := os.ReadFile(cfg.CertificateKeyPath)
		if err != nil {
			return fmt.Errorf("can not read certificate key file: %w", err)
		}

		x509Cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return fmt.Errorf("can not create x509 key pair: %w", err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{x509Cert},
			MinVersion:   tls.VersionTLS13,
		}
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	go func() {
		<-ctx.Done()
		err = server.Shutdown(ctx)
		if err != nil {
			mainLogger.Fatal("Can not shutdown server", zap.Error(err))
		}
	}()

	if cfg.EnableHTTPS {
		err = server.ListenAndServeTLS("", "")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("can not listen and serve: %w", err)
		}

		fmt.Println("Server shutdown gracefully")

		return nil
	}

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("can not listen and serve: %w", err)
	}

	fmt.Println("Server shutdown gracefully")

	return nil
}
