package middlewares

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGRPCLogger(t *testing.T) {
	t.Run("successful logging", func(t *testing.T) {
		core, recordedLogs := observer.New(zap.InfoLevel)
		testLogger := zap.New(core)

		middleware := &Middleware{
			logger: testLogger,
		}

		testHandler := func(
			ctx context.Context,
			in interface{},
		) (interface{}, error) {
			return nil, nil
		}

		_, err := middleware.GRPCLogger(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "test"}, testHandler)
		assert.NoError(t, err)

		statusCode, _ := status.FromError(err)

		assert.Equal(t, 1, recordedLogs.Len())
		logEntry := recordedLogs.All()[0]
		assert.Equal(t, "gRPC Request", logEntry.Message)
		code, ok := logEntry.ContextMap()["code"].(string)
		assert.True(t, ok)
		assert.Equal(t, statusCode.Code().String(), code)
		method, ok := logEntry.ContextMap()["method"].(string)
		assert.True(t, ok)
		assert.Equal(t, "test", method)
	})
}

func TestGRPCAuth(t *testing.T) {
	t.Run("parse token", func(t *testing.T) {
		userID := uuid.New()

		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		publicKeyFile, err := os.ReadFile("../../../public.pem")
		assert.NoError(t, err)
		publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyFile)
		assert.NoError(t, err)

		privateKeyFile, err := os.ReadFile("../../../private.pem")
		assert.NoError(t, err)
		privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
		assert.NoError(t, err)

		middleware := &Middleware{
			publicKey:  publicKey,
			privateKey: privateKey,
			logger:     testLogger,
		}

		testHandler := func(
			ctx context.Context,
			in interface{},
		) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			assert.True(t, ok)
			var userIDVar uuid.UUID
			userIDs := md.Get(UserID)
			for _, item := range userIDs {
				var err error
				userIDVar, err = uuid.Parse(item)
				if err == nil {
					break
				}
			}
			assert.Equal(t, userID.String(), userIDVar.String())

			return nil, nil
		}

		claims := &JWT{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "url_shortener",
				Subject:   userID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.New().String(),
			},
			UserID: userID,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		tokenString, err := token.SignedString(middleware.privateKey)
		assert.NoError(t, err)

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(Authorization, tokenString))
		_, err = middleware.GRPCAuth(
			ctx,
			nil,
			&grpc.UnaryServerInfo{FullMethod: "test"},
			testHandler)
		assert.NoError(t, err)
	})

	t.Run("generate token", func(t *testing.T) {
		testLogger, err := logger.InitLogger()
		assert.NoError(t, err)

		publicKeyFile, err := os.ReadFile("../../../public.pem")
		assert.NoError(t, err)
		publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyFile)
		assert.NoError(t, err)

		privateKeyFile, err := os.ReadFile("../../../private.pem")
		assert.NoError(t, err)
		privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
		assert.NoError(t, err)

		middleware := &Middleware{
			publicKey:  publicKey,
			privateKey: privateKey,
			logger:     testLogger,
		}

		testHandler := func(
			ctx context.Context,
			in interface{},
		) (interface{}, error) {
			return nil, nil
		}

		ctx := metadata.NewIncomingContext(context.Background(), nil)
		_, err = middleware.GRPCAuth(
			ctx,
			nil,
			&grpc.UnaryServerInfo{FullMethod: "test"},
			testHandler)
		assert.Error(t, err)
	})
}
