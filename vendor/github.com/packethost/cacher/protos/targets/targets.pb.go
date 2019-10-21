// Code generated by protoc-gen-go. DO NOT EDIT.
// source: targets.proto

package targets

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PushRequest struct {
	Data                 string   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushRequest) Reset()         { *m = PushRequest{} }
func (m *PushRequest) String() string { return proto.CompactTextString(m) }
func (*PushRequest) ProtoMessage()    {}
func (*PushRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{0}
}

func (m *PushRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushRequest.Unmarshal(m, b)
}
func (m *PushRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushRequest.Marshal(b, m, deterministic)
}
func (m *PushRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushRequest.Merge(m, src)
}
func (m *PushRequest) XXX_Size() int {
	return xxx_messageInfo_PushRequest.Size(m)
}
func (m *PushRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PushRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PushRequest proto.InternalMessageInfo

func (m *PushRequest) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

type GetRequest struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{1}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

type UpdateRequest struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Data                 string   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateRequest) Reset()         { *m = UpdateRequest{} }
func (m *UpdateRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateRequest) ProtoMessage()    {}
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{2}
}

func (m *UpdateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateRequest.Unmarshal(m, b)
}
func (m *UpdateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateRequest.Marshal(b, m, deterministic)
}
func (m *UpdateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateRequest.Merge(m, src)
}
func (m *UpdateRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateRequest.Size(m)
}
func (m *UpdateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateRequest proto.InternalMessageInfo

func (m *UpdateRequest) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *UpdateRequest) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

type UUID struct {
	Uuid                 string   `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UUID) Reset()         { *m = UUID{} }
func (m *UUID) String() string { return proto.CompactTextString(m) }
func (*UUID) ProtoMessage()    {}
func (*UUID) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{3}
}

func (m *UUID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UUID.Unmarshal(m, b)
}
func (m *UUID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UUID.Marshal(b, m, deterministic)
}
func (m *UUID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UUID.Merge(m, src)
}
func (m *UUID) XXX_Size() int {
	return xxx_messageInfo_UUID.Size(m)
}
func (m *UUID) XXX_DiscardUnknown() {
	xxx_messageInfo_UUID.DiscardUnknown(m)
}

var xxx_messageInfo_UUID proto.InternalMessageInfo

func (m *UUID) GetUuid() string {
	if m != nil {
		return m.Uuid
	}
	return ""
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{4}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Targets struct {
	JSON                 string   `protobuf:"bytes,1,opt,name=JSON,proto3" json:"JSON,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Targets) Reset()         { *m = Targets{} }
func (m *Targets) String() string { return proto.CompactTextString(m) }
func (*Targets) ProtoMessage()    {}
func (*Targets) Descriptor() ([]byte, []int) {
	return fileDescriptor_4009e2e15debba2c, []int{5}
}

func (m *Targets) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Targets.Unmarshal(m, b)
}
func (m *Targets) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Targets.Marshal(b, m, deterministic)
}
func (m *Targets) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Targets.Merge(m, src)
}
func (m *Targets) XXX_Size() int {
	return xxx_messageInfo_Targets.Size(m)
}
func (m *Targets) XXX_DiscardUnknown() {
	xxx_messageInfo_Targets.DiscardUnknown(m)
}

var xxx_messageInfo_Targets proto.InternalMessageInfo

func (m *Targets) GetJSON() string {
	if m != nil {
		return m.JSON
	}
	return ""
}

func init() {
	proto.RegisterType((*PushRequest)(nil), "targets.PushRequest")
	proto.RegisterType((*GetRequest)(nil), "targets.GetRequest")
	proto.RegisterType((*UpdateRequest)(nil), "targets.UpdateRequest")
	proto.RegisterType((*UUID)(nil), "targets.UUID")
	proto.RegisterType((*Empty)(nil), "targets.Empty")
	proto.RegisterType((*Targets)(nil), "targets.Targets")
}

func init() { proto.RegisterFile("targets.proto", fileDescriptor_4009e2e15debba2c) }

var fileDescriptor_4009e2e15debba2c = []byte{
	// 244 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x49, 0x2c, 0x4a,
	0x4f, 0x2d, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0x95, 0x14, 0xb9,
	0xb8, 0x03, 0x4a, 0x8b, 0x33, 0x82, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0x84, 0xb8, 0x58,
	0x52, 0x12, 0x4b, 0x12, 0x25, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xc0, 0x6c, 0x25, 0x19, 0x2e,
	0x2e, 0xf7, 0xd4, 0x12, 0x98, 0x0a, 0x3e, 0x2e, 0x26, 0x4f, 0x17, 0xa8, 0x3c, 0x93, 0xa7, 0x8b,
	0x92, 0x31, 0x17, 0x6f, 0x68, 0x41, 0x4a, 0x62, 0x49, 0x2a, 0x0e, 0x05, 0x70, 0x23, 0x99, 0x90,
	0x8c, 0x94, 0xe2, 0x62, 0x09, 0x0d, 0x85, 0xc8, 0x95, 0x96, 0x66, 0xa6, 0xc0, 0xac, 0x03, 0xb1,
	0x95, 0xd8, 0xb9, 0x58, 0x5d, 0x73, 0x0b, 0x4a, 0x2a, 0x95, 0x64, 0xb9, 0xd8, 0x43, 0x20, 0xae,
	0x04, 0xa9, 0xf3, 0x0a, 0xf6, 0xf7, 0x83, 0xa9, 0x03, 0xb1, 0x8d, 0x5e, 0x30, 0x72, 0xb1, 0x41,
	0xe4, 0x85, 0x4c, 0xb8, 0x78, 0x9d, 0x8b, 0x52, 0x13, 0x4b, 0x52, 0x61, 0xea, 0x45, 0xf4, 0x60,
	0xde, 0x45, 0xf2, 0x9c, 0x14, 0x2f, 0x5c, 0x14, 0x6c, 0xb9, 0x31, 0x17, 0x17, 0x44, 0xbd, 0x53,
	0xa5, 0xa7, 0x8b, 0x90, 0x30, 0x5c, 0x12, 0xe1, 0x59, 0x29, 0x01, 0xb8, 0x20, 0xcc, 0x64, 0x2b,
	0x2e, 0x01, 0x88, 0x77, 0x91, 0xb4, 0x8a, 0x21, 0xcc, 0x45, 0x0e, 0x09, 0x29, 0x3e, 0xb8, 0x38,
	0xd8, 0x43, 0x42, 0xe6, 0x5c, 0x02, 0x2e, 0xa9, 0x39, 0xa9, 0x28, 0x7a, 0xb1, 0x5a, 0x8b, 0xa6,
	0x31, 0x89, 0x0d, 0x1c, 0x69, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x07, 0x97, 0x49, 0x08,
	0xc5, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TargetClient is the client API for Target service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TargetClient interface {
	CreateTargets(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*UUID, error)
	TargetByID(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Targets, error)
	UpdateTargetByID(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error)
	DeleteTargetByID(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Empty, error)
}

type targetClient struct {
	cc *grpc.ClientConn
}

func NewTargetClient(cc *grpc.ClientConn) TargetClient {
	return &targetClient{cc}
}

func (c *targetClient) CreateTargets(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*UUID, error) {
	out := new(UUID)
	err := c.cc.Invoke(ctx, "/targets.Target/CreateTargets", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *targetClient) TargetByID(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Targets, error) {
	out := new(Targets)
	err := c.cc.Invoke(ctx, "/targets.Target/TargetByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *targetClient) UpdateTargetByID(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/targets.Target/UpdateTargetByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *targetClient) DeleteTargetByID(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/targets.Target/DeleteTargetByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TargetServer is the server API for Target service.
type TargetServer interface {
	CreateTargets(context.Context, *PushRequest) (*UUID, error)
	TargetByID(context.Context, *GetRequest) (*Targets, error)
	UpdateTargetByID(context.Context, *UpdateRequest) (*Empty, error)
	DeleteTargetByID(context.Context, *GetRequest) (*Empty, error)
}

// UnimplementedTargetServer can be embedded to have forward compatible implementations.
type UnimplementedTargetServer struct {
}

func (*UnimplementedTargetServer) CreateTargets(ctx context.Context, req *PushRequest) (*UUID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTargets not implemented")
}
func (*UnimplementedTargetServer) TargetByID(ctx context.Context, req *GetRequest) (*Targets, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TargetByID not implemented")
}
func (*UnimplementedTargetServer) UpdateTargetByID(ctx context.Context, req *UpdateRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTargetByID not implemented")
}
func (*UnimplementedTargetServer) DeleteTargetByID(ctx context.Context, req *GetRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTargetByID not implemented")
}

func RegisterTargetServer(s *grpc.Server, srv TargetServer) {
	s.RegisterService(&_Target_serviceDesc, srv)
}

func _Target_CreateTargets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TargetServer).CreateTargets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/targets.Target/CreateTargets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TargetServer).CreateTargets(ctx, req.(*PushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Target_TargetByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TargetServer).TargetByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/targets.Target/TargetByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TargetServer).TargetByID(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Target_UpdateTargetByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TargetServer).UpdateTargetByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/targets.Target/UpdateTargetByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TargetServer).UpdateTargetByID(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Target_DeleteTargetByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TargetServer).DeleteTargetByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/targets.Target/DeleteTargetByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TargetServer).DeleteTargetByID(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Target_serviceDesc = grpc.ServiceDesc{
	ServiceName: "targets.Target",
	HandlerType: (*TargetServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTargets",
			Handler:    _Target_CreateTargets_Handler,
		},
		{
			MethodName: "TargetByID",
			Handler:    _Target_TargetByID_Handler,
		},
		{
			MethodName: "UpdateTargetByID",
			Handler:    _Target_UpdateTargetByID_Handler,
		},
		{
			MethodName: "DeleteTargetByID",
			Handler:    _Target_DeleteTargetByID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "targets.proto",
}
