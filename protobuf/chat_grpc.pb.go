// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ChatClient is the client API for Chat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	Logout(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AddConversation(ctx context.Context, in *AddConversationRequest, opts ...grpc.CallOption) (*AddConversationResponse, error)
	GetConversation(ctx context.Context, in *GetConversationRequest, opts ...grpc.CallOption) (*GetConversationResponse, error)
	JoinGroup(ctx context.Context, in *JoinGroupRequest, opts ...grpc.CallOption) (*JoinGroupResponse, error)
	GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error)
	GetUsernames(ctx context.Context, in *GetUsernamesRequest, opts ...grpc.CallOption) (*GetUsernamesResponse, error)
	StreamMessages(ctx context.Context, in *StreamMessagesRequest, opts ...grpc.CallOption) (Chat_StreamMessagesClient, error)
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
}

type chatClient struct {
	cc grpc.ClientConnInterface
}

func NewChatClient(cc grpc.ClientConnInterface) ChatClient {
	return &chatClient{cc}
}

func (c *chatClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) Logout(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/Logout", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) AddConversation(ctx context.Context, in *AddConversationRequest, opts ...grpc.CallOption) (*AddConversationResponse, error) {
	out := new(AddConversationResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/AddConversation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) GetConversation(ctx context.Context, in *GetConversationRequest, opts ...grpc.CallOption) (*GetConversationResponse, error) {
	out := new(GetConversationResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/GetConversation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) JoinGroup(ctx context.Context, in *JoinGroupRequest, opts ...grpc.CallOption) (*JoinGroupResponse, error) {
	out := new(JoinGroupResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/JoinGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error) {
	out := new(GetMessagesResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/GetMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) GetUsernames(ctx context.Context, in *GetUsernamesRequest, opts ...grpc.CallOption) (*GetUsernamesResponse, error) {
	out := new(GetUsernamesResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/GetUsernames", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) StreamMessages(ctx context.Context, in *StreamMessagesRequest, opts ...grpc.CallOption) (Chat_StreamMessagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &Chat_ServiceDesc.Streams[0], "/protobuf.Chat/StreamMessages", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatStreamMessagesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Chat_StreamMessagesClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type chatStreamMessagesClient struct {
	grpc.ClientStream
}

func (x *chatStreamMessagesClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chatClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, "/protobuf.Chat/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChatServer is the server API for Chat service.
// All implementations must embed UnimplementedChatServer
// for forward compatibility
type ChatServer interface {
	Register(context.Context, *RegisterRequest) (*emptypb.Empty, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	Logout(context.Context, *LogoutRequest) (*emptypb.Empty, error)
	AddConversation(context.Context, *AddConversationRequest) (*AddConversationResponse, error)
	GetConversation(context.Context, *GetConversationRequest) (*GetConversationResponse, error)
	JoinGroup(context.Context, *JoinGroupRequest) (*JoinGroupResponse, error)
	GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error)
	GetUsernames(context.Context, *GetUsernamesRequest) (*GetUsernamesResponse, error)
	StreamMessages(*StreamMessagesRequest, Chat_StreamMessagesServer) error
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	mustEmbedUnimplementedChatServer()
}

// UnimplementedChatServer must be embedded to have forward compatible implementations.
type UnimplementedChatServer struct {
}

func (UnimplementedChatServer) Register(context.Context, *RegisterRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedChatServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedChatServer) Logout(context.Context, *LogoutRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedChatServer) AddConversation(context.Context, *AddConversationRequest) (*AddConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddConversation not implemented")
}
func (UnimplementedChatServer) GetConversation(context.Context, *GetConversationRequest) (*GetConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConversation not implemented")
}
func (UnimplementedChatServer) JoinGroup(context.Context, *JoinGroupRequest) (*JoinGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinGroup not implemented")
}
func (UnimplementedChatServer) GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessages not implemented")
}
func (UnimplementedChatServer) GetUsernames(context.Context, *GetUsernamesRequest) (*GetUsernamesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUsernames not implemented")
}
func (UnimplementedChatServer) StreamMessages(*StreamMessagesRequest, Chat_StreamMessagesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamMessages not implemented")
}
func (UnimplementedChatServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedChatServer) mustEmbedUnimplementedChatServer() {}

// UnsafeChatServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatServer will
// result in compilation errors.
type UnsafeChatServer interface {
	mustEmbedUnimplementedChatServer()
}

func RegisterChatServer(s grpc.ServiceRegistrar, srv ChatServer) {
	s.RegisterService(&Chat_ServiceDesc, srv)
}

func _Chat_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/Logout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).Logout(ctx, req.(*LogoutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_AddConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).AddConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/AddConversation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).AddConversation(ctx, req.(*AddConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_GetConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).GetConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/GetConversation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).GetConversation(ctx, req.(*GetConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_JoinGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).JoinGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/JoinGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).JoinGroup(ctx, req.(*JoinGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_GetMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).GetMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/GetMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).GetMessages(ctx, req.(*GetMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_GetUsernames_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUsernamesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).GetUsernames(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/GetUsernames",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).GetUsernames(ctx, req.(*GetUsernamesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_StreamMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamMessagesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServer).StreamMessages(m, &chatStreamMessagesServer{stream})
}

type Chat_StreamMessagesServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type chatStreamMessagesServer struct {
	grpc.ServerStream
}

func (x *chatStreamMessagesServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _Chat_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protobuf.Chat/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Chat_ServiceDesc is the grpc.ServiceDesc for Chat service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Chat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protobuf.Chat",
	HandlerType: (*ChatServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Chat_Register_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _Chat_Login_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _Chat_Logout_Handler,
		},
		{
			MethodName: "AddConversation",
			Handler:    _Chat_AddConversation_Handler,
		},
		{
			MethodName: "GetConversation",
			Handler:    _Chat_GetConversation_Handler,
		},
		{
			MethodName: "JoinGroup",
			Handler:    _Chat_JoinGroup_Handler,
		},
		{
			MethodName: "GetMessages",
			Handler:    _Chat_GetMessages_Handler,
		},
		{
			MethodName: "GetUsernames",
			Handler:    _Chat_GetUsernames_Handler,
		},
		{
			MethodName: "SendMessage",
			Handler:    _Chat_SendMessage_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMessages",
			Handler:       _Chat_StreamMessages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "protobuf/chat.proto",
}
