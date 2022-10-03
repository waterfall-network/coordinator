package blockchain

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/beacon-chain/blockchain/store"
	testDB "github.com/waterfall-foundation/coordinator/beacon-chain/db/testing"
	"github.com/waterfall-foundation/coordinator/beacon-chain/forkchoice/protoarray"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state/stategen"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestService_newSlot(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	fcs := protoarray.New(0, 0, [32]byte{})
	opts := []Option{
		WithDatabase(beaconDB),
		WithStateGen(stategen.New(beaconDB)),
		WithForkChoiceStore(fcs),
	}
	ctx := context.Background()

	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 0, [32]byte{}, [32]byte{}, [32]byte{}, 0, 0))        // genesis
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 32, [32]byte{'a'}, [32]byte{}, [32]byte{}, 0, 0))    // finalized
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 64, [32]byte{'b'}, [32]byte{'a'}, [32]byte{}, 0, 0)) // justified
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 96, [32]byte{'c'}, [32]byte{'a'}, [32]byte{}, 0, 0)) // best justified
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 97, [32]byte{'d'}, [32]byte{}, [32]byte{}, 0, 0))    // bad

	type args struct {
		slot          types.Slot
		finalized     *ethpb.Checkpoint
		justified     *ethpb.Checkpoint
		bestJustified *ethpb.Checkpoint
		shouldEqual   bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Not epoch boundary. No change",
			args: args{
				slot:          params.BeaconConfig().SlotsPerEpoch + 1,
				finalized:     &ethpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'a'}, 32)},
				justified:     &ethpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'b'}, 32)},
				bestJustified: &ethpb.Checkpoint{Epoch: 3, Root: bytesutil.PadTo([]byte{'c'}, 32)},
				shouldEqual:   false,
			},
		},
		{
			name: "Justified higher than best justified. No change",
			args: args{
				slot:          params.BeaconConfig().SlotsPerEpoch,
				finalized:     &ethpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'a'}, 32)},
				justified:     &ethpb.Checkpoint{Epoch: 3, Root: bytesutil.PadTo([]byte{'b'}, 32)},
				bestJustified: &ethpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'c'}, 32)},
				shouldEqual:   false,
			},
		},
		{
			name: "Best justified not on the same chain as finalized. No change",
			args: args{
				slot:          params.BeaconConfig().SlotsPerEpoch,
				finalized:     &ethpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'a'}, 32)},
				justified:     &ethpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'b'}, 32)},
				bestJustified: &ethpb.Checkpoint{Epoch: 3, Root: bytesutil.PadTo([]byte{'d'}, 32)},
				shouldEqual:   false,
			},
		},
		{
			name: "Best justified on the same chain as finalized. Yes change",
			args: args{
				slot:          params.BeaconConfig().SlotsPerEpoch,
				finalized:     &ethpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'a'}, 32)},
				justified:     &ethpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'b'}, 32)},
				bestJustified: &ethpb.Checkpoint{Epoch: 3, Root: bytesutil.PadTo([]byte{'c'}, 32)},
				shouldEqual:   true,
			},
		},
	}
	for _, test := range tests {
		service, err := NewService(ctx, opts...)
		require.NoError(t, err)
		s := store.New(test.args.justified, test.args.finalized)
		s.SetBestJustifiedCheckpt(test.args.bestJustified)
		service.store = s

		require.NoError(t, service.NewSlot(ctx, test.args.slot))
		if test.args.shouldEqual {
			require.DeepSSZEqual(t, service.store.BestJustifiedCheckpt(), service.store.JustifiedCheckpt())
		} else {
			require.DeepNotSSZEqual(t, service.store.BestJustifiedCheckpt(), service.store.JustifiedCheckpt())
		}
	}
}
