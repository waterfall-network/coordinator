// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.15.8
// source: proto/eth/v2/sync_committee.proto

package eth

import (
	reflect "reflect"
	sync "sync"

	github_com_prysmaticlabs_eth2_types "github.com/prysmaticlabs/eth2-types"
	_ "github.com/waterfall-foundation/coordinator/proto/eth/ext"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SubmitSyncCommitteeSignaturesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []*SyncCommitteeMessage `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *SubmitSyncCommitteeSignaturesRequest) Reset() {
	*x = SubmitSyncCommitteeSignaturesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubmitSyncCommitteeSignaturesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitSyncCommitteeSignaturesRequest) ProtoMessage() {}

func (x *SubmitSyncCommitteeSignaturesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitSyncCommitteeSignaturesRequest.ProtoReflect.Descriptor instead.
func (*SubmitSyncCommitteeSignaturesRequest) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{0}
}

func (x *SubmitSyncCommitteeSignaturesRequest) GetData() []*SyncCommitteeMessage {
	if x != nil {
		return x.Data
	}
	return nil
}

type SyncCommittee struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pubkeys         [][]byte `protobuf:"bytes,1,rep,name=pubkeys,proto3" json:"pubkeys,omitempty" ssz-size:"512,48"`
	AggregatePubkey []byte   `protobuf:"bytes,2,opt,name=aggregate_pubkey,json=aggregatePubkey,proto3" json:"aggregate_pubkey,omitempty" ssz-size:"48"`
}

func (x *SyncCommittee) Reset() {
	*x = SyncCommittee{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncCommittee) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncCommittee) ProtoMessage() {}

func (x *SyncCommittee) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncCommittee.ProtoReflect.Descriptor instead.
func (*SyncCommittee) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{1}
}

func (x *SyncCommittee) GetPubkeys() [][]byte {
	if x != nil {
		return x.Pubkeys
	}
	return nil
}

func (x *SyncCommittee) GetAggregatePubkey() []byte {
	if x != nil {
		return x.AggregatePubkey
	}
	return nil
}

type SubmitPoolSyncCommitteeSignatures struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []*SyncCommitteeMessage `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *SubmitPoolSyncCommitteeSignatures) Reset() {
	*x = SubmitPoolSyncCommitteeSignatures{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubmitPoolSyncCommitteeSignatures) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitPoolSyncCommitteeSignatures) ProtoMessage() {}

func (x *SubmitPoolSyncCommitteeSignatures) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitPoolSyncCommitteeSignatures.ProtoReflect.Descriptor instead.
func (*SubmitPoolSyncCommitteeSignatures) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{2}
}

func (x *SubmitPoolSyncCommitteeSignatures) GetData() []*SyncCommitteeMessage {
	if x != nil {
		return x.Data
	}
	return nil
}

type SyncCommitteeMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Slot            github_com_prysmaticlabs_eth2_types.Slot           `protobuf:"varint,1,opt,name=slot,proto3" json:"slot,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Slot"`
	BeaconBlockRoot []byte                                             `protobuf:"bytes,2,opt,name=beacon_block_root,json=beaconBlockRoot,proto3" json:"beacon_block_root,omitempty" ssz-size:"32"`
	ValidatorIndex  github_com_prysmaticlabs_eth2_types.ValidatorIndex `protobuf:"varint,3,opt,name=validator_index,json=validatorIndex,proto3" json:"validator_index,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.ValidatorIndex"`
	Signature       []byte                                             `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *SyncCommitteeMessage) Reset() {
	*x = SyncCommitteeMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncCommitteeMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncCommitteeMessage) ProtoMessage() {}

func (x *SyncCommitteeMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncCommitteeMessage.ProtoReflect.Descriptor instead.
func (*SyncCommitteeMessage) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{3}
}

func (x *SyncCommitteeMessage) GetSlot() github_com_prysmaticlabs_eth2_types.Slot {
	if x != nil {
		return x.Slot
	}
	return github_com_prysmaticlabs_eth2_types.Slot(0)
}

func (x *SyncCommitteeMessage) GetBeaconBlockRoot() []byte {
	if x != nil {
		return x.BeaconBlockRoot
	}
	return nil
}

func (x *SyncCommitteeMessage) GetValidatorIndex() github_com_prysmaticlabs_eth2_types.ValidatorIndex {
	if x != nil {
		return x.ValidatorIndex
	}
	return github_com_prysmaticlabs_eth2_types.ValidatorIndex(0)
}

func (x *SyncCommitteeMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type StateSyncCommitteesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StateId []byte                                     `protobuf:"bytes,1,opt,name=state_id,json=stateId,proto3" json:"state_id,omitempty"`
	Epoch   *github_com_prysmaticlabs_eth2_types.Epoch `protobuf:"varint,2,opt,name=epoch,proto3,oneof" json:"epoch,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Epoch"`
}

func (x *StateSyncCommitteesRequest) Reset() {
	*x = StateSyncCommitteesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateSyncCommitteesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateSyncCommitteesRequest) ProtoMessage() {}

func (x *StateSyncCommitteesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateSyncCommitteesRequest.ProtoReflect.Descriptor instead.
func (*StateSyncCommitteesRequest) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{4}
}

func (x *StateSyncCommitteesRequest) GetStateId() []byte {
	if x != nil {
		return x.StateId
	}
	return nil
}

func (x *StateSyncCommitteesRequest) GetEpoch() github_com_prysmaticlabs_eth2_types.Epoch {
	if x != nil && x.Epoch != nil {
		return *x.Epoch
	}
	return github_com_prysmaticlabs_eth2_types.Epoch(0)
}

type StateSyncCommitteesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data                *SyncCommitteeValidators `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	ExecutionOptimistic bool                     `protobuf:"varint,2,opt,name=execution_optimistic,json=executionOptimistic,proto3" json:"execution_optimistic,omitempty"`
}

func (x *StateSyncCommitteesResponse) Reset() {
	*x = StateSyncCommitteesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StateSyncCommitteesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StateSyncCommitteesResponse) ProtoMessage() {}

func (x *StateSyncCommitteesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StateSyncCommitteesResponse.ProtoReflect.Descriptor instead.
func (*StateSyncCommitteesResponse) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{5}
}

func (x *StateSyncCommitteesResponse) GetData() *SyncCommitteeValidators {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *StateSyncCommitteesResponse) GetExecutionOptimistic() bool {
	if x != nil {
		return x.ExecutionOptimistic
	}
	return false
}

type SyncCommitteeValidators struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Validators          []github_com_prysmaticlabs_eth2_types.ValidatorIndex `protobuf:"varint,1,rep,packed,name=validators,proto3" json:"validators,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.ValidatorIndex"`
	ValidatorAggregates []*SyncSubcommitteeValidators                        `protobuf:"bytes,2,rep,name=validator_aggregates,json=validatorAggregates,proto3" json:"validator_aggregates,omitempty"`
}

func (x *SyncCommitteeValidators) Reset() {
	*x = SyncCommitteeValidators{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncCommitteeValidators) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncCommitteeValidators) ProtoMessage() {}

func (x *SyncCommitteeValidators) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncCommitteeValidators.ProtoReflect.Descriptor instead.
func (*SyncCommitteeValidators) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{6}
}

func (x *SyncCommitteeValidators) GetValidators() []github_com_prysmaticlabs_eth2_types.ValidatorIndex {
	if x != nil {
		return x.Validators
	}
	return []github_com_prysmaticlabs_eth2_types.ValidatorIndex(nil)
}

func (x *SyncCommitteeValidators) GetValidatorAggregates() []*SyncSubcommitteeValidators {
	if x != nil {
		return x.ValidatorAggregates
	}
	return nil
}

type SyncSubcommitteeValidators struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Validators []github_com_prysmaticlabs_eth2_types.ValidatorIndex `protobuf:"varint,1,rep,packed,name=validators,proto3" json:"validators,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.ValidatorIndex"`
}

func (x *SyncSubcommitteeValidators) Reset() {
	*x = SyncSubcommitteeValidators{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncSubcommitteeValidators) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncSubcommitteeValidators) ProtoMessage() {}

func (x *SyncSubcommitteeValidators) ProtoReflect() protoreflect.Message {
	mi := &file_proto_eth_v2_sync_committee_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncSubcommitteeValidators.ProtoReflect.Descriptor instead.
func (*SyncSubcommitteeValidators) Descriptor() ([]byte, []int) {
	return file_proto_eth_v2_sync_committee_proto_rawDescGZIP(), []int{7}
}

func (x *SyncSubcommitteeValidators) GetValidators() []github_com_prysmaticlabs_eth2_types.ValidatorIndex {
	if x != nil {
		return x.Validators
	}
	return []github_com_prysmaticlabs_eth2_types.ValidatorIndex(nil)
}

var File_proto_eth_v2_sync_committee_proto protoreflect.FileDescriptor

var file_proto_eth_v2_sync_committee_proto_rawDesc = []byte{
	0x0a, 0x21, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x32, 0x2f, 0x73,
	0x79, 0x6e, 0x63, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74,
	0x68, 0x2e, 0x76, 0x32, 0x1a, 0x1b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f,
	0x65, 0x78, 0x74, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x61, 0x0a, 0x24, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x53, 0x79, 0x6e, 0x63, 0x43,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x39, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65,
	0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x68, 0x0a, 0x0d, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x12, 0x24, 0x0a, 0x07, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x42, 0x0a, 0x8a, 0xb5, 0x18, 0x06, 0x35, 0x31, 0x32, 0x2c,
	0x34, 0x38, 0x52, 0x07, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x31, 0x0a, 0x10, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x5f, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x34, 0x38, 0x52, 0x0f, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x50, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x22, 0x5e,
	0x0a, 0x21, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x50, 0x6f, 0x6f, 0x6c, 0x53, 0x79, 0x6e, 0x63,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x73, 0x12, 0x39, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68,
	0x2e, 0x76, 0x32, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65,
	0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x93,
	0x02, 0x0a, 0x14, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x40, 0x0a, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2c, 0x82, 0xb5, 0x18, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c,
	0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53,
	0x6c, 0x6f, 0x74, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x12, 0x32, 0x0a, 0x11, 0x62, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x33, 0x32, 0x52, 0x0f, 0x62, 0x65,
	0x61, 0x63, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x5f, 0x0a,
	0x0f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63,
	0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0e,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x24,
	0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x39, 0x36, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x22, 0x8b, 0x01, 0x0a, 0x1a, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x79,
	0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x73, 0x74, 0x61, 0x74, 0x65, 0x49, 0x64, 0x12, 0x48,
	0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2d, 0x82,
	0xb5, 0x18, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72,
	0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32,
	0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x48, 0x00, 0x52, 0x05,
	0x65, 0x70, 0x6f, 0x63, 0x68, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x65, 0x70, 0x6f,
	0x63, 0x68, 0x22, 0x8e, 0x01, 0x0a, 0x1b, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x79, 0x6e, 0x63,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x3c, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x28, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e,
	0x76, 0x32, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x31, 0x0a, 0x14, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6f, 0x70,
	0x74, 0x69, 0x6d, 0x69, 0x73, 0x74, 0x69, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13,
	0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x74, 0x69, 0x6d, 0x69, 0x73,
	0x74, 0x69, 0x63, 0x22, 0xd1, 0x01, 0x0a, 0x17, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x12,
	0x56, 0x0a, 0x0a, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62,
	0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0a, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x5e, 0x0a, 0x14, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x5f, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x53, 0x75, 0x62, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x73, 0x52, 0x13, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x41, 0x67, 0x67,
	0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x73, 0x22, 0x74, 0x0a, 0x1a, 0x53, 0x79, 0x6e, 0x63, 0x53,
	0x75, 0x62, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x56, 0x0a, 0x0a, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61,
	0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65,
	0x78, 0x52, 0x0a, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x42, 0x80, 0x01,
	0x0a, 0x13, 0x6f, 0x72, 0x67, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65,
	0x74, 0x68, 0x2e, 0x76, 0x32, 0x42, 0x12, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x74, 0x65, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69,
	0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x32, 0x3b, 0x65, 0x74, 0x68, 0xaa, 0x02, 0x0f, 0x45,
	0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x45, 0x74, 0x68, 0x2e, 0x56, 0x32, 0xca, 0x02,
	0x0f, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x5c, 0x45, 0x74, 0x68, 0x5c, 0x76, 0x32,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_eth_v2_sync_committee_proto_rawDescOnce sync.Once
	file_proto_eth_v2_sync_committee_proto_rawDescData = file_proto_eth_v2_sync_committee_proto_rawDesc
)

func file_proto_eth_v2_sync_committee_proto_rawDescGZIP() []byte {
	file_proto_eth_v2_sync_committee_proto_rawDescOnce.Do(func() {
		file_proto_eth_v2_sync_committee_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_eth_v2_sync_committee_proto_rawDescData)
	})
	return file_proto_eth_v2_sync_committee_proto_rawDescData
}

var file_proto_eth_v2_sync_committee_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_proto_eth_v2_sync_committee_proto_goTypes = []interface{}{
	(*SubmitSyncCommitteeSignaturesRequest)(nil), // 0: ethereum.eth.v2.SubmitSyncCommitteeSignaturesRequest
	(*SyncCommittee)(nil),                        // 1: ethereum.eth.v2.SyncCommittee
	(*SubmitPoolSyncCommitteeSignatures)(nil),    // 2: ethereum.eth.v2.SubmitPoolSyncCommitteeSignatures
	(*SyncCommitteeMessage)(nil),                 // 3: ethereum.eth.v2.SyncCommitteeMessage
	(*StateSyncCommitteesRequest)(nil),           // 4: ethereum.eth.v2.StateSyncCommitteesRequest
	(*StateSyncCommitteesResponse)(nil),          // 5: ethereum.eth.v2.StateSyncCommitteesResponse
	(*SyncCommitteeValidators)(nil),              // 6: ethereum.eth.v2.SyncCommitteeValidators
	(*SyncSubcommitteeValidators)(nil),           // 7: ethereum.eth.v2.SyncSubcommitteeValidators
}
var file_proto_eth_v2_sync_committee_proto_depIdxs = []int32{
	3, // 0: ethereum.eth.v2.SubmitSyncCommitteeSignaturesRequest.data:type_name -> ethereum.eth.v2.SyncCommitteeMessage
	3, // 1: ethereum.eth.v2.SubmitPoolSyncCommitteeSignatures.data:type_name -> ethereum.eth.v2.SyncCommitteeMessage
	6, // 2: ethereum.eth.v2.StateSyncCommitteesResponse.data:type_name -> ethereum.eth.v2.SyncCommitteeValidators
	7, // 3: ethereum.eth.v2.SyncCommitteeValidators.validator_aggregates:type_name -> ethereum.eth.v2.SyncSubcommitteeValidators
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_eth_v2_sync_committee_proto_init() }
func file_proto_eth_v2_sync_committee_proto_init() {
	if File_proto_eth_v2_sync_committee_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_eth_v2_sync_committee_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubmitSyncCommitteeSignaturesRequest); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncCommittee); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubmitPoolSyncCommitteeSignatures); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncCommitteeMessage); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StateSyncCommitteesRequest); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StateSyncCommitteesResponse); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncCommitteeValidators); i {
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
		file_proto_eth_v2_sync_committee_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncSubcommitteeValidators); i {
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
	file_proto_eth_v2_sync_committee_proto_msgTypes[4].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_eth_v2_sync_committee_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_eth_v2_sync_committee_proto_goTypes,
		DependencyIndexes: file_proto_eth_v2_sync_committee_proto_depIdxs,
		MessageInfos:      file_proto_eth_v2_sync_committee_proto_msgTypes,
	}.Build()
	File_proto_eth_v2_sync_committee_proto = out.File
	file_proto_eth_v2_sync_committee_proto_rawDesc = nil
	file_proto_eth_v2_sync_committee_proto_goTypes = nil
	file_proto_eth_v2_sync_committee_proto_depIdxs = nil
}
