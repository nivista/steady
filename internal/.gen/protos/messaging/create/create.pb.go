// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.3
// source: protos/messaging/create.proto

package create

import (
	proto "github.com/golang/protobuf/proto"
	common "github.com/nivista/steady/.gen/protos/common"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Task     *common.Task     `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
	Schedule *common.Schedule `protobuf:"bytes,2,opt,name=schedule,proto3" json:"schedule,omitempty"`
	Meta     *common.Meta     `protobuf:"bytes,3,opt,name=meta,proto3" json:"meta,omitempty"`
}

func (x *Value) Reset() {
	*x = Value{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_messaging_create_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Value) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Value) ProtoMessage() {}

func (x *Value) ProtoReflect() protoreflect.Message {
	mi := &file_protos_messaging_create_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Value.ProtoReflect.Descriptor instead.
func (*Value) Descriptor() ([]byte, []int) {
	return file_protos_messaging_create_proto_rawDescGZIP(), []int{0}
}

func (x *Value) GetTask() *common.Task {
	if x != nil {
		return x.Task
	}
	return nil
}

func (x *Value) GetSchedule() *common.Schedule {
	if x != nil {
		return x.Schedule
	}
	return nil
}

func (x *Value) GetMeta() *common.Meta {
	if x != nil {
		return x.Meta
	}
	return nil
}

var File_protos_messaging_create_proto protoreflect.FileDescriptor

var file_protos_messaging_create_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69,
	0x6e, 0x67, 0x2f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x10, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2e, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x1a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x64, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x19, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e,
	0x54, 0x61, 0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x25, 0x0a, 0x08, 0x73, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x53,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x08, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x65, 0x12, 0x19, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x05, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x42, 0x41, 0x5a, 0x3f,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x69, 0x76, 0x69, 0x73,
	0x74, 0x61, 0x2f, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x2e, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_messaging_create_proto_rawDescOnce sync.Once
	file_protos_messaging_create_proto_rawDescData = file_protos_messaging_create_proto_rawDesc
)

func file_protos_messaging_create_proto_rawDescGZIP() []byte {
	file_protos_messaging_create_proto_rawDescOnce.Do(func() {
		file_protos_messaging_create_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_messaging_create_proto_rawDescData)
	})
	return file_protos_messaging_create_proto_rawDescData
}

var file_protos_messaging_create_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_messaging_create_proto_goTypes = []interface{}{
	(*Value)(nil),           // 0: messaging.create.Value
	(*common.Task)(nil),     // 1: Task
	(*common.Schedule)(nil), // 2: Schedule
	(*common.Meta)(nil),     // 3: Meta
}
var file_protos_messaging_create_proto_depIdxs = []int32{
	1, // 0: messaging.create.Value.task:type_name -> Task
	2, // 1: messaging.create.Value.schedule:type_name -> Schedule
	3, // 2: messaging.create.Value.meta:type_name -> Meta
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_messaging_create_proto_init() }
func file_protos_messaging_create_proto_init() {
	if File_protos_messaging_create_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_messaging_create_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Value); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protos_messaging_create_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_messaging_create_proto_goTypes,
		DependencyIndexes: file_protos_messaging_create_proto_depIdxs,
		MessageInfos:      file_protos_messaging_create_proto_msgTypes,
	}.Build()
	File_protos_messaging_create_proto = out.File
	file_protos_messaging_create_proto_rawDesc = nil
	file_protos_messaging_create_proto_goTypes = nil
	file_protos_messaging_create_proto_depIdxs = nil
}
