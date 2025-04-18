// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: stats_users.proto

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

type StatsUsers struct {
	state                  protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_StatsUsers  uint64                 `protobuf:"varint,1,opt,name=stats_users,json=statsUsers"`
	XXX_raceDetectHookData protoimpl.RaceDetectHookData
	XXX_presence           [1]uint32
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *StatsUsers) Reset() {
	*x = StatsUsers{}
	mi := &file_stats_users_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatsUsers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsUsers) ProtoMessage() {}

func (x *StatsUsers) ProtoReflect() protoreflect.Message {
	mi := &file_stats_users_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *StatsUsers) GetStatsUsers() uint64 {
	if x != nil {
		return x.xxx_hidden_StatsUsers
	}
	return 0
}

func (x *StatsUsers) SetStatsUsers(v uint64) {
	x.xxx_hidden_StatsUsers = v
	protoimpl.X.SetPresent(&(x.XXX_presence[0]), 0, 1)
}

func (x *StatsUsers) HasStatsUsers() bool {
	if x == nil {
		return false
	}
	return protoimpl.X.Present(&(x.XXX_presence[0]), 0)
}

func (x *StatsUsers) ClearStatsUsers() {
	protoimpl.X.ClearPresent(&(x.XXX_presence[0]), 0)
	x.xxx_hidden_StatsUsers = 0
}

type StatsUsers_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	StatsUsers *uint64
}

func (b0 StatsUsers_builder) Build() *StatsUsers {
	m0 := &StatsUsers{}
	b, x := &b0, m0
	_, _ = b, x
	if b.StatsUsers != nil {
		protoimpl.X.SetPresentNonAtomic(&(x.XXX_presence[0]), 0, 1)
		x.xxx_hidden_StatsUsers = *b.StatsUsers
	}
	return m0
}

var File_stats_users_proto protoreflect.FileDescriptor

const file_stats_users_proto_rawDesc = "" +
	"\n" +
	"\x11stats_users.proto\x12\vproto.model\x1a!google/protobuf/go_features.proto\"-\n" +
	"\n" +
	"StatsUsers\x12\x1f\n" +
	"\vstats_users\x18\x01 \x01(\x04R\n" +
	"statsUsersBLZBgithub.com/RexArseny/url_shortener/internal/app/models/proto/model\x92\x03\x05\xd2>\x02\x10\x03b\beditionsp\xe8\a"

var file_stats_users_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_stats_users_proto_goTypes = []any{
	(*StatsUsers)(nil), // 0: proto.model.StatsUsers
}
var file_stats_users_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_stats_users_proto_init() }
func file_stats_users_proto_init() {
	if File_stats_users_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_stats_users_proto_rawDesc), len(file_stats_users_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_stats_users_proto_goTypes,
		DependencyIndexes: file_stats_users_proto_depIdxs,
		MessageInfos:      file_stats_users_proto_msgTypes,
	}.Build()
	File_stats_users_proto = out.File
	file_stats_users_proto_goTypes = nil
	file_stats_users_proto_depIdxs = nil
}
