// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.15.8
// source: proto/prysm/v1alpha1/prevoting.proto

package eth

import (
	reflect "reflect"
	sync "sync"

	github_com_prysmaticlabs_eth2_types "github.com/prysmaticlabs/eth2-types"
	github_com_prysmaticlabs_go_bitfield "github.com/prysmaticlabs/go-bitfield"
	_ "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/ext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PreVoteData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Slot       github_com_prysmaticlabs_eth2_types.Slot           `protobuf:"varint,1,opt,name=slot,proto3" json:"slot,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Slot"`
	Index      github_com_prysmaticlabs_eth2_types.CommitteeIndex `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.CommitteeIndex"`
	Candidates []byte                                             `protobuf:"bytes,3,opt,name=candidates,proto3" json:"candidates,omitempty" ssz-max:"4096"`
}

func (x *PreVoteData) Reset() {
	*x = PreVoteData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreVoteData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreVoteData) ProtoMessage() {}

func (x *PreVoteData) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreVoteData.ProtoReflect.Descriptor instead.
func (*PreVoteData) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP(), []int{0}
}

func (x *PreVoteData) GetSlot() github_com_prysmaticlabs_eth2_types.Slot {
	if x != nil {
		return x.Slot
	}
	return github_com_prysmaticlabs_eth2_types.Slot(0)
}

func (x *PreVoteData) GetIndex() github_com_prysmaticlabs_eth2_types.CommitteeIndex {
	if x != nil {
		return x.Index
	}
	return github_com_prysmaticlabs_eth2_types.CommitteeIndex(0)
}

func (x *PreVoteData) GetCandidates() []byte {
	if x != nil {
		return x.Candidates
	}
	return nil
}

type PreVoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Slot           github_com_prysmaticlabs_eth2_types.Slot           `protobuf:"varint,1,opt,name=slot,proto3" json:"slot,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Slot"`
	CommitteeIndex github_com_prysmaticlabs_eth2_types.CommitteeIndex `protobuf:"varint,2,opt,name=committee_index,json=committeeIndex,proto3" json:"committee_index,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.CommitteeIndex"`
}

func (x *PreVoteRequest) Reset() {
	*x = PreVoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreVoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreVoteRequest) ProtoMessage() {}

func (x *PreVoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreVoteRequest.ProtoReflect.Descriptor instead.
func (*PreVoteRequest) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP(), []int{1}
}

func (x *PreVoteRequest) GetSlot() github_com_prysmaticlabs_eth2_types.Slot {
	if x != nil {
		return x.Slot
	}
	return github_com_prysmaticlabs_eth2_types.Slot(0)
}

func (x *PreVoteRequest) GetCommitteeIndex() github_com_prysmaticlabs_eth2_types.CommitteeIndex {
	if x != nil {
		return x.CommitteeIndex
	}
	return github_com_prysmaticlabs_eth2_types.CommitteeIndex(0)
}

type IndexedPreVote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AttestingIndices []uint64     `protobuf:"varint,1,rep,packed,name=attesting_indices,json=attestingIndices,proto3" json:"attesting_indices,omitempty" ssz-max:"2048"`
	Data             *PreVoteData `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Signature        []byte       `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *IndexedPreVote) Reset() {
	*x = IndexedPreVote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IndexedPreVote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IndexedPreVote) ProtoMessage() {}

func (x *IndexedPreVote) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IndexedPreVote.ProtoReflect.Descriptor instead.
func (*IndexedPreVote) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP(), []int{2}
}

func (x *IndexedPreVote) GetAttestingIndices() []uint64 {
	if x != nil {
		return x.AttestingIndices
	}
	return nil
}

func (x *IndexedPreVote) GetData() *PreVoteData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *IndexedPreVote) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type PreVote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AggregationBits github_com_prysmaticlabs_go_bitfield.Bitlist `protobuf:"bytes,1,opt,name=aggregation_bits,json=aggregationBits,proto3" json:"aggregation_bits,omitempty" cast-type:"github.com/prysmaticlabs/go-bitfield.Bitlist" ssz-max:"2048"`
	Data            *PreVoteData                                 `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Signature       []byte                                       `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *PreVote) Reset() {
	*x = PreVote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreVote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreVote) ProtoMessage() {}

func (x *PreVote) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreVote.ProtoReflect.Descriptor instead.
func (*PreVote) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP(), []int{3}
}

func (x *PreVote) GetAggregationBits() github_com_prysmaticlabs_go_bitfield.Bitlist {
	if x != nil {
		return x.AggregationBits
	}
	return github_com_prysmaticlabs_go_bitfield.Bitlist(nil)
}

func (x *PreVote) GetData() *PreVoteData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *PreVote) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type PreVotePacket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PreVotes []*PreVote `protobuf:"bytes,1,rep,name=pre_votes,json=preVotes,proto3" json:"pre_votes,omitempty"`
}

func (x *PreVotePacket) Reset() {
	*x = PreVotePacket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreVotePacket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreVotePacket) ProtoMessage() {}

func (x *PreVotePacket) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreVotePacket.ProtoReflect.Descriptor instead.
func (*PreVotePacket) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP(), []int{4}
}

func (x *PreVotePacket) GetPreVotes() []*PreVote {
	if x != nil {
		return x.PreVotes
	}
	return nil
}

var File_proto_prysm_v1alpha1_prevoting_proto protoreflect.FileDescriptor

var file_proto_prysm_v1alpha1_prevoting_proto_rawDesc = []byte{
	0x0a, 0x24, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x70, 0x72, 0x65, 0x76, 0x6f, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1b, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc7, 0x01, 0x0a, 0x0b, 0x50,
	0x72, 0x65, 0x56, 0x6f, 0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x40, 0x0a, 0x04, 0x73, 0x6c,
	0x6f, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2c, 0x82, 0xb5, 0x18, 0x28, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74,
	0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x53, 0x6c, 0x6f, 0x74, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x12, 0x4c, 0x0a, 0x05,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18,
	0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73,
	0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x28, 0x0a, 0x0a, 0x63, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x08,
	0x92, 0xb5, 0x18, 0x04, 0x34, 0x30, 0x39, 0x36, 0x52, 0x0a, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x73, 0x22, 0xb3, 0x01, 0x0a, 0x0e, 0x50, 0x72, 0x65, 0x56, 0x6f, 0x74, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x40, 0x0a, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2c, 0x82, 0xb5, 0x18, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c,
	0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53,
	0x6c, 0x6f, 0x74, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x12, 0x5f, 0x0a, 0x0f, 0x63, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73,
	0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0e, 0x63, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0xa5, 0x01, 0x0a, 0x0e, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x50, 0x72, 0x65, 0x56, 0x6f, 0x74, 0x65, 0x12, 0x35, 0x0a,
	0x11, 0x61, 0x74, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x69, 0x63,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x42, 0x08, 0x92, 0xb5, 0x18, 0x04, 0x32, 0x30,
	0x34, 0x38, 0x52, 0x10, 0x61, 0x74, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x64,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x36, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74,
	0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x65, 0x56, 0x6f,
	0x74, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x24, 0x0a, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x42,
	0x06, 0x8a, 0xb5, 0x18, 0x02, 0x39, 0x36, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x22, 0xcc, 0x01, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x56, 0x6f, 0x74, 0x65, 0x12, 0x63,
	0x0a, 0x10, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x62, 0x69,
	0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x38, 0x82, 0xb5, 0x18, 0x2c, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74,
	0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x67, 0x6f, 0x2d, 0x62, 0x69, 0x74, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x2e, 0x42, 0x69, 0x74, 0x6c, 0x69, 0x73, 0x74, 0x92, 0xb5, 0x18, 0x04, 0x32, 0x30,
	0x34, 0x38, 0x52, 0x0f, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42,
	0x69, 0x74, 0x73, 0x12, 0x36, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x65, 0x56, 0x6f, 0x74,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x24, 0x0a, 0x09, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x06,
	0x8a, 0xb5, 0x18, 0x02, 0x39, 0x36, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x22, 0x4c, 0x0a, 0x0d, 0x50, 0x72, 0x65, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x61, 0x63, 0x6b,
	0x65, 0x74, 0x12, 0x3b, 0x0a, 0x09, 0x70, 0x72, 0x65, 0x5f, 0x76, 0x6f, 0x74, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x50, 0x72,
	0x65, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x08, 0x70, 0x72, 0x65, 0x56, 0x6f, 0x74, 0x65, 0x73, 0x42,
	0xaf, 0x01, 0x0a, 0x19, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x0e, 0x50,
	0x72, 0x65, 0x76, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x50, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x77, 0x61, 0x74, 0x65, 0x72, 0x66, 0x61, 0x6c,
	0x6c, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x77, 0x61, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x6c, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x63, 0x6f, 0x6f,
	0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70,
	0x72, 0x79, 0x73, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x65, 0x74,
	0x68, 0xaa, 0x02, 0x15, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x45, 0x74, 0x68,
	0x2e, 0x56, 0x31, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02, 0x15, 0x45, 0x74, 0x68, 0x65,
	0x72, 0x65, 0x75, 0x6d, 0x5c, 0x45, 0x74, 0x68, 0x5c, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_prysm_v1alpha1_prevoting_proto_rawDescOnce sync.Once
	file_proto_prysm_v1alpha1_prevoting_proto_rawDescData = file_proto_prysm_v1alpha1_prevoting_proto_rawDesc
)

func file_proto_prysm_v1alpha1_prevoting_proto_rawDescGZIP() []byte {
	file_proto_prysm_v1alpha1_prevoting_proto_rawDescOnce.Do(func() {
		file_proto_prysm_v1alpha1_prevoting_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_prysm_v1alpha1_prevoting_proto_rawDescData)
	})
	return file_proto_prysm_v1alpha1_prevoting_proto_rawDescData
}

var file_proto_prysm_v1alpha1_prevoting_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_proto_prysm_v1alpha1_prevoting_proto_goTypes = []interface{}{
	(*PreVoteData)(nil),    // 0: ethereum.eth.v1alpha1.PreVoteData
	(*PreVoteRequest)(nil), // 1: ethereum.eth.v1alpha1.PreVoteRequest
	(*IndexedPreVote)(nil), // 2: ethereum.eth.v1alpha1.IndexedPreVote
	(*PreVote)(nil),        // 3: ethereum.eth.v1alpha1.PreVote
	(*PreVotePacket)(nil),  // 4: ethereum.eth.v1alpha1.PreVotePacket
}
var file_proto_prysm_v1alpha1_prevoting_proto_depIdxs = []int32{
	0, // 0: ethereum.eth.v1alpha1.IndexedPreVote.data:type_name -> ethereum.eth.v1alpha1.PreVoteData
	0, // 1: ethereum.eth.v1alpha1.PreVote.data:type_name -> ethereum.eth.v1alpha1.PreVoteData
	3, // 2: ethereum.eth.v1alpha1.PreVotePacket.pre_votes:type_name -> ethereum.eth.v1alpha1.PreVote
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_prysm_v1alpha1_prevoting_proto_init() }
func file_proto_prysm_v1alpha1_prevoting_proto_init() {
	if File_proto_prysm_v1alpha1_prevoting_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreVoteData); i {
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
		file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreVoteRequest); i {
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
		file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IndexedPreVote); i {
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
		file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreVote); i {
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
		file_proto_prysm_v1alpha1_prevoting_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreVotePacket); i {
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
			RawDescriptor: file_proto_prysm_v1alpha1_prevoting_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_prysm_v1alpha1_prevoting_proto_goTypes,
		DependencyIndexes: file_proto_prysm_v1alpha1_prevoting_proto_depIdxs,
		MessageInfos:      file_proto_prysm_v1alpha1_prevoting_proto_msgTypes,
	}.Build()
	File_proto_prysm_v1alpha1_prevoting_proto = out.File
	file_proto_prysm_v1alpha1_prevoting_proto_rawDesc = nil
	file_proto_prysm_v1alpha1_prevoting_proto_goTypes = nil
	file_proto_prysm_v1alpha1_prevoting_proto_depIdxs = nil
}