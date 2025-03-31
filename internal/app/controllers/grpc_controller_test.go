package controllers

import (
	"context"
	"net"
	"net/url"
	"path"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/config"
	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/middlewares"
	pb "github.com/RexArseny/url_shortener/internal/app/models/proto"
	"github.com/RexArseny/url_shortener/internal/app/repository"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGRPCControllerCreateShortLink(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		in  *pb.CreateShortLinkRequest
		ctx context.Context
	}
	type want struct {
		err      bool
		response bool
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid request",
			request: request{
				in: &pb.CreateShortLinkRequest{
					OriginalUrl: "https://ya.ru",
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      false,
				response: true,
			},
		},
		{
			name: "no metadata",
			request: request{
				in: &pb.CreateShortLinkRequest{
					OriginalUrl: "https://ya.ru",
				},
				ctx: metadata.NewIncomingContext(context.Background(), nil),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid url",
			request: request{
				in: &pb.CreateShortLinkRequest{
					OriginalUrl: "abc",
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid metadata",
			request: request{
				in: &pb.CreateShortLinkRequest{
					OriginalUrl: "abc",
				},
				ctx: context.Background(),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp, err := conntroller.CreateShortLink(tt.request.ctx, tt.request.in)

			if tt.want.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.want.response {
				assert.NotEmpty(t, resp.ShortUrl)
			}
		})
	}
}

func TestGRPCControllerCreateShortLinkJSON(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		in  *pb.CreateShortLinkJSONRequest
		ctx context.Context
	}
	type want struct {
		err      bool
		response bool
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid request",
			request: request{
				in: &pb.CreateShortLinkJSONRequest{
					Request: &pb.CreateShortLinkJSONRequest_Request{
						Url: "https://ya.ru",
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      false,
				response: true,
			},
		},
		{
			name: "no metadata",
			request: request{
				in: &pb.CreateShortLinkJSONRequest{
					Request: &pb.CreateShortLinkJSONRequest_Request{
						Url: "https://ya.ru",
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), nil),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid url",
			request: request{
				in: &pb.CreateShortLinkJSONRequest{
					Request: &pb.CreateShortLinkJSONRequest_Request{
						Url: "abc",
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid metadata",
			request: request{
				in: &pb.CreateShortLinkJSONRequest{
					Request: &pb.CreateShortLinkJSONRequest_Request{
						Url: "abc",
					},
				},
				ctx: context.Background(),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp, err := conntroller.CreateShortLinkJSON(tt.request.ctx, tt.request.in)

			if tt.want.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.want.response {
				assert.NotEmpty(t, resp.Response.Result)
			}
		})
	}
}

func TestGRPCControllerCreateShortLinkJSONBatch(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		in  *pb.CreateShortLinkJSONBatchRequest
		ctx context.Context
	}
	type want struct {
		err      bool
		response bool
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid request",
			request: request{
				in: &pb.CreateShortLinkJSONBatchRequest{
					Requests: []*pb.CreateShortLinkJSONBatchRequest_Request{
						{
							CorrelationId: "1",
							OriginalUrl:   "https://ya.ru",
						},
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      false,
				response: true,
			},
		},
		{
			name: "no metadata",
			request: request{
				in: &pb.CreateShortLinkJSONBatchRequest{
					Requests: []*pb.CreateShortLinkJSONBatchRequest_Request{
						{
							CorrelationId: "1",
							OriginalUrl:   "https://ya.ru",
						},
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), nil),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid url",
			request: request{
				in: &pb.CreateShortLinkJSONBatchRequest{
					Requests: []*pb.CreateShortLinkJSONBatchRequest_Request{
						{
							CorrelationId: "1",
							OriginalUrl:   "abc",
						},
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid metadata",
			request: request{
				in: &pb.CreateShortLinkJSONBatchRequest{
					Requests: []*pb.CreateShortLinkJSONBatchRequest_Request{
						{
							CorrelationId: "1",
							OriginalUrl:   "abc",
						},
					},
				},
				ctx: context.Background(),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp, err := conntroller.CreateShortLinkJSONBatch(tt.request.ctx, tt.request.in)

			if tt.want.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.want.response {
				assert.NotEmpty(t, resp.Responses[0].ShortUrl)
				assert.NotEmpty(t, resp.Responses[0].CorrelationId)
			}
		})
	}
}

func TestGRPCControllerGetShortLink(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		valid bool
		md    metadata.MD
	}
	type want struct {
		err      bool
		response bool
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid request",
			request: request{
				valid: true,
				md:    metadata.Pairs(middlewares.UserID, testUserID.String()),
			},
			want: want{
				err:      false,
				response: true,
			},
		},
		{
			name: "no metadata",
			request: request{
				valid: false,
				md:    nil,
			},
			want: want{
				err:      true,
				response: false,
			},
		},
		{
			name: "invalid url",
			request: request{
				valid: false,
				md:    metadata.Pairs(middlewares.UserID, testUserID.String()),
			},
			want: want{
				err:      true,
				response: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			ctx := metadata.NewIncomingContext(context.Background(), tt.request.md)
			data, err := conntroller.CreateShortLink(ctx, &pb.CreateShortLinkRequest{
				OriginalUrl: "https://ya.ru",
			})
			if tt.want.response {
				assert.NoError(t, err)
			} else {
				if !tt.want.err {
					assert.Error(t, err)
				}
			}

			var request *pb.GetShortLinkRequest
			if tt.request.valid {
				parsedURL, err := url.ParseRequestURI(data.ShortUrl)
				assert.NoError(t, err)
				assert.NotEmpty(t, parsedURL)

				request = &pb.GetShortLinkRequest{
					Id: path.Base(parsedURL.Path),
				}
			}

			resp, err := conntroller.GetShortLink(ctx, request)

			if tt.want.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.want.response {
				assert.NotEmpty(t, resp.OriginalUrl)
			}
		})
	}
}

func TestGRPCControllerPingDB(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		md metadata.MD
	}
	tests := []struct {
		name    string
		request request
	}{
		{
			name: "valid request",
			request: request{
				md: metadata.Pairs(middlewares.UserID, testUserID.String()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			ctx := metadata.NewIncomingContext(context.Background(), tt.request.md)
			resp, err := conntroller.PingDB(ctx, &pb.PingDBRequest{})

			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Status)
		})
	}
}

func TestGRPCControllerGetShortLinksOfUser(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		request request
		err     error
	}{
		{
			name: "valid request",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			err: nil,
		},
		{
			name: "invalid metadata",
			request: request{
				ctx: context.Background(),
			},
			err: status.Error(codes.Unauthenticated, codes.Unauthenticated.String()),
		},
		{
			name: "authorization new",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.AuthorizationNew, middlewares.AuthorizationNew)),
			},
			err: status.Errorf(codes.NotFound, "no content"),
		},
		{
			name: "invalid user id",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			},
			err: status.Error(codes.Unauthenticated, codes.Unauthenticated.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp1, err := conntroller.GetShortLinksOfUser(tt.request.ctx, &pb.GetShortLinksOfUserRequest{})

			assert.Error(t, err)
			assert.Empty(t, resp1)

			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
				return
			}

			data, err := conntroller.CreateShortLink(tt.request.ctx, &pb.CreateShortLinkRequest{
				OriginalUrl: "https://ya.ru",
			})
			assert.NoError(t, err)
			assert.NotEmpty(t, data.ShortUrl)

			resp2, err := conntroller.GetShortLinksOfUser(tt.request.ctx, &pb.GetShortLinksOfUserRequest{})

			assert.NoError(t, err)
			assert.NotEmpty(t, resp2.Responses)
		})
	}
}

func TestGRPCControllerDeleteURLs(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		request request
		err     error
	}{
		{
			name: "valid request",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String())),
			},
			err: nil,
		},
		{
			name: "invalid metadata",
			request: request{
				ctx: context.Background(),
			},
			err: status.Error(codes.Unauthenticated, codes.Unauthenticated.String()),
		},
		{
			name: "authorization new",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.AuthorizationNew, middlewares.AuthorizationNew)),
			},
			err: status.Errorf(codes.NotFound, "no content"),
		},
		{
			name: "invalid user id",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			},
			err: status.Error(codes.Unauthenticated, codes.Unauthenticated.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp1, err := conntroller.DeleteURLs(tt.request.ctx, &pb.DeleteURLsRequest{Urls: []string{}})

			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
				assert.Empty(t, resp1)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, resp1.Status)

			data, err := conntroller.CreateShortLink(tt.request.ctx, &pb.CreateShortLinkRequest{
				OriginalUrl: "https://ya.ru",
			})
			assert.NoError(t, err)
			assert.NotEmpty(t, data.ShortUrl)

			parsedURL, err := url.ParseRequestURI(data.ShortUrl)
			assert.NoError(t, err)
			assert.NotEmpty(t, parsedURL)

			resp2, err := conntroller.DeleteURLs(tt.request.ctx, &pb.DeleteURLsRequest{Urls: []string{path.Base(parsedURL.Path)}})
			assert.NoError(t, err)
			assert.NotEmpty(t, resp2.Status)

			resp3, err := conntroller.GetShortLinksOfUser(tt.request.ctx, &pb.GetShortLinksOfUserRequest{})

			assert.NoError(t, err)
			assert.NotEmpty(t, resp3.Responses)
		})
	}
}

func TestGRPCControllerStats(t *testing.T) {
	testUserID := uuid.New()
	type request struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		request request
		err     error
	}{
		{
			name: "valid request",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String(), "X-Real-IP", "127.0.0.1")),
			},
			err: nil,
		},
		{
			name: "invalid request",
			request: request{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(middlewares.UserID, testUserID.String(), "X-Real-IP", "128.0.0.1")),
			},
			err: status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
		},
		{
			name: "invalid metadata",
			request: request{
				ctx: context.Background(),
			},
			err: status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger, err := logger.InitLogger()
			assert.NoError(t, err)
			interactor := usecases.NewInteractor(
				context.Background(),
				testLogger.Named("interactor"),
				config.DefaultBasicPath,
				repository.NewLinks(),
			)
			_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/24")
			assert.NoError(t, err)
			conntroller := NewGRPCController(testLogger.Named("controller"), interactor, trustedSubnet)

			resp1, err := conntroller.Stats(tt.request.ctx, &pb.StatsRequest{})

			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
				assert.Empty(t, resp1)
				return
			}

			assert.NoError(t, err)

			data, err := conntroller.CreateShortLink(tt.request.ctx, &pb.CreateShortLinkRequest{
				OriginalUrl: "https://ya.ru",
			})
			assert.NoError(t, err)
			assert.NotEmpty(t, data.ShortUrl)

			resp2, err := conntroller.Stats(tt.request.ctx, &pb.StatsRequest{})

			assert.NoError(t, err)
			assert.NotEmpty(t, resp2.Urls)
			assert.NotEmpty(t, resp2.Users)
		})
	}
}
