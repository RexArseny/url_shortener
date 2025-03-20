package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-contrib/pprof"
	"go.uber.org/zap"
)

// NewServer create new server with new interactor, controller, middleware and router.
func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	urlRepository repository.Repository,
) (*http.Server, error) {
	interactor := usecases.NewInteractor(ctx, logger.Named("interactor"), cfg.BasicPath, urlRepository)
	controller := controllers.NewController(logger.Named("controller"), interactor)
	middleware, err := middlewares.NewMiddleware(
		cfg.PublicKeyPath,
		cfg.PrivateKeyPath,
		logger.Named("middleware"),
	)
	if err != nil {
		return nil, fmt.Errorf("can not init middleware: %w", err)
	}
	router, err := routers.NewRouter(cfg, controller, middleware)
	if err != nil {
		return nil, fmt.Errorf("can not init router: %w", err)
	}

	pprof.Register(router)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	if cfg.EnableHTTPS && cfg.CertificatePath != "" && cfg.CertificateKeyPath != "" {
		certBytes, err := os.ReadFile(cfg.CertificatePath)
		if err != nil {
			return nil, fmt.Errorf("can not read certificate file: %w", err)
		}

		keyBytes, err := os.ReadFile(cfg.CertificateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("can not read certificate key file: %w", err)
		}

		x509Cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("can not create x509 key pair: %w", err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{x509Cert},
			MinVersion:   tls.VersionTLS13,
		}
	}

	return server, nil
}
