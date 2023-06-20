package mock

import (
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Exits []*eth.VoluntaryExit
}

func (m *PoolMock) InsertVoluntaryExitByGwat(ctx context.Context, exit *eth.VoluntaryExit) {
	//TODO implement me
	panic("implement me")
}

// PendingExits --
func (m *PoolMock) PendingExits(_ state.ReadOnlyBeaconState, _ types.Slot, _ bool) []*eth.VoluntaryExit {
	return m.Exits
}

// InsertVoluntaryExit --
func (m *PoolMock) InsertVoluntaryExit(_ context.Context, _ state.ReadOnlyBeaconState, exit *eth.VoluntaryExit) {
	m.Exits = append(m.Exits, exit)
}

// MarkIncluded --
func (*PoolMock) MarkIncluded(_ *eth.VoluntaryExit) {
	panic("implement me")
}
