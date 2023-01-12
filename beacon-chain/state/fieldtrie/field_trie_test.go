package fieldtrie_test

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/go-bitfield"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/fieldtrie"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stateutil"
	stateTypes "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestFieldTrie_NewTrie(t *testing.T) {
	newState, _ := util.DeterministicGenesisState(t, 40)

	// 5 represents the enum value of state roots
	trie, err := fieldtrie.NewFieldTrie(5, stateTypes.BasicArray, newState.StateRoots(), uint64(params.BeaconConfig().SlotsPerHistoricalRoot))
	require.NoError(t, err)
	root, err := stateutil.RootsArrayHashTreeRoot(newState.StateRoots(), uint64(params.BeaconConfig().SlotsPerHistoricalRoot))
	require.NoError(t, err)
	newRoot, err := trie.TrieRoot()
	require.NoError(t, err)
	assert.Equal(t, root, newRoot)
}

func TestFieldTrie_RecomputeTrie(t *testing.T) {
	newState, _ := util.DeterministicGenesisState(t, 32)
	// 10 represents the enum value of validators
	trie, err := fieldtrie.NewFieldTrie(11, stateTypes.CompositeArray, newState.Validators(), params.BeaconConfig().ValidatorRegistryLimit)
	require.NoError(t, err)

	changedIdx := []uint64{2, 29}
	val1, err := newState.ValidatorAtIndex(10)
	require.NoError(t, err)
	val2, err := newState.ValidatorAtIndex(11)
	require.NoError(t, err)
	val1.Slashed = true
	val1.ExitEpoch = 20

	val2.Slashed = true
	val2.ExitEpoch = 40

	changedVals := []*ethpb.Validator{val1, val2}
	require.NoError(t, newState.UpdateValidatorAtIndex(types.ValidatorIndex(changedIdx[0]), changedVals[0]))
	require.NoError(t, newState.UpdateValidatorAtIndex(types.ValidatorIndex(changedIdx[1]), changedVals[1]))

	expectedRoot, err := stateutil.ValidatorRegistryRoot(newState.Validators())
	require.NoError(t, err)
	root, err := trie.RecomputeTrie(changedIdx, newState.Validators())
	require.NoError(t, err)
	assert.Equal(t, expectedRoot, root)
}

func TestFieldTrie_RecomputeTrieBlockVoting(t *testing.T) {
	newState, _ := util.DeterministicGenesisState(t, 32)
	// 10 represents the enum value of validators

	trie, err := fieldtrie.NewFieldTrie(21, stateTypes.CompositeArray, newState.BlockVoting(), 2048)
	require.NoError(t, err)

	changedIdx := []uint64{0}
	beaconBlockRoot := [32]byte{0x11, 0x22, 0x33}

	bitlist := bitfield.NewBitlist(128)
	bitlist.SetBitAt(0, true)
	bitlist.SetBitAt(1, true)

	err = newState.AppendBlockVotingAtt(&ethpb.Attestation{
		AggregationBits: bitlist,
		Data: &ethpb.AttestationData{
			Slot:            1,
			CommitteeIndex:  0,
			BeaconBlockRoot: beaconBlockRoot[:],
			Source: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  beaconBlockRoot[:],
			},
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  beaconBlockRoot[:],
			},
		},
		Signature: []byte{0x77},
	})
	require.NoError(t, err)
	err = newState.AppendBlockVotingAtt(&ethpb.Attestation{
		AggregationBits: bitlist,
		Data: &ethpb.AttestationData{
			Slot:            1,
			CommitteeIndex:  0,
			BeaconBlockRoot: beaconBlockRoot[:],
			Source: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  beaconBlockRoot[:],
			},
			Target: &ethpb.Checkpoint{
				Epoch: 0,
				Root:  beaconBlockRoot[:],
			},
		},
		Signature: []byte{0x77},
	})
	require.NoError(t, err)

	expectedRoot, err := stateutil.BlockVotingsRoot(newState.BlockVoting())
	require.NoError(t, err)
	root, err := trie.RecomputeTrie(changedIdx, newState.BlockVoting())
	require.NoError(t, err)
	assert.Equal(t, expectedRoot, root)
}

func TestFieldTrie_CopyTrieImmutable(t *testing.T) {
	newState, _ := util.DeterministicGenesisState(t, 32)
	// 12 represents the enum value of randao mixes.
	trie, err := fieldtrie.NewFieldTrie(13, stateTypes.BasicArray, newState.RandaoMixes(), uint64(params.BeaconConfig().EpochsPerHistoricalVector))
	require.NoError(t, err)

	newTrie := trie.CopyTrie()

	changedIdx := []uint64{2, 29}

	changedVals := [][32]byte{{'A', 'B'}, {'C', 'D'}}
	require.NoError(t, newState.UpdateRandaoMixesAtIndex(changedIdx[0], changedVals[0][:]))
	require.NoError(t, newState.UpdateRandaoMixesAtIndex(changedIdx[1], changedVals[1][:]))

	root, err := trie.RecomputeTrie(changedIdx, newState.RandaoMixes())
	require.NoError(t, err)
	newRoot, err := newTrie.TrieRoot()
	require.NoError(t, err)
	if root == newRoot {
		t.Errorf("Wanted roots to be different, but they are the same: %#x", root)
	}
}
