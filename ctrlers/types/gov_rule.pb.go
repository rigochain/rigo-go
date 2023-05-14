// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: gov_rule.proto

package types

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

type GovRuleProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version                int64  `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	MaxValidatorCnt        int64  `protobuf:"varint,2,opt,name=max_validator_cnt,json=maxValidatorCnt,proto3" json:"max_validator_cnt,omitempty"`
	XAmountPerPower        []byte `protobuf:"bytes,3,opt,name=_amount_per_power,json=AmountPerPower,proto3" json:"_amount_per_power,omitempty"`
	XRewardPerPower        []byte `protobuf:"bytes,4,opt,name=_reward_per_power,json=RewardPerPower,proto3" json:"_reward_per_power,omitempty"`
	LazyRewardBlocks       int64  `protobuf:"varint,5,opt,name=lazy_reward_blocks,json=lazyRewardBlocks,proto3" json:"lazy_reward_blocks,omitempty"`
	LazyApplyingBlocks     int64  `protobuf:"varint,6,opt,name=lazy_applying_blocks,json=lazyApplyingBlocks,proto3" json:"lazy_applying_blocks,omitempty"`
	XMinTrxFee             []byte `protobuf:"bytes,7,opt,name=_min_trx_fee,json=MinTrxFee,proto3" json:"_min_trx_fee,omitempty"`
	MinVotingPeriodBlocks  int64  `protobuf:"varint,8,opt,name=min_voting_period_blocks,json=minVotingPeriodBlocks,proto3" json:"min_voting_period_blocks,omitempty"`
	MaxVotingPeriodBlocks  int64  `protobuf:"varint,9,opt,name=max_voting_period_blocks,json=maxVotingPeriodBlocks,proto3" json:"max_voting_period_blocks,omitempty"`
	MinSelfStakeRatio      int64  `protobuf:"varint,10,opt,name=min_self_stake_ratio,json=minSelfStakeRatio,proto3" json:"min_self_stake_ratio,omitempty"`
	MaxUpdatableStakeRatio int64  `protobuf:"varint,11,opt,name=max_updatable_stake_ratio,json=maxUpdatableStakeRatio,proto3" json:"max_updatable_stake_ratio,omitempty"`
}

func (x *GovRuleProto) Reset() {
	*x = GovRuleProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gov_rule_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GovRuleProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GovRuleProto) ProtoMessage() {}

func (x *GovRuleProto) ProtoReflect() protoreflect.Message {
	mi := &file_gov_rule_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GovRuleProto.ProtoReflect.Descriptor instead.
func (*GovRuleProto) Descriptor() ([]byte, []int) {
	return file_gov_rule_proto_rawDescGZIP(), []int{0}
}

func (x *GovRuleProto) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *GovRuleProto) GetMaxValidatorCnt() int64 {
	if x != nil {
		return x.MaxValidatorCnt
	}
	return 0
}

func (x *GovRuleProto) GetXAmountPerPower() []byte {
	if x != nil {
		return x.XAmountPerPower
	}
	return nil
}

func (x *GovRuleProto) GetXRewardPerPower() []byte {
	if x != nil {
		return x.XRewardPerPower
	}
	return nil
}

func (x *GovRuleProto) GetLazyRewardBlocks() int64 {
	if x != nil {
		return x.LazyRewardBlocks
	}
	return 0
}

func (x *GovRuleProto) GetLazyApplyingBlocks() int64 {
	if x != nil {
		return x.LazyApplyingBlocks
	}
	return 0
}

func (x *GovRuleProto) GetXMinTrxFee() []byte {
	if x != nil {
		return x.XMinTrxFee
	}
	return nil
}

func (x *GovRuleProto) GetMinVotingPeriodBlocks() int64 {
	if x != nil {
		return x.MinVotingPeriodBlocks
	}
	return 0
}

func (x *GovRuleProto) GetMaxVotingPeriodBlocks() int64 {
	if x != nil {
		return x.MaxVotingPeriodBlocks
	}
	return 0
}

func (x *GovRuleProto) GetMinSelfStakeRatio() int64 {
	if x != nil {
		return x.MinSelfStakeRatio
	}
	return 0
}

func (x *GovRuleProto) GetMaxUpdatableStakeRatio() int64 {
	if x != nil {
		return x.MaxUpdatableStakeRatio
	}
	return 0
}

var File_gov_rule_proto protoreflect.FileDescriptor

var file_gov_rule_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x67, 0x6f, 0x76, 0x5f, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x22, 0x89, 0x04, 0x0a, 0x0c, 0x47, 0x6f, 0x76, 0x52,
	0x75, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x11, 0x6d, 0x61, 0x78, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x5f, 0x63, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x6d,
	0x61, 0x78, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x43, 0x6e, 0x74, 0x12, 0x29,
	0x0a, 0x11, 0x5f, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x70, 0x6f,
	0x77, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x41, 0x6d, 0x6f, 0x75, 0x6e,
	0x74, 0x50, 0x65, 0x72, 0x50, 0x6f, 0x77, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x11, 0x5f, 0x72, 0x65,
	0x77, 0x61, 0x72, 0x64, 0x5f, 0x70, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x77, 0x65, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x50, 0x65, 0x72, 0x50,
	0x6f, 0x77, 0x65, 0x72, 0x12, 0x2c, 0x0a, 0x12, 0x6c, 0x61, 0x7a, 0x79, 0x5f, 0x72, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x10, 0x6c, 0x61, 0x7a, 0x79, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x73, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x61, 0x7a, 0x79, 0x5f, 0x61, 0x70, 0x70, 0x6c, 0x79,
	0x69, 0x6e, 0x67, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x12, 0x6c, 0x61, 0x7a, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x69, 0x6e, 0x67, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x12, 0x1f, 0x0a, 0x0c, 0x5f, 0x6d, 0x69, 0x6e, 0x5f, 0x74, 0x72, 0x78,
	0x5f, 0x66, 0x65, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x4d, 0x69, 0x6e, 0x54,
	0x72, 0x78, 0x46, 0x65, 0x65, 0x12, 0x37, 0x0a, 0x18, 0x6d, 0x69, 0x6e, 0x5f, 0x76, 0x6f, 0x74,
	0x69, 0x6e, 0x67, 0x5f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x03, 0x52, 0x15, 0x6d, 0x69, 0x6e, 0x56, 0x6f, 0x74, 0x69,
	0x6e, 0x67, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x37,
	0x0a, 0x18, 0x6d, 0x61, 0x78, 0x5f, 0x76, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x65, 0x72,
	0x69, 0x6f, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x15, 0x6d, 0x61, 0x78, 0x56, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x65, 0x72, 0x69, 0x6f,
	0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x2f, 0x0a, 0x14, 0x6d, 0x69, 0x6e, 0x5f, 0x73,
	0x65, 0x6c, 0x66, 0x5f, 0x73, 0x74, 0x61, 0x6b, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x6d, 0x69, 0x6e, 0x53, 0x65, 0x6c, 0x66, 0x53, 0x74,
	0x61, 0x6b, 0x65, 0x52, 0x61, 0x74, 0x69, 0x6f, 0x12, 0x39, 0x0a, 0x19, 0x6d, 0x61, 0x78, 0x5f,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x6b, 0x65, 0x5f,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x16, 0x6d, 0x61, 0x78,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x6b, 0x65, 0x52, 0x61,
	0x74, 0x69, 0x6f, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x72, 0x69, 0x67, 0x6f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x72, 0x69, 0x67, 0x6f,
	0x2d, 0x67, 0x6f, 0x2f, 0x63, 0x74, 0x72, 0x6c, 0x65, 0x72, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gov_rule_proto_rawDescOnce sync.Once
	file_gov_rule_proto_rawDescData = file_gov_rule_proto_rawDesc
)

func file_gov_rule_proto_rawDescGZIP() []byte {
	file_gov_rule_proto_rawDescOnce.Do(func() {
		file_gov_rule_proto_rawDescData = protoimpl.X.CompressGZIP(file_gov_rule_proto_rawDescData)
	})
	return file_gov_rule_proto_rawDescData
}

var file_gov_rule_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_gov_rule_proto_goTypes = []interface{}{
	(*GovRuleProto)(nil), // 0: types.GovRuleProto
}
var file_gov_rule_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_gov_rule_proto_init() }
func file_gov_rule_proto_init() {
	if File_gov_rule_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gov_rule_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GovRuleProto); i {
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
			RawDescriptor: file_gov_rule_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_gov_rule_proto_goTypes,
		DependencyIndexes: file_gov_rule_proto_depIdxs,
		MessageInfos:      file_gov_rule_proto_msgTypes,
	}.Build()
	File_gov_rule_proto = out.File
	file_gov_rule_proto_rawDesc = nil
	file_gov_rule_proto_goTypes = nil
	file_gov_rule_proto_depIdxs = nil
}
