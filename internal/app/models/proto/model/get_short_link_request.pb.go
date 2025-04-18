// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: get_short_link_request.proto

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

type GetShortLinkRequest struct {
	state         protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_Id *ID                    `protobuf:"bytes,1,opt,name=id"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetShortLinkRequest) Reset() {
	*x = GetShortLinkRequest{}
	mi := &file_get_short_link_request_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetShortLinkRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShortLinkRequest) ProtoMessage() {}

func (x *GetShortLinkRequest) ProtoReflect() protoreflect.Message {
	mi := &file_get_short_link_request_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *GetShortLinkRequest) GetId() *ID {
	if x != nil {
		return x.xxx_hidden_Id
	}
	return nil
}

func (x *GetShortLinkRequest) SetId(v *ID) {
	x.xxx_hidden_Id = v
}

func (x *GetShortLinkRequest) HasId() bool {
	if x == nil {
		return false
	}
	return x.xxx_hidden_Id != nil
}

func (x *GetShortLinkRequest) ClearId() {
	x.xxx_hidden_Id = nil
}

type GetShortLinkRequest_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	Id *ID
}

func (b0 GetShortLinkRequest_builder) Build() *GetShortLinkRequest {
	m0 := &GetShortLinkRequest{}
	b, x := &b0, m0
	_, _ = b, x
	x.xxx_hidden_Id = b.Id
	return m0
}

var File_get_short_link_request_proto protoreflect.FileDescriptor

const file_get_short_link_request_proto_rawDesc = "" +
	"\n" +
	"\x1cget_short_link_request.proto\x12\vproto.model\x1a\bid.proto\x1a!google/protobuf/go_features.proto\"6\n" +
	"\x13GetShortLinkRequest\x12\x1f\n" +
	"\x02id\x18\x01 \x01(\v2\x0f.proto.model.IDR\x02idBLZBgithub.com/RexArseny/url_shortener/internal/app/models/proto/model\x92\x03\x05\xd2>\x02\x10\x03b\beditionsp\xe8\a"

var file_get_short_link_request_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_get_short_link_request_proto_goTypes = []any{
	(*GetShortLinkRequest)(nil), // 0: proto.model.GetShortLinkRequest
	(*ID)(nil),                  // 1: proto.model.ID
}
var file_get_short_link_request_proto_depIdxs = []int32{
	1, // 0: proto.model.GetShortLinkRequest.id:type_name -> proto.model.ID
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_get_short_link_request_proto_init() }
func file_get_short_link_request_proto_init() {
	if File_get_short_link_request_proto != nil {
		return
	}
	file_id_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_get_short_link_request_proto_rawDesc), len(file_get_short_link_request_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_get_short_link_request_proto_goTypes,
		DependencyIndexes: file_get_short_link_request_proto_depIdxs,
		MessageInfos:      file_get_short_link_request_proto_msgTypes,
	}.Build()
	File_get_short_link_request_proto = out.File
	file_get_short_link_request_proto_goTypes = nil
	file_get_short_link_request_proto_depIdxs = nil
}
