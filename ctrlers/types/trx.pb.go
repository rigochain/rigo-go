// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: trx.proto

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

type TrxProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version   uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Time      int64  `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`
	Nonce     uint64 `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	From      []byte `protobuf:"bytes,4,opt,name=from,proto3" json:"from,omitempty"`
	To        []byte `protobuf:"bytes,5,opt,name=to,proto3" json:"to,omitempty"`
	XAmount   []byte `protobuf:"bytes,6,opt,name=_amount,json=Amount,proto3" json:"_amount,omitempty"`
	Gas       uint64 `protobuf:"varint,7,opt,name=gas,proto3" json:"gas,omitempty"`
	XGasPrice []byte `protobuf:"bytes,8,opt,name=_gasPrice,json=GasPrice,proto3" json:"_gasPrice,omitempty"`
	Type      int32  `protobuf:"varint,9,opt,name=type,proto3" json:"type,omitempty"`
	XPayload  []byte `protobuf:"bytes,10,opt,name=_payload,json=Payload,proto3" json:"_payload,omitempty"`
	Sig       []byte `protobuf:"bytes,11,opt,name=sig,proto3" json:"sig,omitempty"`
}

func (x *TrxProto) Reset() {
	*x = TrxProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxProto) ProtoMessage() {}

func (x *TrxProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxProto.ProtoReflect.Descriptor instead.
func (*TrxProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{0}
}

func (x *TrxProto) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *TrxProto) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *TrxProto) GetNonce() uint64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *TrxProto) GetFrom() []byte {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *TrxProto) GetTo() []byte {
	if x != nil {
		return x.To
	}
	return nil
}

func (x *TrxProto) GetXAmount() []byte {
	if x != nil {
		return x.XAmount
	}
	return nil
}

func (x *TrxProto) GetGas() uint64 {
	if x != nil {
		return x.Gas
	}
	return 0
}

func (x *TrxProto) GetXGasPrice() []byte {
	if x != nil {
		return x.XGasPrice
	}
	return nil
}

func (x *TrxProto) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *TrxProto) GetXPayload() []byte {
	if x != nil {
		return x.XPayload
	}
	return nil
}

func (x *TrxProto) GetSig() []byte {
	if x != nil {
		return x.Sig
	}
	return nil
}

type TrxPayloadAssetTransferProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrxPayloadAssetTransferProto) Reset() {
	*x = TrxPayloadAssetTransferProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadAssetTransferProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadAssetTransferProto) ProtoMessage() {}

func (x *TrxPayloadAssetTransferProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadAssetTransferProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadAssetTransferProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{1}
}

type TrxPayloadStakingProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TrxPayloadStakingProto) Reset() {
	*x = TrxPayloadStakingProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadStakingProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadStakingProto) ProtoMessage() {}

func (x *TrxPayloadStakingProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadStakingProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadStakingProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{2}
}

type TrxPayloadUnstakingProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash []byte `protobuf:"bytes,1,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
}

func (x *TrxPayloadUnstakingProto) Reset() {
	*x = TrxPayloadUnstakingProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadUnstakingProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadUnstakingProto) ProtoMessage() {}

func (x *TrxPayloadUnstakingProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadUnstakingProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadUnstakingProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{3}
}

func (x *TrxPayloadUnstakingProto) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

type TrxPayloadWithdrawProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	XReqAmt []byte `protobuf:"bytes,1,opt,name=_reqAmt,json=ReqAmt,proto3" json:"_reqAmt,omitempty"`
}

func (x *TrxPayloadWithdrawProto) Reset() {
	*x = TrxPayloadWithdrawProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadWithdrawProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadWithdrawProto) ProtoMessage() {}

func (x *TrxPayloadWithdrawProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadWithdrawProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadWithdrawProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{4}
}

func (x *TrxPayloadWithdrawProto) GetXReqAmt() []byte {
	if x != nil {
		return x.XReqAmt
	}
	return nil
}

type TrxPayloadContractProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	XData []byte `protobuf:"bytes,1,opt,name=_data,json=Data,proto3" json:"_data,omitempty"`
}

func (x *TrxPayloadContractProto) Reset() {
	*x = TrxPayloadContractProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadContractProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadContractProto) ProtoMessage() {}

func (x *TrxPayloadContractProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadContractProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadContractProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{5}
}

func (x *TrxPayloadContractProto) GetXData() []byte {
	if x != nil {
		return x.XData
	}
	return nil
}

type TrxPayloadProposalProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message           string   `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	StartVotingHeight int64    `protobuf:"varint,2,opt,name=start_voting_height,json=startVotingHeight,proto3" json:"start_voting_height,omitempty"`
	VotingBlocks      int64    `protobuf:"varint,3,opt,name=voting_blocks,json=votingBlocks,proto3" json:"voting_blocks,omitempty"`
	ApplyHeight       int64    `protobuf:"varint,4,opt,name=apply_height,json=applyHeight,proto3" json:"apply_height,omitempty"`
	OptType           int32    `protobuf:"varint,5,opt,name=opt_type,json=optType,proto3" json:"opt_type,omitempty"`
	Options           [][]byte `protobuf:"bytes,6,rep,name=options,proto3" json:"options,omitempty"`
}

func (x *TrxPayloadProposalProto) Reset() {
	*x = TrxPayloadProposalProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadProposalProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadProposalProto) ProtoMessage() {}

func (x *TrxPayloadProposalProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadProposalProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadProposalProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{6}
}

func (x *TrxPayloadProposalProto) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *TrxPayloadProposalProto) GetStartVotingHeight() int64 {
	if x != nil {
		return x.StartVotingHeight
	}
	return 0
}

func (x *TrxPayloadProposalProto) GetVotingBlocks() int64 {
	if x != nil {
		return x.VotingBlocks
	}
	return 0
}

func (x *TrxPayloadProposalProto) GetApplyHeight() int64 {
	if x != nil {
		return x.ApplyHeight
	}
	return 0
}

func (x *TrxPayloadProposalProto) GetOptType() int32 {
	if x != nil {
		return x.OptType
	}
	return 0
}

func (x *TrxPayloadProposalProto) GetOptions() [][]byte {
	if x != nil {
		return x.Options
	}
	return nil
}

type TrxPayloadVotingProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash []byte `protobuf:"bytes,1,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	Choice int32  `protobuf:"varint,2,opt,name=choice,proto3" json:"choice,omitempty"`
}

func (x *TrxPayloadVotingProto) Reset() {
	*x = TrxPayloadVotingProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadVotingProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadVotingProto) ProtoMessage() {}

func (x *TrxPayloadVotingProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadVotingProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadVotingProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{7}
}

func (x *TrxPayloadVotingProto) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

func (x *TrxPayloadVotingProto) GetChoice() int32 {
	if x != nil {
		return x.Choice
	}
	return 0
}

type TrxPayloadSetDocProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Url  string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *TrxPayloadSetDocProto) Reset() {
	*x = TrxPayloadSetDocProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_trx_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TrxPayloadSetDocProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TrxPayloadSetDocProto) ProtoMessage() {}

func (x *TrxPayloadSetDocProto) ProtoReflect() protoreflect.Message {
	mi := &file_trx_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TrxPayloadSetDocProto.ProtoReflect.Descriptor instead.
func (*TrxPayloadSetDocProto) Descriptor() ([]byte, []int) {
	return file_trx_proto_rawDescGZIP(), []int{8}
}

func (x *TrxPayloadSetDocProto) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TrxPayloadSetDocProto) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_trx_proto protoreflect.FileDescriptor

var file_trx_proto_rawDesc = []byte{
	0x0a, 0x09, 0x74, 0x72, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x22, 0xfb, 0x01, 0x0a, 0x08, 0x54, 0x72, 0x78, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6e, 0x6f,
	0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x17, 0x0a, 0x07, 0x5f, 0x61, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x61, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x67,
	0x61, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x5f, 0x67, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x47, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x5f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x73, 0x69, 0x67, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x73, 0x69, 0x67,
	0x22, 0x1e, 0x0a, 0x1c, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x41, 0x73,
	0x73, 0x65, 0x74, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x18, 0x0a, 0x16, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x74,
	0x61, 0x6b, 0x69, 0x6e, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x33, 0x0a, 0x18, 0x54, 0x72,
	0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x55, 0x6e, 0x73, 0x74, 0x61, 0x6b, 0x69, 0x6e,
	0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x22,
	0x32, 0x0a, 0x17, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x57, 0x69, 0x74,
	0x68, 0x64, 0x72, 0x61, 0x77, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x0a, 0x07, 0x5f, 0x72,
	0x65, 0x71, 0x41, 0x6d, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x52, 0x65, 0x71,
	0x41, 0x6d, 0x74, 0x22, 0x2e, 0x0a, 0x17, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13,
	0x0a, 0x05, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x44,
	0x61, 0x74, 0x61, 0x22, 0xe0, 0x01, 0x0a, 0x17, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x13, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x5f, 0x76, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x73, 0x74, 0x61, 0x72, 0x74, 0x56, 0x6f, 0x74,
	0x69, 0x6e, 0x67, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x76, 0x6f, 0x74,
	0x69, 0x6e, 0x67, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0c, 0x76, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x21,
	0x0a, 0x0c, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x48, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x70, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x6f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x48, 0x0a, 0x15, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x56, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x68, 0x6f, 0x69,
	0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65,
	0x22, 0x3d, 0x0a, 0x15, 0x54, 0x72, 0x78, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x65,
	0x74, 0x44, 0x6f, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x42,
	0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x69,
	0x67, 0x6f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x72, 0x69, 0x67, 0x6f, 0x2d, 0x67, 0x6f, 0x2f,
	0x63, 0x74, 0x72, 0x6c, 0x65, 0x72, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_trx_proto_rawDescOnce sync.Once
	file_trx_proto_rawDescData = file_trx_proto_rawDesc
)

func file_trx_proto_rawDescGZIP() []byte {
	file_trx_proto_rawDescOnce.Do(func() {
		file_trx_proto_rawDescData = protoimpl.X.CompressGZIP(file_trx_proto_rawDescData)
	})
	return file_trx_proto_rawDescData
}

var file_trx_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_trx_proto_goTypes = []interface{}{
	(*TrxProto)(nil),                     // 0: types.TrxProto
	(*TrxPayloadAssetTransferProto)(nil), // 1: types.TrxPayloadAssetTransferProto
	(*TrxPayloadStakingProto)(nil),       // 2: types.TrxPayloadStakingProto
	(*TrxPayloadUnstakingProto)(nil),     // 3: types.TrxPayloadUnstakingProto
	(*TrxPayloadWithdrawProto)(nil),      // 4: types.TrxPayloadWithdrawProto
	(*TrxPayloadContractProto)(nil),      // 5: types.TrxPayloadContractProto
	(*TrxPayloadProposalProto)(nil),      // 6: types.TrxPayloadProposalProto
	(*TrxPayloadVotingProto)(nil),        // 7: types.TrxPayloadVotingProto
	(*TrxPayloadSetDocProto)(nil),        // 8: types.TrxPayloadSetDocProto
}
var file_trx_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_trx_proto_init() }
func file_trx_proto_init() {
	if File_trx_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_trx_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxProto); i {
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
		file_trx_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadAssetTransferProto); i {
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
		file_trx_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadStakingProto); i {
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
		file_trx_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadUnstakingProto); i {
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
		file_trx_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadWithdrawProto); i {
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
		file_trx_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadContractProto); i {
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
		file_trx_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadProposalProto); i {
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
		file_trx_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadVotingProto); i {
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
		file_trx_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TrxPayloadSetDocProto); i {
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
			RawDescriptor: file_trx_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_trx_proto_goTypes,
		DependencyIndexes: file_trx_proto_depIdxs,
		MessageInfos:      file_trx_proto_msgTypes,
	}.Build()
	File_trx_proto = out.File
	file_trx_proto_rawDesc = nil
	file_trx_proto_goTypes = nil
	file_trx_proto_depIdxs = nil
}
