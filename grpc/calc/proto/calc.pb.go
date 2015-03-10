// Code generated by protoc-gen-go.
// source: calc.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	calc.proto

It has these top-level messages:
	Request
	Response
*/
package proto

import proto1 "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal

type Request struct {
	X float32 `protobuf:"fixed32,1,opt,name=x" json:"x,omitempty"`
	Y float32 `protobuf:"fixed32,2,opt,name=y" json:"y,omitempty"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto1.CompactTextString(m) }
func (*Request) ProtoMessage()    {}

type Response struct {
	Z float32 `protobuf:"fixed32,1,opt,name=z" json:"z,omitempty"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto1.CompactTextString(m) }
func (*Response) ProtoMessage()    {}

func init() {
}

// Client API for Calc service

type CalcClient interface {
	Add(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type calcClient struct {
	cc *grpc.ClientConn
}

func NewCalcClient(cc *grpc.ClientConn) CalcClient {
	return &calcClient{cc}
}

func (c *calcClient) Add(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.calc/Add", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Calc service

type CalcServer interface {
	Add(context.Context, *Request) (*Response, error)
}

func RegisterCalcServer(s *grpc.Server, srv CalcServer) {
	s.RegisterService(&_Calc_serviceDesc, srv)
}

func _Calc_Add_Handler(srv interface{}, ctx context.Context, buf []byte) (proto1.Message, error) {
	in := new(Request)
	if err := proto1.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CalcServer).Add(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Calc_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.calc",
	HandlerType: (*CalcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Add",
			Handler:    _Calc_Add_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
