//nolint:wrapcheck // errors of grpc
package controllers

import (
	"context"
	"errors"
	"net"

	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	"github.com/RexArseny/url_shortener/internal/app/models"
	pb "github.com/RexArseny/url_shortener/internal/app/models/proto"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCController is responsible for managing the network interactions of the service with gRPC.
type GRPCController struct {
	pb.UnimplementedURLShortenerServer
	logger        *zap.Logger
	trustedSubnet *net.IPNet
	interactor    usecases.Interactor
}

// NewGRPCController create new GRPCController.
func NewGRPCController(
	logger *zap.Logger,
	interactor usecases.Interactor,
	trustedSubnet *net.IPNet,
) GRPCController {
	return GRPCController{
		logger:        logger,
		trustedSubnet: trustedSubnet,
		interactor:    interactor,
	}
}

// CreateShortLink create new short URL from original URL.
// Generate new JWT if it is not presented.
func (c *GRPCController) CreateShortLink(
	ctx context.Context,
	in *pb.CreateShortLinkRequest,
) (*pb.CreateShortLinkResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	var userID uuid.UUID
	userIDs := md.Get(middlewares.UserID)
	for _, item := range userIDs {
		var err error
		userID, err = uuid.Parse(item)
		if err == nil {
			break
		}
	}
	if userID.String() == "" || userID.String() == uuid.Nil.String() {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	result, err := c.interactor.CreateShortLink(ctx, in.GetOriginalUrl(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidURL) {
			return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
		}
		c.logger.Error("Can not create short link", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	if result == nil || *result == "" {
		c.logger.Error("Short link is empty", zap.String("originalURL", in.GetOriginalUrl()))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.CreateShortLinkResponse{
		ShortUrl: *result,
	}, nil
}

// CreateShortLinkJSON create new short URL from original URL.
// Generate new JWT if it is not presented.
func (c *GRPCController) CreateShortLinkJSON(
	ctx context.Context,
	in *pb.CreateShortLinkJSONRequest,
) (*pb.CreateShortLinkJSONResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	var userID uuid.UUID
	userIDs := md.Get(middlewares.UserID)
	for _, item := range userIDs {
		var err error
		userID, err = uuid.Parse(item)
		if err == nil {
			break
		}
	}
	if userID.String() == "" || userID.String() == uuid.Nil.String() {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	result, err := c.interactor.CreateShortLink(ctx, in.GetRequest().GetUrl(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidURL) {
			return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
		}
		c.logger.Error("Can not create short link from json", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	if result == nil || *result == "" {
		c.logger.Error("Short link from json is empty", zap.String("url", in.GetRequest().GetUrl()))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.CreateShortLinkJSONResponse{
		Response: &pb.CreateShortLinkJSONResponse_Response{
			Result: *result,
		},
	}, nil
}

// CreateShortLinkJSONBatch create new short URLs from original URLs.
// Generate new JWT if it is not presented.
func (c *GRPCController) CreateShortLinkJSONBatch(
	ctx context.Context,
	in *pb.CreateShortLinkJSONBatchRequest,
) (*pb.CreateShortLinkJSONBatchResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	var userID uuid.UUID
	userIDs := md.Get(middlewares.UserID)
	for _, item := range userIDs {
		var err error
		userID, err = uuid.Parse(item)
		if err == nil {
			break
		}
	}
	if userID.String() == "" || userID.String() == uuid.Nil.String() {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	requests := in.GetRequests()
	request := make([]models.ShortenBatchRequest, 0, len(requests))
	for i := range requests {
		request = append(request, models.ShortenBatchRequest{
			CorrelationID: requests[i].GetCorrelationId(),
			OriginalURL:   requests[i].GetOriginalUrl(),
		})
	}

	result, err := c.interactor.CreateShortLinks(ctx, request, userID)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidURL) {
			return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
		}
		c.logger.Error("Can not create short links", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	response := make([]*pb.CreateShortLinkJSONBatchResponse_Response, 0, len(result))
	for i := range result {
		response = append(response, &pb.CreateShortLinkJSONBatchResponse_Response{
			CorrelationId: result[i].CorrelationID,
			ShortUrl:      result[i].ShortURL,
		})
	}

	return &pb.CreateShortLinkJSONBatchResponse{
		Responses: response,
	}, nil
}

// GetShortLink return original URL from short URL.
func (c *GRPCController) GetShortLink(
	ctx context.Context,
	in *pb.GetShortLinkRequest,
) (*pb.GetShortLinkResponse, error) {
	result, err := c.interactor.GetShortLink(ctx, in.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrURLIsDeleted) {
			return nil, status.Errorf(codes.NotFound, "url is deleted")
		}
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	if result == nil || *result == "" {
		c.logger.Error("Original URL is empty", zap.String("id", in.GetId()))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.GetShortLinkResponse{
		OriginalUrl: *result,
	}, nil
}

// PingDB ping and return the status of database.
func (c *GRPCController) PingDB(
	ctx context.Context,
	in *pb.PingDBRequest,
) (*pb.PingDBResponse, error) {
	err := c.interactor.PingDB(ctx)
	if err != nil {
		c.logger.Error("Can not ping db", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.PingDBResponse{
		Status: "OK",
	}, nil
}

// GetShortLinksOfUser return all short and original URLs of user if such exist and JWT is presented.
func (c *GRPCController) GetShortLinksOfUser(
	ctx context.Context,
	in *pb.GetShortLinksOfUserRequest,
) (*pb.GetShortLinksOfUserResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	authorizationNew := md.Get(middlewares.AuthorizationNew)
	for _, item := range authorizationNew {
		if item == middlewares.AuthorizationNew {
			return nil, status.Errorf(codes.NotFound, "no content")
		}
	}
	var userID uuid.UUID
	userIDs := md.Get(middlewares.UserID)
	for _, item := range userIDs {
		var err error
		userID, err = uuid.Parse(item)
		if err == nil {
			break
		}
	}
	if userID.String() == "" || userID.String() == uuid.Nil.String() {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	result, err := c.interactor.GetShortLinksOfUser(ctx, userID)
	if err != nil {
		c.logger.Error("Can not get short links of user", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	if len(result) == 0 {
		return nil, status.Errorf(codes.NotFound, "no content")
	}

	response := make([]*pb.GetShortLinksOfUserResponse_Response, 0, len(result))
	for i := range result {
		response = append(response, &pb.GetShortLinksOfUserResponse_Response{
			ShortUrl:    result[i].ShortURL,
			OriginalUrl: result[i].OriginalURL,
		})
	}

	return &pb.GetShortLinksOfUserResponse{
		Responses: response,
	}, nil
}

// DeleteURLs delete short URLs of user if such exist and JWT is presented.
func (c *GRPCController) DeleteURLs(
	ctx context.Context,
	in *pb.DeleteURLsRequest,
) (*pb.DeleteURLsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	authorizationNew := md.Get(middlewares.AuthorizationNew)
	for _, item := range authorizationNew {
		if item == middlewares.AuthorizationNew {
			return nil, status.Errorf(codes.NotFound, "no content")
		}
	}
	var userID uuid.UUID
	userIDs := md.Get(middlewares.UserID)
	for _, item := range userIDs {
		var err error
		userID, err = uuid.Parse(item)
		if err == nil {
			break
		}
	}
	if userID.String() == "" || userID.String() == uuid.Nil.String() {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	err := c.interactor.DeleteURLs(ctx, in.GetUrls(), userID)
	if err != nil {
		c.logger.Error("Can not delete urls", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.DeleteURLsResponse{
		Status: "OK",
	}, nil
}

// Stats return statistic of shortened urls and users in service.
func (c *GRPCController) Stats(
	ctx context.Context,
	in *pb.StatsRequest,
) (*pb.StatsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}
	var ip net.IP
	for _, item := range md.Get("X-Real-IP") {
		ip = net.ParseIP(item)
		if ip.String() != "" {
			break
		}
	}
	if ip.String() == "" || c.trustedSubnet == nil || !c.trustedSubnet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	stats, err := c.interactor.Stats(ctx)
	if err != nil {
		c.logger.Error("Can not get stats", zap.Error(err))
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &pb.StatsResponse{
		Urls:  uint64(stats.URLs),
		Users: uint64(stats.Users),
	}, nil
}
