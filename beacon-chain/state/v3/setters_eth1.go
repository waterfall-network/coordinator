package v3

import (
	"bytes"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state/stateutil"
	fieldparams "github.com/waterfall-foundation/coordinator/config/fieldparams"
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

// AddBlockVoting adds or update the new BlockVoting data
// for the beacon state.
func (b *BeaconState) AddBlockVoting(root []byte, totalAttrs uint64, candidates []byte) error {
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
	dirtyIxs := []uint64{}
	addItem := true
	addKey := root
	for i, itm := range votes {
		if bytes.Equal(itm.Root, addKey) {
			dirtyIxs = append(dirtyIxs, uint64(i))
			addItem = false
			itm.TotalAttesters = totalAttrs
			itm.Candidates = candidates
		}
	}
	if addItem {
		newItem := &ethpb.BlockVoting{
			Root:           addKey,
			Attestations:   []*ethpb.Attestation{},
			TotalAttesters: totalAttrs,
			Candidates:     candidates,
		}
		votes = append(votes, newItem)
		dirtyIxs = append(dirtyIxs, uint64(len(votes)-1))
	}
	b.state.BlockVoting = votes
	b.markFieldAsDirty(blockVoting)
	b.addDirtyIndices(blockVoting, dirtyIxs)
	return nil
}

// AppendBlockVotingAtt appends the new attestation to BlockVoting data
// for the beacon state.
func (b *BeaconState) AppendBlockVotingAtt(val *ethpb.Attestation) error {
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
	dirtyIxs := []uint64{}
	addItem := true
	var err error
	addKey := val.GetData().BeaconBlockRoot
	for i, itm := range votes {
		if bytes.Equal(itm.Root, addKey) {
			dirtyIxs = append(dirtyIxs, uint64(i))
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
			Root:           addKey,
			Attestations:   []*ethpb.Attestation{val},
			TotalAttesters: 0,
			Candidates:     nil,
		}
		votes = append(votes, newItem)
		dirtyIxs = append(dirtyIxs, uint64(len(votes)-1))
	}
	b.state.BlockVoting = votes
	b.markFieldAsDirty(blockVoting)
	b.addDirtyIndices(blockVoting, dirtyIxs)
	return nil
}

// RemoveBlockVoting removes deprecated values from BlockVoting
// of the beacon state.
func (b *BeaconState) RemoveBlockVoting(roots [][]byte) error {
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
	upVotes := []*ethpb.BlockVoting{}
	for _, itm := range votes {
		for _, rmRoot := range roots {
			if !bytes.Equal(itm.Root, rmRoot) {
				upVotes = append(upVotes, itm)
			}
		}
	}

	b.state.BlockVoting = upVotes
	b.markFieldAsDirty(blockVoting)
	dirtyIxs := make([]uint64, len(b.state.BlockVoting))
	err := b.resetFieldTrie(blockVoting, b.state.BlockVoting, fieldparams.BlockVotingLength)
	if err != nil {
		return err
	}
	for i := 0; i < len(b.state.BlockVoting); i++ {
		dirtyIxs[i] = uint64(i)
	}
	b.addDirtyIndices(blockVoting, dirtyIxs)
	return nil
}
