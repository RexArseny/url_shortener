package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	pb "github.com/RexArseny/url_shortener/internal/app/models/proto"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/routers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/gin-contrib/pprof"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	var trustedSubnet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, trustedSubnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("can not parse cidr from trusted subnet: %w", err)
		}
	}

	interactor := usecases.NewInteractor(ctx, mainLogger.Named("interactor"), cfg.BasicPath, urlRepository)
	controller := controllers.NewController(mainLogger.Named("controller"), interactor, trustedSubnet)
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

	grpcServerOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(
		middleware.GRPCLogger,
		middleware.GRPCAuth,
	)}

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

		grpcServerOpts = append(grpcServerOpts, grpc.Creds(credentials.NewServerTLSFromCert(&x509Cert)))
	}

	listener, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("can not init listener: %w", err)
	}
	defer func() {
		err = listener.Close()
		if err != nil {
			mainLogger.Fatal("Can not close listener", zap.Error(err))
		}
	}()

	grpcController := controllers.NewGRPCController(mainLogger.Named("grpccontroller"), interactor, trustedSubnet)

	grpcServer := grpc.NewServer(grpcServerOpts...)

	pb.RegisterURLShortenerServer(grpcServer, &grpcController)

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	go func() {
		<-ctx.Done()
		err = server.Shutdown(ctx)
		if err != nil {
			mainLogger.Fatal("Can not shutdown server", zap.Error(err))
		}
		grpcServer.GracefulStop()
	}()

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			mainLogger.Fatal("Can not serve grpc", zap.Error(err))
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
