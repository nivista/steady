// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.3
// source: protos/messaging.proto

package messaging

import (
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type Key struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domain    string `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	TimerUUID string `protobuf:"bytes,2,opt,name=timerUUID,proto3" json:"timerUUID,omitempty"`
}

func (x *Key) Reset() {
	*x = Key{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_messaging_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Key) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Key) ProtoMessage() {}

func (x *Key) ProtoReflect() protoreflect.Message {
	mi := &file_protos_messaging_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Key.ProtoReflect.Descriptor instead.
func (*Key) Descriptor() ([]byte, []int) {
	return file_protos_messaging_proto_rawDescGZIP(), []int{0}
}

func (x *Key) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *Key) GetTimerUUID() string {
	if x != nil {
		return x.TimerUUID
	}
	return ""
}

type Create struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Task     *common.Task     `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
	Schedule *common.Schedule `protobuf:"bytes,2,opt,name=schedule,proto3" json:"schedule,omitempty"`
	Meta     *common.Meta     `protobuf:"bytes,3,opt,name=meta,proto3" json:"meta,omitempty"`
}

func (x *Create) Reset() {
	*x = Create{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_messaging_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Create) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Create) ProtoMessage() {}

func (x *Create) ProtoReflect() protoreflect.Message {
	mi := &file_protos_messaging_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Create.ProtoReflect.Descriptor instead.
func (*Create) Descriptor() ([]byte, []int) {
	return file_protos_messaging_proto_rawDescGZIP(), []int{1}
}

func (x *Create) GetTask() *common.Task {
	if x != nil {
		return x.Task
	}
	return nil
}

func (x *Create) GetSchedule() *common.Schedule {
	if x != nil {
		return x.Schedule
	}
	return nil
}

func (x *Create) GetMeta() *common.Meta {
	if x != nil {
		return x.Meta
	}
	return nil
}

type Execute struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Progress *Progress `protobuf:"bytes,1,opt,name=progress,proto3" json:"progress,omitempty"`
	Result   []byte    `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *Execute) Reset() {
	*x = Execute{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_messaging_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Execute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Execute) ProtoMessage() {}

func (x *Execute) ProtoReflect() protoreflect.Message {
	mi := &file_protos_messaging_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Execute.ProtoReflect.Descriptor instead.
func (*Execute) Descriptor() ([]byte, []int) {
	return file_protos_messaging_proto_rawDescGZIP(), []int{2}
}

func (x *Execute) GetProgress() *Progress {
	if x != nil {
		return x.Progress
	}
	return nil
}

func (x *Execute) GetResult() []byte {
	if x != nil {
		return x.Result
	}
	return nil
}

type Progress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CompletedExecutions int32                `protobuf:"varint,1,opt,name=completedExecutions,proto3" json:"completedExecutions,omitempty"`
	LastExecution       *timestamp.Timestamp `protobuf:"bytes,2,opt,name=lastExecution,proto3" json:"lastExecution,omitempty"`
}

func (x *Progress) Reset() {
	*x = Progress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_messaging_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Progress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Progress) ProtoMessage() {}

func (x *Progress) ProtoReflect() protoreflect.Message {
	mi := &file_protos_messaging_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Progress.ProtoReflect.Descriptor instead.
func (*Progress) Descriptor() ([]byte, []int) {
	return file_protos_messaging_proto_rawDescGZIP(), []int{3}
}

func (x *Progress) GetCompletedExecutions() int32 {
	if x != nil {
		return x.CompletedExecutions
	}
	return 0
}

func (x *Progress) GetLastExecution() *timestamp.Timestamp {
	if x != nil {
		return x.LastExecution
	}
	return nil
}

var File_protos_messaging_proto protoreflect.FileDescriptor

var file_protos_messaging_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69,
	0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3b,
	0x0a, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x1c, 0x0a,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x72, 0x55, 0x55, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x72, 0x55, 0x55, 0x49, 0x44, 0x22, 0x65, 0x0a, 0x06, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x19, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b,
	0x12, 0x25, 0x0a, 0x08, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x09, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x08, 0x73,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x19, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x04, 0x6d, 0x65,
	0x74, 0x61, 0x22, 0x48, 0x0a, 0x07, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x12, 0x25, 0x0a,
	0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x09, 0x2e, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x67,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x7e, 0x0a, 0x08,
	0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x30, 0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x70,
	0x6c, 0x65, 0x74, 0x65, 0x64, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x13, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x64,
	0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x40, 0x0a, 0x0d, 0x6c, 0x61,
	0x73, 0x74, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0d, 0x6c,
	0x61, 0x73, 0x74, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x3a, 0x5a, 0x38,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x69, 0x76, 0x69, 0x73,
	0x74, 0x61, 0x2f, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x2e, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_messaging_proto_rawDescOnce sync.Once
	file_protos_messaging_proto_rawDescData = file_protos_messaging_proto_rawDesc
)

func file_protos_messaging_proto_rawDescGZIP() []byte {
	file_protos_messaging_proto_rawDescOnce.Do(func() {
		file_protos_messaging_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_messaging_proto_rawDescData)
	})
	return file_protos_messaging_proto_rawDescData
}

var file_protos_messaging_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_protos_messaging_proto_goTypes = []interface{}{
	(*Key)(nil),                 // 0: Key
	(*Create)(nil),              // 1: Create
	(*Execute)(nil),             // 2: Execute
	(*Progress)(nil),            // 3: Progress
	(*common.Task)(nil),         // 4: Task
	(*common.Schedule)(nil),     // 5: Schedule
	(*common.Meta)(nil),         // 6: Meta
	(*timestamp.Timestamp)(nil), // 7: google.protobuf.Timestamp
}
var file_protos_messaging_proto_depIdxs = []int32{
	4, // 0: Create.task:type_name -> Task
	5, // 1: Create.schedule:type_name -> Schedule
	6, // 2: Create.meta:type_name -> Meta
	3, // 3: Execute.progress:type_name -> Progress
	7, // 4: Progress.lastExecution:type_name -> google.protobuf.Timestamp
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_protos_messaging_proto_init() }
func file_protos_messaging_proto_init() {
	if File_protos_messaging_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_messaging_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Key); i {
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
		file_protos_messaging_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Create); i {
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
		file_protos_messaging_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Execute); i {
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
		file_protos_messaging_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Progress); i {
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
			RawDescriptor: file_protos_messaging_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_messaging_proto_goTypes,
		DependencyIndexes: file_protos_messaging_proto_depIdxs,
		MessageInfos:      file_protos_messaging_proto_msgTypes,
	}.Build()
	File_protos_messaging_proto = out.File
	file_protos_messaging_proto_rawDesc = nil
	file_protos_messaging_proto_goTypes = nil
	file_protos_messaging_proto_depIdxs = nil
}
