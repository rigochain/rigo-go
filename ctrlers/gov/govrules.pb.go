// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: govrules.proto

package gov

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

type GovRulesProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version            int64  `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	MaxValidatorCnt    int64  `protobuf:"varint,2,opt,name=maxValidatorCnt,proto3" json:"maxValidatorCnt,omitempty"`
	XAmountPerPower    []byte `protobuf:"bytes,3,opt,name=_amount_per_power,json=AmountPerPower,proto3" json:"_amount_per_power,omitempty"`
	XRewardPerPower    []byte `protobuf:"bytes,4,opt,name=_reward_per_power,json=RewardPerPower,proto3" json:"_reward_per_power,omitempty"`
	LazyRewardBlocks   int64  `protobuf:"varint,5,opt,name=lazyRewardBlocks,proto3" json:"lazyRewardBlocks,omitempty"`
	LazyApplyingBlocks int64  `protobuf:"varint,6,opt,name=lazyApplyingBlocks,proto3" json:"lazyApplyingBlocks,omitempty"`
}

func (x *GovRulesProto) Reset() {
	*x = GovRulesProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_govrules_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GovRulesProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GovRulesProto) ProtoMessage() {}

func (x *GovRulesProto) ProtoReflect() protoreflect.Message {
	mi := &file_govrules_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GovRulesProto.ProtoReflect.Descriptor instead.
func (*GovRulesProto) Descriptor() ([]byte, []int) {
	return file_govrules_proto_rawDescGZIP(), []int{0}
}

func (x *GovRulesProto) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *GovRulesProto) GetMaxValidatorCnt() int64 {
	if x != nil {
		return x.MaxValidatorCnt
	}
	return 0
}

func (x *GovRulesProto) GetXAmountPerPower() []byte {
	if x != nil {
		return x.XAmountPerPower
	}
	return nil
}

func (x *GovRulesProto) GetXRewardPerPower() []byte {
	if x != nil {
		return x.XRewardPerPower
	}
	return nil
}

func (x *GovRulesProto) GetLazyRewardBlocks() int64 {
	if x != nil {
		return x.LazyRewardBlocks
	}
	return 0
}

func (x *GovRulesProto) GetLazyApplyingBlocks() int64 {
	if x != nil {
		return x.LazyApplyingBlocks
	}
	return 0
}

var File_govrules_proto protoreflect.FileDescriptor

var file_govrules_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x67, 0x6f, 0x76, 0x72, 0x75, 0x6c, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x22, 0x85, 0x02, 0x0a, 0x0d, 0x47, 0x6f, 0x76, 0x52,
	0x75, 0x6c, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x28, 0x0a, 0x0f, 0x6d, 0x61, 0x78, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x43, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x6d, 0x61,
	0x78, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x43, 0x6e, 0x74, 0x12, 0x29, 0x0a,
	0x11, 0x5f, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x77,
	0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x50, 0x65, 0x72, 0x50, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x11, 0x5f, 0x72, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x77, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0e, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x50, 0x65, 0x72, 0x50, 0x6f,
	0x77, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x10, 0x6c, 0x61, 0x7a, 0x79, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x10, 0x6c,
	0x61, 0x7a, 0x79, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12,
	0x2e, 0x0a, 0x12, 0x6c, 0x61, 0x7a, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6c, 0x61, 0x7a,
	0x79, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x42,
	0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x79,
	0x73, 0x65, 0x65, 0x2f, 0x61, 0x72, 0x63, 0x61, 0x6e, 0x75, 0x73, 0x2f, 0x63, 0x74, 0x72, 0x6c,
	0x65, 0x72, 0x73, 0x2f, 0x67, 0x6f, 0x76, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_govrules_proto_rawDescOnce sync.Once
	file_govrules_proto_rawDescData = file_govrules_proto_rawDesc
)

func file_govrules_proto_rawDescGZIP() []byte {
	file_govrules_proto_rawDescOnce.Do(func() {
		file_govrules_proto_rawDescData = protoimpl.X.CompressGZIP(file_govrules_proto_rawDescData)
	})
	return file_govrules_proto_rawDescData
}

var file_govrules_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_govrules_proto_goTypes = []interface{}{
	(*GovRulesProto)(nil), // 0: types.GovRulesProto
}
var file_govrules_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_govrules_proto_init() }
func file_govrules_proto_init() {
	if File_govrules_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_govrules_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GovRulesProto); i {
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
			RawDescriptor: file_govrules_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_govrules_proto_goTypes,
		DependencyIndexes: file_govrules_proto_depIdxs,
		MessageInfos:      file_govrules_proto_msgTypes,
	}.Build()
	File_govrules_proto = out.File
	file_govrules_proto_rawDesc = nil
	file_govrules_proto_goTypes = nil
	file_govrules_proto_depIdxs = nil
}
