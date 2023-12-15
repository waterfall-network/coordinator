package mock

import (
	"bytes"
	"context"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Withdrawals []*eth.Withdrawal
}

func (m *PoolMock) InsertWithdrawalByGwat(ctx context.Context, withdrawal *eth.Withdrawal) {
	m.Withdrawals = append(m.Withdrawals, withdrawal)
}

// PendingWithdrawals --
func (m *PoolMock) PendingWithdrawals(_ state.ReadOnlyBeaconState, _ types.Slot, _ bool) []*eth.Withdrawal {
	return m.Withdrawals
}

// InsertWithdrawal --
func (m *PoolMock) InsertWithdrawal(_ context.Context, withdrawal *eth.Withdrawal) {
	m.Withdrawals = append(m.Withdrawals, withdrawal)
}

// MarkIncluded --
func (m *PoolMock) MarkIncluded(withdrawal *eth.Withdrawal) {
	res := make([]*eth.Withdrawal, 0, len(m.Withdrawals))
	for _, w := range m.Withdrawals {
		if bytes.Equal(w.InitTxHash, withdrawal.InitTxHash) {
			continue
		}
		res = append(res, w)
	}
	m.Withdrawals = res
}
