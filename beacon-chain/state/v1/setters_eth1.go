package v1

import (
	"bytes"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state/stateutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

// SetEth1Data for the beacon state.
func (b *BeaconState) SetEth1Data(val *ethpb.Eth1Data) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	b.state.Eth1Data = val
	b.markFieldAsDirty(eth1Data)
	return nil
}

// SetEth1DataVotes for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetEth1DataVotes(val []*ethpb.Eth1Data) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[eth1DataVotes].MinusRef()
	b.sharedFieldReferences[eth1DataVotes] = stateutil.NewRef(1)

	b.state.Eth1DataVotes = val
	b.markFieldAsDirty(eth1DataVotes)
	b.rebuildTrie[eth1DataVotes] = true
	return nil
}

// SetEth1DepositIndex for the beacon state.
func (b *BeaconState) SetEth1DepositIndex(val uint64) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	b.state.Eth1DepositIndex = val
	b.markFieldAsDirty(eth1DepositIndex)
	return nil
}

// AppendEth1DataVotes for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendEth1DataVotes(val *ethpb.Eth1Data) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	votes := b.state.Eth1DataVotes
	if b.sharedFieldReferences[eth1DataVotes].Refs() > 1 {
		// Copy elements in underlying array by reference.
		votes = make([]*ethpb.Eth1Data, len(b.state.Eth1DataVotes))
		copy(votes, b.state.Eth1DataVotes)
		b.sharedFieldReferences[eth1DataVotes].MinusRef()
		b.sharedFieldReferences[eth1DataVotes] = stateutil.NewRef(1)
	}

	b.state.Eth1DataVotes = append(votes, val)
	b.markFieldAsDirty(eth1DataVotes)
	b.addDirtyIndices(eth1DataVotes, []uint64{uint64(len(b.state.Eth1DataVotes) - 1)})
	return nil
}

// SetBlockVoting for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetBlockVoting(val []*ethpb.BlockVoting) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[blockVoting].MinusRef()
	b.sharedFieldReferences[blockVoting] = stateutil.NewRef(1)

	b.state.BlockVoting = val
	b.markFieldAsDirty(blockVoting)
	b.rebuildTrie[blockVoting] = true
	return nil
}

// AppendBlockVoting for the beacon state. Appends the new value.
func (b *BeaconState) AppendBlockVoting(val *ethpb.Attestation) error {
	if !b.hasInnerState() {
		return ErrNilInnerState
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	votes := b.state.BlockVoting

	if b.sharedFieldReferences[blockVoting].Refs() > 1 {
		// Copy elements in underlying array by reference.
		votes = make([]*ethpb.BlockVoting, len(b.state.BlockVoting))
		copy(votes, b.state.BlockVoting)
		b.sharedFieldReferences[blockVoting].MinusRef()
		b.sharedFieldReferences[blockVoting] = stateutil.NewRef(1)
	}

	addItem := true
	var err error
	addKey := val.GetData().BeaconBlockRoot
	for _, itm := range votes {
		if bytes.Equal(itm.Root, addKey) {
			addItem = false
			atts := append(itm.GetAttestations(), val)
			itm.Attestations, err = stateutil.Dedup(atts)
			if err != nil {
				return err
			}
		}
	}
	if addItem {
		newItem := &ethpb.BlockVoting{
			Root:         addKey,
			Attestations: []*ethpb.Attestation{val},
		}
		votes = append(votes, newItem)
	}

	b.state.BlockVoting = votes
	b.markFieldAsDirty(blockVoting)
	b.addDirtyIndices(blockVoting, []uint64{uint64(len(b.state.BlockVoting) - 1)})
	return nil
}
