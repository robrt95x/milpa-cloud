package core

import (
	"context"

	"github.com/robrt95x/milpa-cloud/pkg/types"

	"google.golang.org/grpc"
)

// PluginServiceServer is the server interface
type PluginServiceServer interface {
	Handshake(context.Context, *HandshakeRequest) (*HandshakeResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	Configure(context.Context, *ConfigureRequest) (*ConfigureResponse, error)
	Stream(*pluginStreamServer) error
}

// pluginStreamServer wraps grpc.ServerStream
type pluginStreamServer struct {
	grpc.ServerStream
}

func (p *pluginStreamServer) SendMsg(m interface{}) error {
	return p.ServerStream.SendMsg(m)
}

func (p *pluginStreamServer) RecvMsg(m interface{}) error {
	return p.ServerStream.RecvMsg(m)
}

// Types are imported from pkg/types
// HandshakeRequest, HandshakeResponse, HeartbeatRequest, HeartbeatResponse,
// ConfigureRequest, ConfigureResponse, PluginEvent, CoreEvent

// Import the types package for use in the service
type (
	HandshakeRequest  = types.HandshakeRequest
	HandshakeResponse = types.HandshakeResponse
	HeartbeatRequest  = types.HeartbeatRequest
	HeartbeatResponse = types.HeartbeatResponse
	ConfigureRequest  = types.ConfigureRequest
	ConfigureResponse = types.ConfigureResponse
	PluginEvent      = types.PluginEvent
	CoreEvent        = types.CoreEvent
)

// RegisterPluginServiceServer registers the service
func RegisterPluginServiceServer(s *grpc.Server, srv PluginServiceServer) {
	s.RegisterService(&_PluginService_serviceDesc, srv)
}

var _PluginService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "milpa.v1.PluginService",
	HandlerType: (*PluginServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Handshake",
			Handler:    _PluginService_Handshake_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _PluginService_Heartbeat_Handler,
		},
		{
			MethodName: "Configure",
			Handler:    _PluginService_Configure_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _PluginService_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "plugin.proto",
}

func _PluginService_Handshake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	var in HandshakeRequest
	if err := dec(&in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).Handshake(ctx, &in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/milpa.v1.PluginService/Handshake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).Handshake(ctx, req.(*HandshakeRequest))
	}
	return interceptor(ctx, &in, info, handler)
}

func _PluginService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	var in HeartbeatRequest
	if err := dec(&in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).Heartbeat(ctx, &in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/milpa.v1.PluginService/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, &in, info, handler)
}

func _PluginService_Configure_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	var in ConfigureRequest
	if err := dec(&in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).Configure(ctx, &in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/milpa.v1.PluginService/Configure",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).Configure(ctx, req.(*ConfigureRequest))
	}
	return interceptor(ctx, &in, info, handler)
}

func _PluginService_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PluginServiceServer).Stream(&pluginStreamServer{stream})
}
