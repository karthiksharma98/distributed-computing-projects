// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

/*
Package service is a generated protocol buffer package.

It is generated from these files:
	service.proto

It has these top-level messages:
	UploadRequest
	UploadReply
	DownloadRequest
	DownloadReply
*/
package service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type UploadRequest struct {
	FileContents []byte `protobuf:"bytes,1,opt,name=fileContents,proto3" json:"fileContents,omitempty"`
	SdfsFileName string `protobuf:"bytes,2,opt,name=sdfsFileName" json:"sdfsFileName,omitempty"`
}

func (m *UploadRequest) Reset()                    { *m = UploadRequest{} }
func (m *UploadRequest) String() string            { return proto.CompactTextString(m) }
func (*UploadRequest) ProtoMessage()               {}
func (*UploadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *UploadRequest) GetFileContents() []byte {
	if m != nil {
		return m.FileContents
	}
	return nil
}

func (m *UploadRequest) GetSdfsFileName() string {
	if m != nil {
		return m.SdfsFileName
	}
	return ""
}

type UploadReply struct {
	Status bool `protobuf:"varint,1,opt,name=status" json:"status,omitempty"`
}

func (m *UploadReply) Reset()                    { *m = UploadReply{} }
func (m *UploadReply) String() string            { return proto.CompactTextString(m) }
func (*UploadReply) ProtoMessage()               {}
func (*UploadReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *UploadReply) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

type DownloadRequest struct {
	SdfsFileName string `protobuf:"bytes,1,opt,name=sdfsFileName" json:"sdfsFileName,omitempty"`
}

func (m *DownloadRequest) Reset()                    { *m = DownloadRequest{} }
func (m *DownloadRequest) String() string            { return proto.CompactTextString(m) }
func (*DownloadRequest) ProtoMessage()               {}
func (*DownloadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *DownloadRequest) GetSdfsFileName() string {
	if m != nil {
		return m.SdfsFileName
	}
	return ""
}

type DownloadReply struct {
	DoesFileExist bool   `protobuf:"varint,1,opt,name=doesFileExist" json:"doesFileExist,omitempty"`
	FileContents  []byte `protobuf:"bytes,2,opt,name=fileContents,proto3" json:"fileContents,omitempty"`
}

func (m *DownloadReply) Reset()                    { *m = DownloadReply{} }
func (m *DownloadReply) String() string            { return proto.CompactTextString(m) }
func (*DownloadReply) ProtoMessage()               {}
func (*DownloadReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DownloadReply) GetDoesFileExist() bool {
	if m != nil {
		return m.DoesFileExist
	}
	return false
}

func (m *DownloadReply) GetFileContents() []byte {
	if m != nil {
		return m.FileContents
	}
	return nil
}

func init() {
	proto.RegisterType((*UploadRequest)(nil), "service.UploadRequest")
	proto.RegisterType((*UploadReply)(nil), "service.UploadReply")
	proto.RegisterType((*DownloadRequest)(nil), "service.DownloadRequest")
	proto.RegisterType((*DownloadReply)(nil), "service.DownloadReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for FileTransfer service

type FileTransferClient interface {
	// Upload a file to another location
	Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadReply, error)
	// Download a file from another location
	Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadReply, error)
}

type fileTransferClient struct {
	cc *grpc.ClientConn
}

func NewFileTransferClient(cc *grpc.ClientConn) FileTransferClient {
	return &fileTransferClient{cc}
}

func (c *fileTransferClient) Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadReply, error) {
	out := new(UploadReply)
	err := grpc.Invoke(ctx, "/service.FileTransfer/Upload", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileTransferClient) Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadReply, error) {
	out := new(DownloadReply)
	err := grpc.Invoke(ctx, "/service.FileTransfer/Download", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for FileTransfer service

type FileTransferServer interface {
	// Upload a file to another location
	Upload(context.Context, *UploadRequest) (*UploadReply, error)
	// Download a file from another location
	Download(context.Context, *DownloadRequest) (*DownloadReply, error)
}

func RegisterFileTransferServer(s *grpc.Server, srv FileTransferServer) {
	s.RegisterService(&_FileTransfer_serviceDesc, srv)
}

func _FileTransfer_Upload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileTransferServer).Upload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.FileTransfer/Upload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileTransferServer).Upload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileTransfer_Download_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileTransferServer).Download(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.FileTransfer/Download",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileTransferServer).Download(ctx, req.(*DownloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FileTransfer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.FileTransfer",
	HandlerType: (*FileTransferServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Upload",
			Handler:    _FileTransfer_Upload_Handler,
		},
		{
			MethodName: "Download",
			Handler:    _FileTransfer_Download_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}

func init() { proto.RegisterFile("service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x87, 0x72, 0x95, 0xc2, 0xb9,
	0x78, 0x43, 0x0b, 0x72, 0xf2, 0x13, 0x53, 0x82, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0x94,
	0xb8, 0x78, 0xd2, 0x32, 0x73, 0x52, 0x9d, 0xf3, 0xf3, 0x4a, 0x52, 0xf3, 0x4a, 0x8a, 0x25, 0x18,
	0x15, 0x18, 0x35, 0x78, 0x82, 0x50, 0xc4, 0x40, 0x6a, 0x8a, 0x53, 0xd2, 0x8a, 0xdd, 0x32, 0x73,
	0x52, 0xfd, 0x12, 0x73, 0x53, 0x25, 0x98, 0x14, 0x18, 0x35, 0x38, 0x83, 0x50, 0xc4, 0x94, 0x54,
	0xb9, 0xb8, 0x61, 0x06, 0x17, 0xe4, 0x54, 0x0a, 0x89, 0x71, 0xb1, 0x15, 0x97, 0x24, 0x96, 0x94,
	0x42, 0x0c, 0xe4, 0x08, 0x82, 0xf2, 0x94, 0x4c, 0xb9, 0xf8, 0x5d, 0xf2, 0xcb, 0xf3, 0xd0, 0x5c,
	0x80, 0x62, 0x3a, 0x23, 0x16, 0xd3, 0x23, 0xb9, 0x78, 0x11, 0xda, 0x40, 0xe6, 0xab, 0x70, 0xf1,
	0xa6, 0xe4, 0xa7, 0x82, 0x15, 0xb8, 0x56, 0x64, 0x16, 0x97, 0x40, 0xad, 0x41, 0x15, 0xc4, 0xf0,
	0x1c, 0x13, 0xa6, 0xe7, 0x8c, 0x3a, 0x18, 0xb9, 0x78, 0x40, 0x3a, 0x42, 0x8a, 0x12, 0xf3, 0x8a,
	0xd3, 0x52, 0x8b, 0x84, 0x2c, 0xb8, 0xd8, 0x20, 0x3e, 0x11, 0x12, 0xd3, 0x83, 0x85, 0x22, 0x4a,
	0x98, 0x49, 0x89, 0x60, 0x88, 0x17, 0xe4, 0x54, 0x2a, 0x31, 0x08, 0xd9, 0x71, 0x71, 0xc0, 0x5c,
	0x29, 0x24, 0x01, 0x57, 0x83, 0xe6, 0x5f, 0x29, 0x31, 0x2c, 0x32, 0x60, 0xfd, 0x49, 0x6c, 0xe0,
	0xc8, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x29, 0x1c, 0xb7, 0xfb, 0xbd, 0x01, 0x00, 0x00,
}
