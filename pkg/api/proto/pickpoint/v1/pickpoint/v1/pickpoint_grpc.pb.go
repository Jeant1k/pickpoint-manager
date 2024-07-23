// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v3.12.4
// source: pickpoint/v1/pickpoint.proto

package pickpoint

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Pickpoint_RegistratePickPointId_FullMethodName  = "/pickpoint.Pickpoint/RegistratePickPointId"
	Pickpoint_AcceptOrderFromCurier_FullMethodName  = "/pickpoint.Pickpoint/AcceptOrderFromCurier"
	Pickpoint_ReturnOrderToCurier_FullMethodName    = "/pickpoint.Pickpoint/ReturnOrderToCurier"
	Pickpoint_IssueOrderToClient_FullMethodName     = "/pickpoint.Pickpoint/IssueOrderToClient"
	Pickpoint_ListOrders_FullMethodName             = "/pickpoint.Pickpoint/ListOrders"
	Pickpoint_AcceptReturnFromClient_FullMethodName = "/pickpoint.Pickpoint/AcceptReturnFromClient"
	Pickpoint_ListReturns_FullMethodName            = "/pickpoint.Pickpoint/ListReturns"
	Pickpoint_Help_FullMethodName                   = "/pickpoint.Pickpoint/Help"
)

// PickpointClient is the client API for Pickpoint service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PickpointClient interface {
	RegistratePickPointId(ctx context.Context, in *RegistratePickPointIdRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AcceptOrderFromCurier(ctx context.Context, in *AcceptOrderFromCurierRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ReturnOrderToCurier(ctx context.Context, in *ReturnOrderToCurierRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	IssueOrderToClient(ctx context.Context, in *IssueOrderToClientRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (*ListOrdersResponse, error)
	AcceptReturnFromClient(ctx context.Context, in *AcceptReturnFromClientRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListReturns(ctx context.Context, in *ListReturnsRequest, opts ...grpc.CallOption) (*ListReturnsResponse, error)
	Help(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*HelpResponse, error)
}

type pickpointClient struct {
	cc grpc.ClientConnInterface
}

func NewPickpointClient(cc grpc.ClientConnInterface) PickpointClient {
	return &pickpointClient{cc}
}

func (c *pickpointClient) RegistratePickPointId(ctx context.Context, in *RegistratePickPointIdRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Pickpoint_RegistratePickPointId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) AcceptOrderFromCurier(ctx context.Context, in *AcceptOrderFromCurierRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Pickpoint_AcceptOrderFromCurier_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) ReturnOrderToCurier(ctx context.Context, in *ReturnOrderToCurierRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Pickpoint_ReturnOrderToCurier_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) IssueOrderToClient(ctx context.Context, in *IssueOrderToClientRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Pickpoint_IssueOrderToClient_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (*ListOrdersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListOrdersResponse)
	err := c.cc.Invoke(ctx, Pickpoint_ListOrders_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) AcceptReturnFromClient(ctx context.Context, in *AcceptReturnFromClientRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Pickpoint_AcceptReturnFromClient_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) ListReturns(ctx context.Context, in *ListReturnsRequest, opts ...grpc.CallOption) (*ListReturnsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListReturnsResponse)
	err := c.cc.Invoke(ctx, Pickpoint_ListReturns_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pickpointClient) Help(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*HelpResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HelpResponse)
	err := c.cc.Invoke(ctx, Pickpoint_Help_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PickpointServer is the server API for Pickpoint service.
// All implementations must embed UnimplementedPickpointServer
// for forward compatibility
type PickpointServer interface {
	RegistratePickPointId(context.Context, *RegistratePickPointIdRequest) (*emptypb.Empty, error)
	AcceptOrderFromCurier(context.Context, *AcceptOrderFromCurierRequest) (*emptypb.Empty, error)
	ReturnOrderToCurier(context.Context, *ReturnOrderToCurierRequest) (*emptypb.Empty, error)
	IssueOrderToClient(context.Context, *IssueOrderToClientRequest) (*emptypb.Empty, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error)
	AcceptReturnFromClient(context.Context, *AcceptReturnFromClientRequest) (*emptypb.Empty, error)
	ListReturns(context.Context, *ListReturnsRequest) (*ListReturnsResponse, error)
	Help(context.Context, *emptypb.Empty) (*HelpResponse, error)
	mustEmbedUnimplementedPickpointServer()
}

// UnimplementedPickpointServer must be embedded to have forward compatible implementations.
type UnimplementedPickpointServer struct {
}

func (UnimplementedPickpointServer) RegistratePickPointId(context.Context, *RegistratePickPointIdRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegistratePickPointId not implemented")
}
func (UnimplementedPickpointServer) AcceptOrderFromCurier(context.Context, *AcceptOrderFromCurierRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptOrderFromCurier not implemented")
}
func (UnimplementedPickpointServer) ReturnOrderToCurier(context.Context, *ReturnOrderToCurierRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReturnOrderToCurier not implemented")
}
func (UnimplementedPickpointServer) IssueOrderToClient(context.Context, *IssueOrderToClientRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IssueOrderToClient not implemented")
}
func (UnimplementedPickpointServer) ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}
func (UnimplementedPickpointServer) AcceptReturnFromClient(context.Context, *AcceptReturnFromClientRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcceptReturnFromClient not implemented")
}
func (UnimplementedPickpointServer) ListReturns(context.Context, *ListReturnsRequest) (*ListReturnsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReturns not implemented")
}
func (UnimplementedPickpointServer) Help(context.Context, *emptypb.Empty) (*HelpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Help not implemented")
}
func (UnimplementedPickpointServer) mustEmbedUnimplementedPickpointServer() {}

// UnsafePickpointServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PickpointServer will
// result in compilation errors.
type UnsafePickpointServer interface {
	mustEmbedUnimplementedPickpointServer()
}

func RegisterPickpointServer(s grpc.ServiceRegistrar, srv PickpointServer) {
	s.RegisterService(&Pickpoint_ServiceDesc, srv)
}

func _Pickpoint_RegistratePickPointId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegistratePickPointIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).RegistratePickPointId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_RegistratePickPointId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).RegistratePickPointId(ctx, req.(*RegistratePickPointIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_AcceptOrderFromCurier_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AcceptOrderFromCurierRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).AcceptOrderFromCurier(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_AcceptOrderFromCurier_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).AcceptOrderFromCurier(ctx, req.(*AcceptOrderFromCurierRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_ReturnOrderToCurier_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReturnOrderToCurierRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).ReturnOrderToCurier(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_ReturnOrderToCurier_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).ReturnOrderToCurier(ctx, req.(*ReturnOrderToCurierRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_IssueOrderToClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueOrderToClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).IssueOrderToClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_IssueOrderToClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).IssueOrderToClient(ctx, req.(*IssueOrderToClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_ListOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).ListOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_ListOrders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).ListOrders(ctx, req.(*ListOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_AcceptReturnFromClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AcceptReturnFromClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).AcceptReturnFromClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_AcceptReturnFromClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).AcceptReturnFromClient(ctx, req.(*AcceptReturnFromClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_ListReturns_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReturnsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).ListReturns(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_ListReturns_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).ListReturns(ctx, req.(*ListReturnsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Pickpoint_Help_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PickpointServer).Help(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Pickpoint_Help_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PickpointServer).Help(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Pickpoint_ServiceDesc is the grpc.ServiceDesc for Pickpoint service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Pickpoint_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pickpoint.Pickpoint",
	HandlerType: (*PickpointServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegistratePickPointId",
			Handler:    _Pickpoint_RegistratePickPointId_Handler,
		},
		{
			MethodName: "AcceptOrderFromCurier",
			Handler:    _Pickpoint_AcceptOrderFromCurier_Handler,
		},
		{
			MethodName: "ReturnOrderToCurier",
			Handler:    _Pickpoint_ReturnOrderToCurier_Handler,
		},
		{
			MethodName: "IssueOrderToClient",
			Handler:    _Pickpoint_IssueOrderToClient_Handler,
		},
		{
			MethodName: "ListOrders",
			Handler:    _Pickpoint_ListOrders_Handler,
		},
		{
			MethodName: "AcceptReturnFromClient",
			Handler:    _Pickpoint_AcceptReturnFromClient_Handler,
		},
		{
			MethodName: "ListReturns",
			Handler:    _Pickpoint_ListReturns_Handler,
		},
		{
			MethodName: "Help",
			Handler:    _Pickpoint_Help_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pickpoint/v1/pickpoint.proto",
}
