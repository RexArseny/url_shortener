// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: user_urls.proto

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

type UserURLs struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_ShortUrl    *string                `protobuf:"bytes,1,opt,name=short_url,json=shortUrl"`
	xxx_hidden_OriginalUrl *string                `protobuf:"bytes,2,opt,name=original_url,json=originalUrl"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *UserURLs) Reset() {
	*x = UserURLs{}
	mi := &file_user_urls_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserURLs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserURLs) ProtoMessage() {}

func (x *UserURLs) ProtoReflect() protoreflect.Message {
	mi := &file_user_urls_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *UserURLs) GetShortUrl() string {
	if x != nil {
		if x.xxx_hidden_ShortUrl != nil {
			return *x.xxx_hidden_ShortUrl
		}
		return ""
	}
	return ""
}

func (x *UserURLs) GetOriginalUrl() string {
	if x != nil {
		if x.xxx_hidden_OriginalUrl != nil {
			return *x.xxx_hidden_OriginalUrl
		}
		return ""
	}
	return ""
}

func (x *UserURLs) SetShortUrl(v string) {
	x.xxx_hidden_ShortUrl = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 0, 2)
}

func (x *UserURLs) SetOriginalUrl(v string) {
	x.xxx_hidden_OriginalUrl = &v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 1, 2)
}

func (x *UserURLs) HasShortUrl() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 0)
}

func (x *UserURLs) HasOriginalUrl() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 1)
}

func (x *UserURLs) ClearShortUrl() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 0)
	x.xxx_hidden_ShortUrl = nil
}

func (x *UserURLs) ClearOriginalUrl() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 1)
	x.xxx_hidden_OriginalUrl = nil
}

type UserURLs_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	ShortUrl    *string
	OriginalUrl *string
}

func (b0 UserURLs_builder) Build() *UserURLs {
	m0 := &UserURLs{}
	b, x := &b0, m0
	_, _ = b, x
	if b.ShortUrl != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 0, 2)
		x.xxx_hidden_ShortUrl = b.ShortUrl
	}
	if b.OriginalUrl != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 1, 2)
		x.xxx_hidden_OriginalUrl = b.OriginalUrl
	}
	return m0
}

var File_user_urls_proto protoreflect.FileDescriptor

const file_user_urls_proto_rawDesc = "" +
	"\n" +
	"\x0fuser_urls.proto\x12\vproto.model\x1a!google/protobuf/go_features.proto\"J\n" +
	"\bUserURLs\x12\x1b\n" +
	"\tshort_url\x18\x01 \x01(\tR\bshortUrl\x12!\n" +
	"\foriginal_url\x18\x02 \x01(\tR\voriginalUrlBLZBgithub.com/RexArseny/url_shortener/internal/app/models/proto/model\x92\x03\x05\xd2>\x02\x10\x03b\beditionsp\xe8\a"

var file_user_urls_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_user_urls_proto_goTypes = []any{
	(*UserURLs)(nil), // 0: proto.model.UserURLs
}
var file_user_urls_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_user_urls_proto_init() }
func file_user_urls_proto_init() {
	if File_user_urls_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_user_urls_proto_rawDesc), len(file_user_urls_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_user_urls_proto_goTypes,
		DependencyIndexes: file_user_urls_proto_depIdxs,
		MessageInfos:      file_user_urls_proto_msgTypes,
	}.Build()
	File_user_urls_proto = out.File
	file_user_urls_proto_goTypes = nil
	file_user_urls_proto_depIdxs = nil
}
