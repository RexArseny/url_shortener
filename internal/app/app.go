package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"time"

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

	if cfg.EnableHTTPS {
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				Organization: []string{"url_shortener"},
			},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().AddDate(1, 0, 0),
			SubjectKeyId: []byte{1, 2, 3, 4, 6},
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:     x509.KeyUsageDigitalSignature,
		}

		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("can generate rsa key: %w", err)
		}

		certificate, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
		if err != nil {
			return nil, fmt.Errorf("can not create certificate: %w", err)
		}

		certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
		keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

		x509Cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("can not create x509 key pair: %w", err)
		}

		server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{x509Cert}}
	}

	return server, nil
}
