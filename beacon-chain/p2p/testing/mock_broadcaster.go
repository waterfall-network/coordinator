package testing

import (
	"context"

	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// MockBroadcaster implements p2p.Broadcaster for testing.
type MockBroadcaster struct {
	BroadcastCalled       bool
	BroadcastMessages     []proto.Message
	BroadcastAttestations []*ethpb.Attestation
}

func (m *MockBroadcaster) BroadcastPrevoting(ctx context.Context, subnet uint64, sMsg *ethpb.PreVote) error {
	//TODO implement me
	panic("implement me")
}

// Broadcast records a broadcast occurred.
func (m *MockBroadcaster) Broadcast(_ context.Context, msg proto.Message) error {
	m.BroadcastCalled = true
	m.BroadcastMessages = append(m.BroadcastMessages, msg)
	return nil
}

// BroadcastAttestation records a broadcast occurred.
func (m *MockBroadcaster) BroadcastAttestation(_ context.Context, _ uint64, a *ethpb.Attestation) error {
	m.BroadcastCalled = true
	m.BroadcastAttestations = append(m.BroadcastAttestations, a)
	return nil
}

// BroadcastSyncCommitteeMessage records a broadcast occurred.
func (m *MockBroadcaster) BroadcastSyncCommitteeMessage(_ context.Context, _ uint64, _ *ethpb.SyncCommitteeMessage) error {
	m.BroadcastCalled = true
	return nil
}
