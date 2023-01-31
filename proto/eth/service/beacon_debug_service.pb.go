// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.15.8
// source: proto/eth/service/beacon_debug_service.proto

package service

import (
	context "context"
	reflect "reflect"

	_ "github.com/golang/protobuf/protoc-gen-go/descriptor"
	empty "github.com/golang/protobuf/ptypes/empty"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/v1"
	v2 "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/eth/v2"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_proto_eth_service_beacon_debug_service_proto protoreflect.FileDescriptor

var file_proto_eth_service_beacon_debug_service_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67,
	0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14,
	0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f,
	0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31,
	0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76,
	0x32, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0x81, 0x07, 0x0a, 0x0b, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x44,
	0x65, 0x62, 0x75, 0x67, 0x12, 0x8e, 0x01, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x42, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1d, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65,
	0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75,
	0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x37, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x31, 0x12, 0x2f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x62, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2f, 0x7b, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x5f, 0x69, 0x64, 0x7d, 0x12, 0x98, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x42, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x53, 0x5a, 0x12, 0x1d, 0x2e, 0x65, 0x74,
	0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x65, 0x74, 0x68,
	0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x53, 0x5a, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x3b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x35, 0x12, 0x33, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65,
	0x62, 0x75, 0x67, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x73, 0x2f, 0x7b, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x73, 0x73, 0x7a,
	0x12, 0x94, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x56, 0x32, 0x12, 0x1f, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d,
	0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x56, 0x32, 0x1a, 0x26, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75,
	0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x56, 0x32, 0x22, 0x37,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x31, 0x12, 0x2f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x32, 0x2f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x62,
	0x65, 0x61, 0x63, 0x6f, 0x6e, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2f, 0x7b, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x7d, 0x12, 0x9e, 0x01, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x42,
	0x65, 0x61, 0x63, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x53, 0x5a, 0x56, 0x32, 0x12,
	0x1f, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76,
	0x32, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x32,
	0x1a, 0x29, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e,
	0x76, 0x32, 0x2e, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x53,
	0x5a, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x56, 0x32, 0x22, 0x3b, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x35, 0x12, 0x33, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65,
	0x74, 0x68, 0x2f, 0x76, 0x32, 0x2f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x62, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x2f, 0x7b, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x5f, 0x69, 0x64, 0x7d, 0x2f, 0x73, 0x73, 0x7a, 0x12, 0x84, 0x01, 0x0a, 0x13, 0x4c, 0x69, 0x73,
	0x74, 0x46, 0x6f, 0x72, 0x6b, 0x43, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x48, 0x65, 0x61, 0x64, 0x73,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x28, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72,
	0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x72, 0x6b, 0x43,
	0x68, 0x6f, 0x69, 0x63, 0x65, 0x48, 0x65, 0x61, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x2b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x25, 0x12, 0x23, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x62,
	0x75, 0x67, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x2f, 0x68, 0x65, 0x61, 0x64, 0x73, 0x12,
	0x86, 0x01, 0x0a, 0x15, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x6f, 0x72, 0x6b, 0x43, 0x68, 0x6f, 0x69,
	0x63, 0x65, 0x48, 0x65, 0x61, 0x64, 0x73, 0x56, 0x32, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x1a, 0x28, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68,
	0x2e, 0x76, 0x32, 0x2e, 0x46, 0x6f, 0x72, 0x6b, 0x43, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x48, 0x65,
	0x61, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x25, 0x12, 0x23, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65,
	0x74, 0x68, 0x2f, 0x76, 0x32, 0x2f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x62, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x2f, 0x68, 0x65, 0x61, 0x64, 0x73, 0x42, 0xa2, 0x01, 0x0a, 0x18, 0x6f, 0x72, 0x67,
	0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x65, 0x74, 0x68, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x17, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x44, 0x65, 0x62,
	0x75, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x61, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x6c, 0x6c, 0x2d, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2f, 0x63, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x74, 0x68, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xaa,
	0x02, 0x14, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x45, 0x74, 0x68, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xca, 0x02, 0x14, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75,
	0x6d, 0x5c, 0x45, 0x74, 0x68, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_proto_eth_service_beacon_debug_service_proto_goTypes = []interface{}{
	(*v1.StateRequest)(nil),             // 0: ethereum.eth.v1.StateRequest
	(*v2.StateRequestV2)(nil),           // 1: ethereum.eth.v2.StateRequestV2
	(*empty.Empty)(nil),                 // 2: google.protobuf.Empty
	(*v1.BeaconStateResponse)(nil),      // 3: ethereum.eth.v1.BeaconStateResponse
	(*v1.BeaconStateSSZResponse)(nil),   // 4: ethereum.eth.v1.BeaconStateSSZResponse
	(*v2.BeaconStateResponseV2)(nil),    // 5: ethereum.eth.v2.BeaconStateResponseV2
	(*v2.BeaconStateSSZResponseV2)(nil), // 6: ethereum.eth.v2.BeaconStateSSZResponseV2
	(*v1.ForkChoiceHeadsResponse)(nil),  // 7: ethereum.eth.v1.ForkChoiceHeadsResponse
	(*v2.ForkChoiceHeadsResponse)(nil),  // 8: ethereum.eth.v2.ForkChoiceHeadsResponse
}
var file_proto_eth_service_beacon_debug_service_proto_depIdxs = []int32{
	0, // 0: ethereum.eth.service.BeaconDebug.GetBeaconState:input_type -> ethereum.eth.v1.StateRequest
	0, // 1: ethereum.eth.service.BeaconDebug.GetBeaconStateSSZ:input_type -> ethereum.eth.v1.StateRequest
	1, // 2: ethereum.eth.service.BeaconDebug.GetBeaconStateV2:input_type -> ethereum.eth.v2.StateRequestV2
	1, // 3: ethereum.eth.service.BeaconDebug.GetBeaconStateSSZV2:input_type -> ethereum.eth.v2.StateRequestV2
	2, // 4: ethereum.eth.service.BeaconDebug.ListForkChoiceHeads:input_type -> google.protobuf.Empty
	2, // 5: ethereum.eth.service.BeaconDebug.ListForkChoiceHeadsV2:input_type -> google.protobuf.Empty
	3, // 6: ethereum.eth.service.BeaconDebug.GetBeaconState:output_type -> ethereum.eth.v1.BeaconStateResponse
	4, // 7: ethereum.eth.service.BeaconDebug.GetBeaconStateSSZ:output_type -> ethereum.eth.v1.BeaconStateSSZResponse
	5, // 8: ethereum.eth.service.BeaconDebug.GetBeaconStateV2:output_type -> ethereum.eth.v2.BeaconStateResponseV2
	6, // 9: ethereum.eth.service.BeaconDebug.GetBeaconStateSSZV2:output_type -> ethereum.eth.v2.BeaconStateSSZResponseV2
	7, // 10: ethereum.eth.service.BeaconDebug.ListForkChoiceHeads:output_type -> ethereum.eth.v1.ForkChoiceHeadsResponse
	8, // 11: ethereum.eth.service.BeaconDebug.ListForkChoiceHeadsV2:output_type -> ethereum.eth.v2.ForkChoiceHeadsResponse
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_eth_service_beacon_debug_service_proto_init() }
func file_proto_eth_service_beacon_debug_service_proto_init() {
	if File_proto_eth_service_beacon_debug_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_eth_service_beacon_debug_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_eth_service_beacon_debug_service_proto_goTypes,
		DependencyIndexes: file_proto_eth_service_beacon_debug_service_proto_depIdxs,
	}.Build()
	File_proto_eth_service_beacon_debug_service_proto = out.File
	file_proto_eth_service_beacon_debug_service_proto_rawDesc = nil
	file_proto_eth_service_beacon_debug_service_proto_goTypes = nil
	file_proto_eth_service_beacon_debug_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BeaconDebugClient is the client API for BeaconDebug service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BeaconDebugClient interface {
	GetBeaconState(ctx context.Context, in *v1.StateRequest, opts ...grpc.CallOption) (*v1.BeaconStateResponse, error)
	GetBeaconStateSSZ(ctx context.Context, in *v1.StateRequest, opts ...grpc.CallOption) (*v1.BeaconStateSSZResponse, error)
	GetBeaconStateV2(ctx context.Context, in *v2.StateRequestV2, opts ...grpc.CallOption) (*v2.BeaconStateResponseV2, error)
	GetBeaconStateSSZV2(ctx context.Context, in *v2.StateRequestV2, opts ...grpc.CallOption) (*v2.BeaconStateSSZResponseV2, error)
	ListForkChoiceHeads(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*v1.ForkChoiceHeadsResponse, error)
	ListForkChoiceHeadsV2(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*v2.ForkChoiceHeadsResponse, error)
}

type beaconDebugClient struct {
	cc grpc.ClientConnInterface
}

func NewBeaconDebugClient(cc grpc.ClientConnInterface) BeaconDebugClient {
	return &beaconDebugClient{cc}
}

func (c *beaconDebugClient) GetBeaconState(ctx context.Context, in *v1.StateRequest, opts ...grpc.CallOption) (*v1.BeaconStateResponse, error) {
	out := new(v1.BeaconStateResponse)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/GetBeaconState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconDebugClient) GetBeaconStateSSZ(ctx context.Context, in *v1.StateRequest, opts ...grpc.CallOption) (*v1.BeaconStateSSZResponse, error) {
	out := new(v1.BeaconStateSSZResponse)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/GetBeaconStateSSZ", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconDebugClient) GetBeaconStateV2(ctx context.Context, in *v2.StateRequestV2, opts ...grpc.CallOption) (*v2.BeaconStateResponseV2, error) {
	out := new(v2.BeaconStateResponseV2)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/GetBeaconStateV2", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconDebugClient) GetBeaconStateSSZV2(ctx context.Context, in *v2.StateRequestV2, opts ...grpc.CallOption) (*v2.BeaconStateSSZResponseV2, error) {
	out := new(v2.BeaconStateSSZResponseV2)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/GetBeaconStateSSZV2", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconDebugClient) ListForkChoiceHeads(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*v1.ForkChoiceHeadsResponse, error) {
	out := new(v1.ForkChoiceHeadsResponse)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/ListForkChoiceHeads", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *beaconDebugClient) ListForkChoiceHeadsV2(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*v2.ForkChoiceHeadsResponse, error) {
	out := new(v2.ForkChoiceHeadsResponse)
	err := c.cc.Invoke(ctx, "/ethereum.eth.service.BeaconDebug/ListForkChoiceHeadsV2", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BeaconDebugServer is the server API for BeaconDebug service.
type BeaconDebugServer interface {
	GetBeaconState(context.Context, *v1.StateRequest) (*v1.BeaconStateResponse, error)
	GetBeaconStateSSZ(context.Context, *v1.StateRequest) (*v1.BeaconStateSSZResponse, error)
	GetBeaconStateV2(context.Context, *v2.StateRequestV2) (*v2.BeaconStateResponseV2, error)
	GetBeaconStateSSZV2(context.Context, *v2.StateRequestV2) (*v2.BeaconStateSSZResponseV2, error)
	ListForkChoiceHeads(context.Context, *empty.Empty) (*v1.ForkChoiceHeadsResponse, error)
	ListForkChoiceHeadsV2(context.Context, *empty.Empty) (*v2.ForkChoiceHeadsResponse, error)
}

// UnimplementedBeaconDebugServer can be embedded to have forward compatible implementations.
type UnimplementedBeaconDebugServer struct {
}

func (*UnimplementedBeaconDebugServer) GetBeaconState(context.Context, *v1.StateRequest) (*v1.BeaconStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBeaconState not implemented")
}
func (*UnimplementedBeaconDebugServer) GetBeaconStateSSZ(context.Context, *v1.StateRequest) (*v1.BeaconStateSSZResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBeaconStateSSZ not implemented")
}
func (*UnimplementedBeaconDebugServer) GetBeaconStateV2(context.Context, *v2.StateRequestV2) (*v2.BeaconStateResponseV2, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBeaconStateV2 not implemented")
}
func (*UnimplementedBeaconDebugServer) GetBeaconStateSSZV2(context.Context, *v2.StateRequestV2) (*v2.BeaconStateSSZResponseV2, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBeaconStateSSZV2 not implemented")
}
func (*UnimplementedBeaconDebugServer) ListForkChoiceHeads(context.Context, *empty.Empty) (*v1.ForkChoiceHeadsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListForkChoiceHeads not implemented")
}
func (*UnimplementedBeaconDebugServer) ListForkChoiceHeadsV2(context.Context, *empty.Empty) (*v2.ForkChoiceHeadsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListForkChoiceHeadsV2 not implemented")
}

func RegisterBeaconDebugServer(s *grpc.Server, srv BeaconDebugServer) {
	s.RegisterService(&_BeaconDebug_serviceDesc, srv)
}

func _BeaconDebug_GetBeaconState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.StateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).GetBeaconState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/GetBeaconState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).GetBeaconState(ctx, req.(*v1.StateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconDebug_GetBeaconStateSSZ_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.StateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).GetBeaconStateSSZ(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/GetBeaconStateSSZ",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).GetBeaconStateSSZ(ctx, req.(*v1.StateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconDebug_GetBeaconStateV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v2.StateRequestV2)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).GetBeaconStateV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/GetBeaconStateV2",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).GetBeaconStateV2(ctx, req.(*v2.StateRequestV2))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconDebug_GetBeaconStateSSZV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v2.StateRequestV2)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).GetBeaconStateSSZV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/GetBeaconStateSSZV2",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).GetBeaconStateSSZV2(ctx, req.(*v2.StateRequestV2))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconDebug_ListForkChoiceHeads_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).ListForkChoiceHeads(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/ListForkChoiceHeads",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).ListForkChoiceHeads(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BeaconDebug_ListForkChoiceHeadsV2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BeaconDebugServer).ListForkChoiceHeadsV2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ethereum.eth.service.BeaconDebug/ListForkChoiceHeadsV2",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BeaconDebugServer).ListForkChoiceHeadsV2(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _BeaconDebug_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ethereum.eth.service.BeaconDebug",
	HandlerType: (*BeaconDebugServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetBeaconState",
			Handler:    _BeaconDebug_GetBeaconState_Handler,
		},
		{
			MethodName: "GetBeaconStateSSZ",
			Handler:    _BeaconDebug_GetBeaconStateSSZ_Handler,
		},
		{
			MethodName: "GetBeaconStateV2",
			Handler:    _BeaconDebug_GetBeaconStateV2_Handler,
		},
		{
			MethodName: "GetBeaconStateSSZV2",
			Handler:    _BeaconDebug_GetBeaconStateSSZV2_Handler,
		},
		{
			MethodName: "ListForkChoiceHeads",
			Handler:    _BeaconDebug_ListForkChoiceHeads_Handler,
		},
		{
			MethodName: "ListForkChoiceHeadsV2",
			Handler:    _BeaconDebug_ListForkChoiceHeadsV2_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/eth/service/beacon_debug_service.proto",
}
