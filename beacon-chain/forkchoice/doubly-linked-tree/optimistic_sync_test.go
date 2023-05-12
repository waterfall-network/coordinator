package doublylinkedtree

import (
	"context"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

// We test the algorithm to update a node from SYNCING to INVALID
// We start with the same diagram as above:
//
//	              E -- F
//	             /
//	       C -- D
//	      /      \
//	A -- B        G -- H -- I
//	      \        \
//	       J        -- K -- L
//
// And every block in the Fork choice is optimistic.
func TestPruneInvalid(t *testing.T) {
	tests := []struct {
		root             [32]byte // the root of the new INVALID block
		payload          [32]byte // the last valid hash
		wantedNodeNumber int
		wantedRoots      [][32]byte
	}{
		{
			[32]byte{'j'},
			[32]byte{'B'},
			12,
			[][32]byte{{'j'}},
		},
		{
			[32]byte{'c'},
			[32]byte{'B'},
			4,
			[][32]byte{{'f'}, {'e'}, {'i'}, {'h'}, {'l'},
				{'k'}, {'g'}, {'d'}, {'c'}},
		},
		{
			[32]byte{'i'},
			[32]byte{'H'},
			12,
			[][32]byte{{'i'}},
		},
		{
			[32]byte{'h'},
			[32]byte{'G'},
			11,
			[][32]byte{{'i'}, {'h'}},
		},
		{
			[32]byte{'g'},
			[32]byte{'D'},
			8,
			[][32]byte{{'i'}, {'h'}, {'l'}, {'k'}, {'g'}},
		},
		{
			[32]byte{'i'},
			[32]byte{'D'},
			8,
			[][32]byte{{'i'}, {'h'}, {'l'}, {'k'}, {'g'}},
		},
		{
			[32]byte{'f'},
			[32]byte{'D'},
			11,
			[][32]byte{{'f'}, {'e'}},
		},
		{
			[32]byte{'h'},
			[32]byte{'C'},
			5,
			[][32]byte{
				{'f'},
				{'e'},
				{'i'},
				{'h'},
				{'l'},
				{'k'},
				{'g'},
				{'d'},
			},
		},
		{
			[32]byte{'g'},
			[32]byte{'E'},
			8,
			[][32]byte{{'i'}, {'h'}, {'l'}, {'k'}, {'g'}},
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		f := setup(1, 1)

		require.NoError(t, f.InsertOptimisticBlock(ctx, 100, [32]byte{'a'}, params.BeaconConfig().ZeroHash, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 101, [32]byte{'b'}, [32]byte{'a'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 102, [32]byte{'c'}, [32]byte{'b'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 102, [32]byte{'j'}, [32]byte{'b'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 103, [32]byte{'d'}, [32]byte{'c'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 104, [32]byte{'e'}, [32]byte{'d'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 104, [32]byte{'g'}, [32]byte{'d'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 105, [32]byte{'f'}, [32]byte{'e'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 105, [32]byte{'h'}, [32]byte{'g'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 105, [32]byte{'k'}, [32]byte{'g'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 106, [32]byte{'i'}, [32]byte{'h'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
		require.NoError(t, f.InsertOptimisticBlock(ctx, 106, [32]byte{'l'}, [32]byte{'k'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))

		roots, err := f.store.setOptimisticToInvalid(context.Background(), tc.root, tc.payload)
		require.NoError(t, err)
		require.DeepEqual(t, tc.wantedRoots, roots)
		require.Equal(t, tc.wantedNodeNumber, f.NodeCount())
	}
}

// This is a regression test (10445)
func TestSetOptimisticToInvalid_ProposerBoost(t *testing.T) {
	ctx := context.Background()
	f := setup(1, 1)

	require.NoError(t, f.InsertOptimisticBlock(ctx, 100, [32]byte{'a'}, params.BeaconConfig().ZeroHash, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	require.NoError(t, f.InsertOptimisticBlock(ctx, 101, [32]byte{'b'}, [32]byte{'a'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	require.NoError(t, f.InsertOptimisticBlock(ctx, 101, [32]byte{'c'}, [32]byte{'b'}, 1, 1, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	f.store.proposerBoostLock.Lock()
	f.store.proposerBoostRoot = [32]byte{'c'}
	f.store.previousProposerBoostScore = 10
	f.store.previousProposerBoostRoot = [32]byte{'b'}
	f.store.proposerBoostLock.Unlock()

	_, err := f.SetOptimisticToInvalid(ctx, [32]byte{'c'}, [32]byte{'A'})
	require.NoError(t, err)
	f.store.proposerBoostLock.RLock()
	require.Equal(t, uint64(0), f.store.previousProposerBoostScore)
	require.DeepEqual(t, [32]byte{}, f.store.proposerBoostRoot)
	require.DeepEqual(t, params.BeaconConfig().ZeroHash, f.store.previousProposerBoostRoot)
	f.store.proposerBoostLock.RUnlock()
}
