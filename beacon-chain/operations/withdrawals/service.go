package withdrawals

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"go.opencensus.io/trace"
)

// PoolManager maintains pending and seen withdrawals.
// This pool is used by proposers to insert withdrawals into new blocks.
type PoolManager interface {
	PendingWithdrawals(slot types.Slot, noLimit bool) []*ethpb.Withdrawal
	InsertWithdrawal(ctx context.Context, state state.ReadOnlyBeaconState, withdrawal *ethpb.Withdrawal)
	InsertWithdrawalByGwat(ctx context.Context, withdrawal *ethpb.Withdrawal)
	MarkIncluded(withdrawal *ethpb.Withdrawal)
}

// Pool is a concrete implementation of PoolManager.
type Pool struct {
	lock    sync.RWMutex
	pending []*ethpb.Withdrawal
}

// NewPool accepts a head fetcher (for reading the validator set) and returns an initialized
// withdrawals pool.
func NewPool() *Pool {
	return &Pool{
		pending: make([]*ethpb.Withdrawal, 0),
	}
}

// PendingWithdrawals returns withdrawals that are ready for inclusion at the given slot. This method will not
// return more than the block enforced MaxWithdrawals.
func (p *Pool) PendingWithdrawals(slot types.Slot, noLimit bool) []*ethpb.Withdrawal {
	p.lock.RLock()
	defer p.lock.RUnlock()

	// Allocate pending slice with a capacity of min(len(p.pending), maxWithdrawals) since the
	// array cannot exceed the max and is typically less than the max value.
	maxWithdrawals := params.BeaconConfig().MaxWithdrawals
	if noLimit {
		maxWithdrawals = uint64(len(p.pending))
	}
	pending := make([]*ethpb.Withdrawal, 0, maxWithdrawals)
	for _, e := range p.pending {
		if e.Epoch > slots.ToEpoch(slot) {
			continue
		}
		pending = append(pending, e)
		if uint64(len(pending)) == maxWithdrawals {
			break
		}
	}
	return pending
}

// InsertWithdrawal into the pool. This method is a no-op if the pending withdrawal already exists,
// or the validator is already withdrawaled.
func (p *Pool) InsertWithdrawalByGwat(ctx context.Context, withdrawal *ethpb.Withdrawal) {
	_, span := trace.StartSpan(ctx, "withdrawalPool.InsertWithdrawal")
	defer span.End()
	p.lock.Lock()
	defer p.lock.Unlock()

	// Prevent malformed messages from being inserted.
	if withdrawal == nil {
		return
	}
	if withdrawal.InitTxHash == nil {
		log.WithFields(log.Fields{
			"InitTxHash": withdrawal.InitTxHash,
		}).Warn("InsertWithdrawalByGwat malformed data: InitTxHash")
		return
	}

	//check exists
	if exists, _ := existsInList(p.pending, withdrawal); exists {
		return
	}

	// Insert into pending list and sort.
	p.pending = append(p.pending, withdrawal)
	sort.Slice(p.pending, func(i, j int) bool {
		return p.pending[i].Epoch < p.pending[j].Epoch
	})
}

// MarkIncluded is used when an withdrawal has been included in a beacon block. Every block seen by this
// node should call this method to include the withdrawal. This will remove the withdrawal from
// the pending withdrawals slice.
func (p *Pool) MarkIncluded(withdrawal *ethpb.Withdrawal) {
	p.lock.Lock()
	defer p.lock.Unlock()
	exists, index := existsInList(p.pending, withdrawal)
	if exists {
		// Withdrawal we want is present at p.pending[index], so we remove it.
		p.pending = append(p.pending[:index], p.pending[index+1:]...)
	}
}

// Binary search to check if the index exists in the list of pending withdrawals.
func existsInList(pending []*ethpb.Withdrawal, withdrawal *ethpb.Withdrawal) (bool, int) {
	if withdrawal == nil {
		return false, -1
	}
	for i, w := range pending {
		if bytes.Equal(w.InitTxHash, withdrawal.InitTxHash) {
			return true, i
		}
	}
	return false, -1
}
