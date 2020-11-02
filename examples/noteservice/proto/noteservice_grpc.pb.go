// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// NoteStoreClient is the client API for NoteStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NoteStoreClient interface {
	GetNote(ctx context.Context, in *GetNoteRequest, opts ...grpc.CallOption) (*Note, error)
}

type noteStoreClient struct {
	cc grpc.ClientConnInterface
}

func NewNoteStoreClient(cc grpc.ClientConnInterface) NoteStoreClient {
	return &noteStoreClient{cc}
}

func (c *noteStoreClient) GetNote(ctx context.Context, in *GetNoteRequest, opts ...grpc.CallOption) (*Note, error) {
	out := new(Note)
	err := c.cc.Invoke(ctx, "/noteservice.NoteStore/GetNote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NoteStoreServer is the server API for NoteStore service.
// All implementations must embed UnimplementedNoteStoreServer
// for forward compatibility
type NoteStoreServer interface {
	GetNote(context.Context, *GetNoteRequest) (*Note, error)
	mustEmbedUnimplementedNoteStoreServer()
}

// UnimplementedNoteStoreServer must be embedded to have forward compatible implementations.
type UnimplementedNoteStoreServer struct {
}

func (UnimplementedNoteStoreServer) GetNote(context.Context, *GetNoteRequest) (*Note, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNote not implemented")
}
func (UnimplementedNoteStoreServer) mustEmbedUnimplementedNoteStoreServer() {}

// UnsafeNoteStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NoteStoreServer will
// result in compilation errors.
type UnsafeNoteStoreServer interface {
	mustEmbedUnimplementedNoteStoreServer()
}

func RegisterNoteStoreServer(s grpc.ServiceRegistrar, srv NoteStoreServer) {
	s.RegisterService(&_NoteStore_serviceDesc, srv)
}

func _NoteStore_GetNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteStoreServer).GetNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/noteservice.NoteStore/GetNote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteStoreServer).GetNote(ctx, req.(*GetNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NoteStore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "noteservice.NoteStore",
	HandlerType: (*NoteStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetNote",
			Handler:    _NoteStore_GetNote_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/noteservice.proto",
}

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Authenticate(ctx context.Context, in *AuthenticateRequest, opts ...grpc.CallOption) (*AuthenticateResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Authenticate(ctx context.Context, in *AuthenticateRequest, opts ...grpc.CallOption) (*AuthenticateResponse, error) {
	out := new(AuthenticateResponse)
	err := c.cc.Invoke(ctx, "/noteservice.AuthService/Authenticate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility
type AuthServiceServer interface {
	Authenticate(context.Context, *AuthenticateRequest) (*AuthenticateResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (UnimplementedAuthServiceServer) Authenticate(context.Context, *AuthenticateRequest) (*AuthenticateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&_AuthService_serviceDesc, srv)
}

func _AuthService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthenticateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/noteservice.AuthService/Authenticate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Authenticate(ctx, req.(*AuthenticateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "noteservice.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _AuthService_Authenticate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/noteservice.proto",
}