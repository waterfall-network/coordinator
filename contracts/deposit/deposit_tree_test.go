package deposit_test

import (
	"strconv"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/container/trie"
	depositcontract "gitlab.waterfall.network/waterfall/protocol/coordinator/contracts/deposit/mock"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/interop"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/accounts/abi/bind"
)

func TestDepositTrieRoot_OK(t *testing.T) {
	testAcc, err := depositcontract.Setup()
	require.NoError(t, err)

	localTrie, err := trie.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)

	depRoot, err := testAcc.Contract.GetDepositRoot(&bind.CallOpts{})
	require.NoError(t, err)

	assert.Equal(t, depRoot, localTrie.HashTreeRoot(), "Local deposit trie root and contract deposit trie root are not equal")

	privKeys, pubKeys, err := interop.DeterministicallyGenerateKeys(0 /*startIndex*/, 101)
	require.NoError(t, err)
	depositDataItems, depositDataRoots, err := interop.DepositDataFromKeys(privKeys, pubKeys)
	require.NoError(t, err)

	testAcc.TxOpts.Value = depositcontract.Amount3200Wat()

	for i := 0; i < 100; i++ {
		data := depositDataItems[i]
		dataRoot := [32]byte{}
		copy(dataRoot[:], depositDataRoots[i])

		_, err := testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, dataRoot)
		require.NoError(t, err, "Could not deposit to deposit contract")

		testAcc.Backend.Commit()
		item, err := data.HashTreeRoot()
		require.NoError(t, err)

		assert.NoError(t, localTrie.Insert(item[:], i))
		depRoot, err = testAcc.Contract.GetDepositRoot(&bind.CallOpts{})
		require.NoError(t, err)
		assert.Equal(t, depRoot, localTrie.HashTreeRoot(), "Local deposit trie root and contract deposit trie root are not equal for index %d", i)
	}
}

func TestDepositTrieRoot_Fail(t *testing.T) {
	testAcc, err := depositcontract.Setup()
	require.NoError(t, err)

	localTrie, err := trie.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)

	depRoot, err := testAcc.Contract.GetDepositRoot(&bind.CallOpts{})
	require.NoError(t, err)

	assert.Equal(t, depRoot, localTrie.HashTreeRoot(), "Local deposit trie root and contract deposit trie root are not equal")

	privKeys, pubKeys, err := interop.DeterministicallyGenerateKeys(0 /*startIndex*/, 101)
	require.NoError(t, err)
	depositDataItems, depositDataRoots, err := interop.DepositDataFromKeys(privKeys, pubKeys)
	require.NoError(t, err)
	testAcc.TxOpts.Value = depositcontract.Amount3200Wat()

	for i := 0; i < 100; i++ {
		data := depositDataItems[i]
		dataRoot := [32]byte{}
		copy(dataRoot[:], depositDataRoots[i])

		_, err := testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, dataRoot)
		require.NoError(t, err, "Could not deposit to deposit contract")

		// Change an element in the data when storing locally
		copy(data.PublicKey, strconv.Itoa(i+10))

		testAcc.Backend.Commit()
		item, err := data.HashTreeRoot()
		require.NoError(t, err)

		assert.NoError(t, localTrie.Insert(item[:], i))

		depRoot, err = testAcc.Contract.GetDepositRoot(&bind.CallOpts{})
		require.NoError(t, err)

		assert.NotEqual(t, depRoot, localTrie.HashTreeRoot(), "Local deposit trie root and contract deposit trie root are equal for index %d", i)
	}
}
