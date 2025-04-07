package middlewares

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserID is a name of gRPC metadata parameter.
const UserID = "userID"

// GRPCLogger logs information about incoming request.
func (m *Middleware) GRPCLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	statusCode, _ := status.FromError(err)

	m.logger.Info("gRPC Request",
		zap.String("code", statusCode.Code().String()),
		zap.Duration("latency", time.Since(start)),
		zap.String("method", info.FullMethod))

	return resp, err
}

// GRPCAuth extract JWT if it is presented and generate new one if it is not presented.
func (m *Middleware) GRPCAuth(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return m.gRPCAuth(ctx, req, handler, md)
	}

	var claims *JWT
	for _, item := range md.Get(Authorization) {
		token, err := jwt.ParseWithClaims(
			item,
			&JWT{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodEdDSA {
					return nil, errors.New("jwt signature mismatch")
				}
				return m.publicKey, nil
			},
		)
		if err != nil {
			continue
		}

		claims, ok = token.Claims.(*JWT)
		if ok {
			break
		}
	}
	if claims == nil {
		return m.gRPCAuth(ctx, req, handler, md)
	}

	md.Set(UserID, []string{claims.UserID.String()}...)
	md.Set(AuthorizationNew, []string{}...)

	ctx = metadata.NewIncomingContext(ctx, md)

	return handler(ctx, req)
}

func (m *Middleware) gRPCAuth(
	ctx context.Context,
	req interface{},
	handler grpc.UnaryHandler,
	md metadata.MD,
) (interface{}, error) {
	userID := uuid.New()
	claims := &JWT{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "url_shortener",
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	tokenString, err := token.SignedString(m.privateKey)
	if err != nil {
		m.logger.Error("Can not sign token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	md.Set(Authorization, tokenString)

	err = grpc.SendHeader(ctx, md)
	if err != nil {
		m.logger.Error("Can not send header", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	md.Set(UserID, []string{userID.String()}...)
	md.Set(AuthorizationNew, []string{AuthorizationNew}...)

	ctx = metadata.NewIncomingContext(ctx, md)

	return handler(ctx, req)
}
