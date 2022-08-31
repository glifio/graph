// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.0
// source: pkg/graph/data.proto

package graph

import (
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

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Height      uint64 `protobuf:"varint,10,opt,name=Height,proto3" json:"Height,omitempty"`
	MessageCbor []byte `protobuf:"bytes,13,opt,name=MessageCbor,proto3" json:"MessageCbor,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_graph_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_graph_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_pkg_graph_data_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Message) GetMessageCbor() []byte {
	if x != nil {
		return x.MessageCbor
	}
	return nil
}

type TipsetMessages struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cids [][]byte `protobuf:"bytes,1,rep,name=cids,proto3" json:"cids,omitempty"`
}

func (x *TipsetMessages) Reset() {
	*x = TipsetMessages{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_graph_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TipsetMessages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TipsetMessages) ProtoMessage() {}

func (x *TipsetMessages) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_graph_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TipsetMessages.ProtoReflect.Descriptor instead.
func (*TipsetMessages) Descriptor() ([]byte, []int) {
	return file_pkg_graph_data_proto_rawDescGZIP(), []int{1}
}

func (x *TipsetMessages) GetCids() [][]byte {
	if x != nil {
		return x.Cids
	}
	return nil
}

type Address struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Address []byte `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *Address) Reset() {
	*x = Address{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_graph_data_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Address) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Address) ProtoMessage() {}

func (x *Address) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_graph_data_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Address.ProtoReflect.Descriptor instead.
func (*Address) Descriptor() ([]byte, []int) {
	return file_pkg_graph_data_proto_rawDescGZIP(), []int{2}
}

func (x *Address) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Address) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

var File_pkg_graph_data_proto protoreflect.FileDescriptor

var file_pkg_graph_data_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x67, 0x72, 0x61, 0x70, 0x68, 0x22, 0x43, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x20, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x62, 0x6f, 0x72, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x62,
	0x6f, 0x72, 0x22, 0x24, 0x0a, 0x0e, 0x54, 0x69, 0x70, 0x73, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x04, 0x63, 0x69, 0x64, 0x73, 0x22, 0x33, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x42, 0x0c, 0x5a,
	0x0a, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x61, 0x70, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_pkg_graph_data_proto_rawDescOnce sync.Once
	file_pkg_graph_data_proto_rawDescData = file_pkg_graph_data_proto_rawDesc
)

func file_pkg_graph_data_proto_rawDescGZIP() []byte {
	file_pkg_graph_data_proto_rawDescOnce.Do(func() {
		file_pkg_graph_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_graph_data_proto_rawDescData)
	})
	return file_pkg_graph_data_proto_rawDescData
}

var file_pkg_graph_data_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pkg_graph_data_proto_goTypes = []interface{}{
	(*Message)(nil),        // 0: graph.Message
	(*TipsetMessages)(nil), // 1: graph.TipsetMessages
	(*Address)(nil),        // 2: graph.Address
}
var file_pkg_graph_data_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_graph_data_proto_init() }
func file_pkg_graph_data_proto_init() {
	if File_pkg_graph_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_graph_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_pkg_graph_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TipsetMessages); i {
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
		file_pkg_graph_data_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Address); i {
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
			RawDescriptor: file_pkg_graph_data_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_graph_data_proto_goTypes,
		DependencyIndexes: file_pkg_graph_data_proto_depIdxs,
		MessageInfos:      file_pkg_graph_data_proto_msgTypes,
	}.Build()
	File_pkg_graph_data_proto = out.File
	file_pkg_graph_data_proto_rawDesc = nil
	file_pkg_graph_data_proto_goTypes = nil
	file_pkg_graph_data_proto_depIdxs = nil
}