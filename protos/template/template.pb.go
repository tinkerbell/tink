// Code generated by protoc-gen-go. DO NOT EDIT.
// source: template.proto

package template

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1b68e1b5f001c74, []int{0}
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

type WorkflowTemplate struct {
	Id                   string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                 string               `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Data                 []byte               `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	InsertedAt           *timestamp.Timestamp `protobuf:"bytes,4,opt,name=insertedAt,proto3" json:"insertedAt,omitempty"`
	DeletedAt            *timestamp.Timestamp `protobuf:"bytes,5,opt,name=deletedAt,proto3" json:"deletedAt,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *WorkflowTemplate) Reset()         { *m = WorkflowTemplate{} }
func (m *WorkflowTemplate) String() string { return proto.CompactTextString(m) }
func (*WorkflowTemplate) ProtoMessage()    {}
func (*WorkflowTemplate) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1b68e1b5f001c74, []int{1}
}

func (m *WorkflowTemplate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WorkflowTemplate.Unmarshal(m, b)
}
func (m *WorkflowTemplate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WorkflowTemplate.Marshal(b, m, deterministic)
}
func (m *WorkflowTemplate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WorkflowTemplate.Merge(m, src)
}
func (m *WorkflowTemplate) XXX_Size() int {
	return xxx_messageInfo_WorkflowTemplate.Size(m)
}
func (m *WorkflowTemplate) XXX_DiscardUnknown() {
	xxx_messageInfo_WorkflowTemplate.DiscardUnknown(m)
}

var xxx_messageInfo_WorkflowTemplate proto.InternalMessageInfo

func (m *WorkflowTemplate) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *WorkflowTemplate) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *WorkflowTemplate) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *WorkflowTemplate) GetInsertedAt() *timestamp.Timestamp {
	if m != nil {
		return m.InsertedAt
	}
	return nil
}

func (m *WorkflowTemplate) GetDeletedAt() *timestamp.Timestamp {
	if m != nil {
		return m.DeletedAt
	}
	return nil
}

type CreateResponse struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateResponse) Reset()         { *m = CreateResponse{} }
func (m *CreateResponse) String() string { return proto.CompactTextString(m) }
func (*CreateResponse) ProtoMessage()    {}
func (*CreateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1b68e1b5f001c74, []int{2}
}

func (m *CreateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateResponse.Unmarshal(m, b)
}
func (m *CreateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateResponse.Marshal(b, m, deterministic)
}
func (m *CreateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateResponse.Merge(m, src)
}
func (m *CreateResponse) XXX_Size() int {
	return xxx_messageInfo_CreateResponse.Size(m)
}
func (m *CreateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CreateResponse proto.InternalMessageInfo

func (m *CreateResponse) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_b1b68e1b5f001c74, []int{3}
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

func (m *GetRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func init() {
	proto.RegisterType((*Empty)(nil), "template.Empty")
	proto.RegisterType((*WorkflowTemplate)(nil), "template.WorkflowTemplate")
	proto.RegisterType((*CreateResponse)(nil), "template.CreateResponse")
	proto.RegisterType((*GetRequest)(nil), "template.GetRequest")
}

func init() { proto.RegisterFile("template.proto", fileDescriptor_b1b68e1b5f001c74) }

var fileDescriptor_b1b68e1b5f001c74 = []byte{
	// 279 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x51, 0x4d, 0x4b, 0xc4, 0x30,
	0x10, 0x25, 0xfb, 0x51, 0x77, 0x47, 0xa9, 0x12, 0x3c, 0x94, 0x22, 0x58, 0x7a, 0xea, 0xa9, 0x85,
	0xf5, 0xa0, 0x78, 0x10, 0x44, 0x65, 0xef, 0x65, 0xc1, 0x73, 0x96, 0xcc, 0x2e, 0xc5, 0xa6, 0x89,
	0xcd, 0x2c, 0xe2, 0xff, 0xf2, 0x07, 0xf8, 0xd3, 0xc4, 0xd4, 0x6e, 0xa4, 0xa2, 0xde, 0x26, 0x6f,
	0xde, 0x9b, 0x79, 0xf3, 0x02, 0x21, 0xa1, 0x32, 0xb5, 0x20, 0xcc, 0x4d, 0xab, 0x49, 0xf3, 0x59,
	0xff, 0x8e, 0xcf, 0xb7, 0x5a, 0x6f, 0x6b, 0x2c, 0x1c, 0xbe, 0xde, 0x6d, 0x0a, 0xaa, 0x14, 0x5a,
	0x12, 0xca, 0x74, 0xd4, 0xf4, 0x00, 0xa6, 0x0f, 0xca, 0xd0, 0x6b, 0xfa, 0xce, 0xe0, 0xe4, 0x51,
	0xb7, 0x4f, 0x9b, 0x5a, 0xbf, 0xac, 0xbe, 0xe4, 0x3c, 0x84, 0x51, 0x25, 0x23, 0x96, 0xb0, 0x6c,
	0x5e, 0x8e, 0x2a, 0xc9, 0x39, 0x4c, 0x1a, 0xa1, 0x30, 0x1a, 0x39, 0xc4, 0xd5, 0x9f, 0x98, 0x14,
	0x24, 0xa2, 0x71, 0xc2, 0xb2, 0xa3, 0xd2, 0xd5, 0xfc, 0x1a, 0xa0, 0x6a, 0x2c, 0xb6, 0x84, 0xf2,
	0x96, 0xa2, 0x49, 0xc2, 0xb2, 0xc3, 0x45, 0x9c, 0x77, 0x5e, 0xf2, 0xde, 0x4b, 0xbe, 0xea, 0xbd,
	0x94, 0xdf, 0xd8, 0xfc, 0x0a, 0xe6, 0x12, 0x6b, 0xec, 0xa4, 0xd3, 0x7f, 0xa5, 0x9e, 0x9c, 0x26,
	0x10, 0xde, 0xb5, 0x28, 0x08, 0x4b, 0xb4, 0x46, 0x37, 0xf6, 0x87, 0xff, 0xf4, 0x0c, 0x60, 0x89,
	0x54, 0xe2, 0xf3, 0x0e, 0x2d, 0x0d, 0xbb, 0x8b, 0x37, 0x06, 0xb3, 0xfd, 0xe9, 0x37, 0x10, 0x74,
	0xc3, 0x78, 0x9c, 0xef, 0xe3, 0x1d, 0x06, 0x14, 0x47, 0xbe, 0x37, 0x58, 0x7d, 0x09, 0xe3, 0x25,
	0x12, 0x3f, 0xf5, 0x04, 0xbf, 0x39, 0xfe, 0x63, 0x24, 0x2f, 0x20, 0xb8, 0x77, 0x27, 0xfd, 0xa2,
	0x3d, 0xf6, 0xa8, 0xfb, 0xb9, 0x75, 0xe0, 0x52, 0xb9, 0xf8, 0x08, 0x00, 0x00, 0xff, 0xff, 0xbf,
	0x0d, 0x91, 0x14, 0x06, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TemplateClient is the client API for Template service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TemplateClient interface {
	Create(ctx context.Context, in *WorkflowTemplate, opts ...grpc.CallOption) (*CreateResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*WorkflowTemplate, error)
	Delete(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Empty, error)
}

type templateClient struct {
	cc *grpc.ClientConn
}

func NewTemplateClient(cc *grpc.ClientConn) TemplateClient {
	return &templateClient{cc}
}

func (c *templateClient) Create(ctx context.Context, in *WorkflowTemplate, opts ...grpc.CallOption) (*CreateResponse, error) {
	out := new(CreateResponse)
	err := c.cc.Invoke(ctx, "/template.Template/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *templateClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*WorkflowTemplate, error) {
	out := new(WorkflowTemplate)
	err := c.cc.Invoke(ctx, "/template.Template/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *templateClient) Delete(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/template.Template/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TemplateServer is the server API for Template service.
type TemplateServer interface {
	Create(context.Context, *WorkflowTemplate) (*CreateResponse, error)
	Get(context.Context, *GetRequest) (*WorkflowTemplate, error)
	Delete(context.Context, *GetRequest) (*Empty, error)
}

// UnimplementedTemplateServer can be embedded to have forward compatible implementations.
type UnimplementedTemplateServer struct {
}

func (*UnimplementedTemplateServer) Create(ctx context.Context, req *WorkflowTemplate) (*CreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (*UnimplementedTemplateServer) Get(ctx context.Context, req *GetRequest) (*WorkflowTemplate, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedTemplateServer) Delete(ctx context.Context, req *GetRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func RegisterTemplateServer(s *grpc.Server, srv TemplateServer) {
	s.RegisterService(&_Template_serviceDesc, srv)
}

func _Template_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkflowTemplate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TemplateServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/template.Template/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TemplateServer).Create(ctx, req.(*WorkflowTemplate))
	}
	return interceptor(ctx, in, info, handler)
}

func _Template_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TemplateServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/template.Template/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TemplateServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Template_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TemplateServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/template.Template/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TemplateServer).Delete(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Template_serviceDesc = grpc.ServiceDesc{
	ServiceName: "template.Template",
	HandlerType: (*TemplateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _Template_Create_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Template_Get_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Template_Delete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "template.proto",
}
