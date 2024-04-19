package protoarray

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestForkChoiceCache_AddGet(t *testing.T) {
	testCache := NewForkChoiceCache()

	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
	}
	f.store.nodes = []*Node{
		{slot: 0, root: nrToHash(0), parent: NonExistentNode, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 1, root: nrToHash(1), parent: 0, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 2, root: nrToHash(2), parent: 1, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}, {'a', '3'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
	}
	f.balances = []uint64{123, 456, 798, 987, 654, 321}

	testCache.Add(f)

	cachedFc := testCache.Get(f.store.nodesIndices)
	require.DeepEqual(t, f, cachedFc)
}

func TestForkChoiceCache_SearchCompatibleFc(t *testing.T) {
	testCache := NewForkChoiceCache()

	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
	}
	f.store.nodes = []*Node{
		{slot: 0, root: nrToHash(0), parent: NonExistentNode, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 1, root: nrToHash(1), parent: 0, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 2, root: nrToHash(2), parent: 1, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}, {'a', '3'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
	}
	f.balances = []uint64{123, 456, 798, 987, 654, 321}

	testCache.Add(f)

	//search success
	nodesIndices := map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
		nrToHash(3): 3,
		nrToHash(4): 4,
	}
	expExld := map[[32]byte]uint64{
		nrToHash(3): 3,
		nrToHash(4): 4,
	}

	cachedFc, excluded := testCache.SearchCompatibleFc(nodesIndices)
	require.DeepEqual(t, f, cachedFc)
	require.DeepEqual(t, expExld, excluded)

	//search failed
	nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
	}
	expExld = nodesIndices
	var expFc *ForkChoice = nil

	cachedFc, excluded = testCache.SearchCompatibleFc(nodesIndices)
	require.DeepEqual(t, expFc, cachedFc)
	require.DeepEqual(t, expExld, excluded)
}

func TestNewForkChoiceCache_getCompatibleFc(t *testing.T) {
	memoCacheForkChoice := cacheForkChoice
	cacheForkChoice = NewForkChoiceCache()
	defer func() {
		cacheForkChoice = memoCacheForkChoice
	}()

	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
	}
	f.store.nodes = []*Node{
		{slot: 0, root: nrToHash(0), parent: NonExistentNode, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 1, root: nrToHash(1), parent: 0, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
		{slot: 2, root: nrToHash(2), parent: 1, spinesData: &SpinesData{
			spines:       gwatCommon.HashArray{{'a', '2'}, {'a', '3'}},
			prefix:       gwatCommon.HashArray{},
			finalization: gwatCommon.HashArray{{'a', '1'}},
			cpFinalized:  gwatCommon.HashArray{{'x', 'x', 'x'}},
		}},
	}
	f.balances = []uint64{123, 456, 798, 987, 654, 321}

	// expect: copy of f (no cache)
	nodesIndices := map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
	}
	expExld := map[[32]byte]uint64{}

	cachedFc, excluded := getCompatibleFc(nodesIndices, f)
	require.DeepEqual(t, f, cachedFc)
	require.DeepEqual(t, expExld, excluded)

	// expect: new empty instance (no cache)
	nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
		nrToHash(3): 3,
		nrToHash(4): 4,
	}
	expExld = nodesIndices
	expNewFc := New(f.store.justifiedEpoch, f.store.finalizedEpoch)

	cachedFc, excluded = getCompatibleFc(nodesIndices, f)
	require.DeepEqual(t, expNewFc, cachedFc)
	require.DeepEqual(t, expExld, excluded)

	//Add cached fc
	cacheForkChoice.Add(f)

	//expect: cached value
	nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
		nrToHash(2): 2,
		nrToHash(3): 3,
		nrToHash(4): 4,
	}
	expExld = map[[32]byte]uint64{
		nrToHash(3): 3,
		nrToHash(4): 4,
	}

	cachedFc, excluded = cacheForkChoice.SearchCompatibleFc(nodesIndices)
	require.DeepEqual(t, f, cachedFc)
	require.DeepEqual(t, expExld, excluded)

	// expect: new empty instance
	nodesIndices = map[[32]byte]uint64{
		nrToHash(0): 0,
		nrToHash(1): 1,
	}
	expExld = nodesIndices
	var expFc *ForkChoice = nil

	cachedFc, excluded = cacheForkChoice.SearchCompatibleFc(nodesIndices)
	require.DeepEqual(t, expFc, cachedFc)
	require.DeepEqual(t, expExld, excluded)
}
