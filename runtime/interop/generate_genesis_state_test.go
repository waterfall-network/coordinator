package interop_test

import (
	"context"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/container/trie"
	eth "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/interop"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestGenerateGenesisState(t *testing.T) {
	numValidators := uint64(64)
	privKeys, pubKeys, err := interop.DeterministicallyGenerateKeys(0 /*startIndex*/, numValidators)
	require.NoError(t, err)
	depositDataItems, depositDataRoots, err := interop.DepositDataFromKeys(privKeys, pubKeys)
	require.NoError(t, err)
	tr, err := trie.GenerateTrieFromItems(depositDataRoots, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)
	deposits, err := interop.GenerateDepositsFromData(depositDataItems, tr)
	require.NoError(t, err)
	root := tr.HashTreeRoot()
	genesisState, err := transition.GenesisBeaconState(context.Background(), deposits, 0, &eth.Eth1Data{
		DepositRoot:  root[:],
		DepositCount: uint64(len(deposits)),
	})
	require.NoError(t, err)
	want := int(numValidators)
	assert.Equal(t, want, genesisState.NumValidators())
	assert.Equal(t, uint64(0), genesisState.GenesisTime())
}
