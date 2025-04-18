// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: create_short_link_json_batch_request.proto

package model

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/gofeaturespb"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateShortLinkJSONBatchRequest struct {
	state               protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Requests *[]*BatchRequest       `protobuf:"bytes,1,rep,name=requests"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *CreateShortLinkJSONBatchRequest) Reset() {
	*x = CreateShortLinkJSONBatchRequest{}
	mi := &file_create_short_link_json_batch_request_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateShortLinkJSONBatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateShortLinkJSONBatchRequest) ProtoMessage() {}

func (x *CreateShortLinkJSONBatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_create_short_link_json_batch_request_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *CreateShortLinkJSONBatchRequest) GetRequests() []*BatchRequest {
	if x != nil {
		if x.xxx_hidden_Requests != nil {
			return *x.xxx_hidden_Requests
		}
	}
	return nil
}

func (x *CreateShortLinkJSONBatchRequest) SetRequests(v []*BatchRequest) {
	x.xxx_hidden_Requests = &v
}

type CreateShortLinkJSONBatchRequest_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Requests []*BatchRequest
}

func (b0 CreateShortLinkJSONBatchRequest_builder) Build() *CreateShortLinkJSONBatchRequest {
	m0 := &CreateShortLinkJSONBatchRequest{}
	b, x := &b0, m0
	_, _ = b, x
	x.xxx_hidden_Requests = &b.Requests
	return m0
}

var File_create_short_link_json_batch_request_proto protoreflect.FileDescriptor

const file_create_short_link_json_batch_request_proto_rawDesc = "" +
	"\n" +
	"*create_short_link_json_batch_request.proto\x12\vproto.model\x1a\x13batch_request.proto\x1a!google/protobuf/go_features.proto\"X\n" +
	"\x1fCreateShortLinkJSONBatchRequest\x125\n" +
	"\brequests\x18\x01 \x03(\v2\x19.proto.model.BatchRequestR\brequestsBLZBgithub.com/RexArseny/url_shortener/internal/app/models/proto/model\x92\x03\x05\xd2>\x02\x10\x03b\beditionsp\xe8\a"

var file_create_short_link_json_batch_request_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_create_short_link_json_batch_request_proto_goTypes = []any{
	(*CreateShortLinkJSONBatchRequest)(nil), // 0: proto.model.CreateShortLinkJSONBatchRequest
	(*BatchRequest)(nil),                    // 1: proto.model.BatchRequest
}
var file_create_short_link_json_batch_request_proto_depIdxs = []int32{
	1, // 0: proto.model.CreateShortLinkJSONBatchRequest.requests:type_name -> proto.model.BatchRequest
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_create_short_link_json_batch_request_proto_init() }
func file_create_short_link_json_batch_request_proto_init() {
	if File_create_short_link_json_batch_request_proto != nil {
		return
	}
	file_batch_request_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_create_short_link_json_batch_request_proto_rawDesc), len(file_create_short_link_json_batch_request_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_create_short_link_json_batch_request_proto_goTypes,
		DependencyIndexes: file_create_short_link_json_batch_request_proto_depIdxs,
		MessageInfos:      file_create_short_link_json_batch_request_proto_msgTypes,
	}.Build()
	File_create_short_link_json_batch_request_proto = out.File
	file_create_short_link_json_batch_request_proto_goTypes = nil
	file_create_short_link_json_batch_request_proto_depIdxs = nil
}
