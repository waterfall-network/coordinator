package mock

import (
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Exits []*eth.SignedVoluntaryExit
}

// PendingExits --
func (m *PoolMock) PendingExits(_ state.ReadOnlyBeaconState, _ types.Slot, _ bool) []*eth.SignedVoluntaryExit {
	return m.Exits
}

// InsertVoluntaryExit --
func (m *PoolMock) InsertVoluntaryExit(_ context.Context, _ state.ReadOnlyBeaconState, exit *eth.SignedVoluntaryExit) {
	m.Exits = append(m.Exits, exit)
}

// MarkIncluded --
func (*PoolMock) MarkIncluded(_ *eth.SignedVoluntaryExit) {
	panic("implement me")
}
