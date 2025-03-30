// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: internal/app/models/proto/dto.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	URLShortener_CreateShortLink_FullMethodName          = "/proto.URLShortener/CreateShortLink"
	URLShortener_CreateShortLinkJSON_FullMethodName      = "/proto.URLShortener/CreateShortLinkJSON"
	URLShortener_CreateShortLinkJSONBatch_FullMethodName = "/proto.URLShortener/CreateShortLinkJSONBatch"
	URLShortener_GetShortLink_FullMethodName             = "/proto.URLShortener/GetShortLink"
	URLShortener_GetShortLinksOfUser_FullMethodName      = "/proto.URLShortener/GetShortLinksOfUser"
	URLShortener_DeleteURLs_FullMethodName               = "/proto.URLShortener/DeleteURLs"
	URLShortener_PingDB_FullMethodName                   = "/proto.URLShortener/PingDB"
	URLShortener_Stats_FullMethodName                    = "/proto.URLShortener/Stats"
)

// URLShortenerClient is the client API for URLShortener service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type URLShortenerClient interface {
	CreateShortLink(ctx context.Context, in *CreateShortLinkRequest, opts ...grpc.CallOption) (*CreateShortLinkResponse, error)
	CreateShortLinkJSON(ctx context.Context, in *CreateShortLinkJSONRequest, opts ...grpc.CallOption) (*CreateShortLinkJSONResponse, error)
	CreateShortLinkJSONBatch(ctx context.Context, in *CreateShortLinkJSONBatchRequest, opts ...grpc.CallOption) (*CreateShortLinkJSONBatchResponse, error)
	GetShortLink(ctx context.Context, in *GetShortLinkRequest, opts ...grpc.CallOption) (*GetShortLinkResponse, error)
	GetShortLinksOfUser(ctx context.Context, in *GetShortLinksOfUserRequest, opts ...grpc.CallOption) (*GetShortLinksOfUserResponse, error)
	DeleteURLs(ctx context.Context, in *DeleteURLsRequest, opts ...grpc.CallOption) (*DeleteURLsResponse, error)
	PingDB(ctx context.Context, in *PingDBRequest, opts ...grpc.CallOption) (*PingDBResponse, error)
	Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error)
}

type uRLShortenerClient struct {
	cc grpc.ClientConnInterface
}

func NewURLShortenerClient(cc grpc.ClientConnInterface) URLShortenerClient {
	return &uRLShortenerClient{cc}
}

func (c *uRLShortenerClient) CreateShortLink(ctx context.Context, in *CreateShortLinkRequest, opts ...grpc.CallOption) (*CreateShortLinkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShortLinkResponse)
	err := c.cc.Invoke(ctx, URLShortener_CreateShortLink_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) CreateShortLinkJSON(ctx context.Context, in *CreateShortLinkJSONRequest, opts ...grpc.CallOption) (*CreateShortLinkJSONResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShortLinkJSONResponse)
	err := c.cc.Invoke(ctx, URLShortener_CreateShortLinkJSON_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) CreateShortLinkJSONBatch(ctx context.Context, in *CreateShortLinkJSONBatchRequest, opts ...grpc.CallOption) (*CreateShortLinkJSONBatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShortLinkJSONBatchResponse)
	err := c.cc.Invoke(ctx, URLShortener_CreateShortLinkJSONBatch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) GetShortLink(ctx context.Context, in *GetShortLinkRequest, opts ...grpc.CallOption) (*GetShortLinkResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShortLinkResponse)
	err := c.cc.Invoke(ctx, URLShortener_GetShortLink_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) GetShortLinksOfUser(ctx context.Context, in *GetShortLinksOfUserRequest, opts ...grpc.CallOption) (*GetShortLinksOfUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShortLinksOfUserResponse)
	err := c.cc.Invoke(ctx, URLShortener_GetShortLinksOfUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) DeleteURLs(ctx context.Context, in *DeleteURLsRequest, opts ...grpc.CallOption) (*DeleteURLsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteURLsResponse)
	err := c.cc.Invoke(ctx, URLShortener_DeleteURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) PingDB(ctx context.Context, in *PingDBRequest, opts ...grpc.CallOption) (*PingDBResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingDBResponse)
	err := c.cc.Invoke(ctx, URLShortener_PingDB_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *uRLShortenerClient) Stats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StatsResponse)
	err := c.cc.Invoke(ctx, URLShortener_Stats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// URLShortenerServer is the server API for URLShortener service.
// All implementations must embed UnimplementedURLShortenerServer
// for forward compatibility.
type URLShortenerServer interface {
	CreateShortLink(context.Context, *CreateShortLinkRequest) (*CreateShortLinkResponse, error)
	CreateShortLinkJSON(context.Context, *CreateShortLinkJSONRequest) (*CreateShortLinkJSONResponse, error)
	CreateShortLinkJSONBatch(context.Context, *CreateShortLinkJSONBatchRequest) (*CreateShortLinkJSONBatchResponse, error)
	GetShortLink(context.Context, *GetShortLinkRequest) (*GetShortLinkResponse, error)
	GetShortLinksOfUser(context.Context, *GetShortLinksOfUserRequest) (*GetShortLinksOfUserResponse, error)
	DeleteURLs(context.Context, *DeleteURLsRequest) (*DeleteURLsResponse, error)
	PingDB(context.Context, *PingDBRequest) (*PingDBResponse, error)
	Stats(context.Context, *StatsRequest) (*StatsResponse, error)
	mustEmbedUnimplementedURLShortenerServer()
}

// UnimplementedURLShortenerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedURLShortenerServer struct{}

func (UnimplementedURLShortenerServer) CreateShortLink(context.Context, *CreateShortLinkRequest) (*CreateShortLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortLink not implemented")
}
func (UnimplementedURLShortenerServer) CreateShortLinkJSON(context.Context, *CreateShortLinkJSONRequest) (*CreateShortLinkJSONResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortLinkJSON not implemented")
}
func (UnimplementedURLShortenerServer) CreateShortLinkJSONBatch(context.Context, *CreateShortLinkJSONBatchRequest) (*CreateShortLinkJSONBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortLinkJSONBatch not implemented")
}
func (UnimplementedURLShortenerServer) GetShortLink(context.Context, *GetShortLinkRequest) (*GetShortLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShortLink not implemented")
}
func (UnimplementedURLShortenerServer) GetShortLinksOfUser(context.Context, *GetShortLinksOfUserRequest) (*GetShortLinksOfUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShortLinksOfUser not implemented")
}
func (UnimplementedURLShortenerServer) DeleteURLs(context.Context, *DeleteURLsRequest) (*DeleteURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteURLs not implemented")
}
func (UnimplementedURLShortenerServer) PingDB(context.Context, *PingDBRequest) (*PingDBResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingDB not implemented")
}
func (UnimplementedURLShortenerServer) Stats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
func (UnimplementedURLShortenerServer) mustEmbedUnimplementedURLShortenerServer() {}
func (UnimplementedURLShortenerServer) testEmbeddedByValue()                      {}

// UnsafeURLShortenerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to URLShortenerServer will
// result in compilation errors.
type UnsafeURLShortenerServer interface {
	mustEmbedUnimplementedURLShortenerServer()
}

func RegisterURLShortenerServer(s grpc.ServiceRegistrar, srv URLShortenerServer) {
	// If the following call pancis, it indicates UnimplementedURLShortenerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&URLShortener_ServiceDesc, srv)
}

func _URLShortener_CreateShortLink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortLinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).CreateShortLink(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_CreateShortLink_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).CreateShortLink(ctx, req.(*CreateShortLinkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_CreateShortLinkJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortLinkJSONRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).CreateShortLinkJSON(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_CreateShortLinkJSON_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).CreateShortLinkJSON(ctx, req.(*CreateShortLinkJSONRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_CreateShortLinkJSONBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortLinkJSONBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).CreateShortLinkJSONBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_CreateShortLinkJSONBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).CreateShortLinkJSONBatch(ctx, req.(*CreateShortLinkJSONBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_GetShortLink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShortLinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).GetShortLink(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_GetShortLink_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).GetShortLink(ctx, req.(*GetShortLinkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_GetShortLinksOfUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShortLinksOfUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).GetShortLinksOfUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_GetShortLinksOfUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).GetShortLinksOfUser(ctx, req.(*GetShortLinksOfUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_DeleteURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).DeleteURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_DeleteURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).DeleteURLs(ctx, req.(*DeleteURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_PingDB_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingDBRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).PingDB(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_PingDB_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).PingDB(ctx, req.(*PingDBRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _URLShortener_Stats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(URLShortenerServer).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: URLShortener_Stats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(URLShortenerServer).Stats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// URLShortener_ServiceDesc is the grpc.ServiceDesc for URLShortener service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var URLShortener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.URLShortener",
	HandlerType: (*URLShortenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateShortLink",
			Handler:    _URLShortener_CreateShortLink_Handler,
		},
		{
			MethodName: "CreateShortLinkJSON",
			Handler:    _URLShortener_CreateShortLinkJSON_Handler,
		},
		{
			MethodName: "CreateShortLinkJSONBatch",
			Handler:    _URLShortener_CreateShortLinkJSONBatch_Handler,
		},
		{
			MethodName: "GetShortLink",
			Handler:    _URLShortener_GetShortLink_Handler,
		},
		{
			MethodName: "GetShortLinksOfUser",
			Handler:    _URLShortener_GetShortLinksOfUser_Handler,
		},
		{
			MethodName: "DeleteURLs",
			Handler:    _URLShortener_DeleteURLs_Handler,
		},
		{
			MethodName: "PingDB",
			Handler:    _URLShortener_PingDB_Handler,
		},
		{
			MethodName: "Stats",
			Handler:    _URLShortener_Stats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/app/models/proto/dto.proto",
}
