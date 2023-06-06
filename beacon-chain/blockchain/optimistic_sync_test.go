package blockchain

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	testDB "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/forkchoice/protoarray"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func Test_GetPayloadAttribute(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	opts := []Option{
		WithDatabase(beaconDB),
		WithStateGen(stategen.New(beaconDB)),
		WithProposerIdsCache(cache.NewProposerPayloadIDsCache()),
	}

	// Cache miss
	service, err := NewService(ctx, opts...)
	require.NoError(t, err)
	hasPayload, _, vId, err := service.getPayloadAttribute(ctx, nil, 0)
	require.NoError(t, err)
	require.Equal(t, false, hasPayload)
	require.Equal(t, types.ValidatorIndex(0), vId)

	// Cache hit, advance state, no fee recipient
	suggestedVid := types.ValidatorIndex(1)
	slot := types.Slot(1)
	service.cfg.ProposerSlotIndexCache.SetProposerAndPayloadIDs(slot, suggestedVid, [8]byte{})
	st, _ := util.DeterministicGenesisState(t, 1)
	hook := logTest.NewGlobal()
	hasPayload, attr, vId, err := service.getPayloadAttribute(ctx, st, slot)
	require.NoError(t, err)
	require.Equal(t, true, hasPayload)
	require.Equal(t, suggestedVid, vId)
	require.Equal(t, fieldparams.EthBurnAddressHex, common.BytesToAddress(attr.SuggestedFeeRecipient).String())
	require.LogsContain(t, hook, "Fee recipient not set. Using burn address")

	// Cache hit, advance state, has fee recipient
	suggestedAddr := common.HexToAddress("123")
	require.NoError(t, service.cfg.BeaconDB.SaveFeeRecipientsByValidatorIDs(ctx, []types.ValidatorIndex{suggestedVid}, []common.Address{suggestedAddr}))
	service.cfg.ProposerSlotIndexCache.SetProposerAndPayloadIDs(slot, suggestedVid, [8]byte{})
	hasPayload, attr, vId, err = service.getPayloadAttribute(ctx, st, slot)
	require.NoError(t, err)
	require.Equal(t, true, hasPayload)
	require.Equal(t, suggestedVid, vId)
	require.Equal(t, suggestedAddr, common.BytesToAddress(attr.SuggestedFeeRecipient))
}

func Test_UpdateLastValidatedCheckpoint(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.OverrideBeaconConfig(params.MainnetConfig())

	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	stateGen := stategen.New(beaconDB)
	fcs := protoarray.New(0, 0)
	opts := []Option{
		WithDatabase(beaconDB),
		WithStateGen(stateGen),
		WithForkChoiceStore(fcs),
	}
	service, err := NewService(ctx, opts...)
	require.NoError(t, err)
	genesisStateRoot := [32]byte{}
	genesisBlk := blocks.NewGenesisBlock(genesisStateRoot[:])
	wr, err := wrapper.WrappedSignedBeaconBlock(genesisBlk)
	require.NoError(t, err)
	assert.NoError(t, beaconDB.SaveBlock(ctx, wr))
	genesisRoot, err := genesisBlk.Block.HashTreeRoot()
	require.NoError(t, err)
	assert.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, genesisRoot))
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 0, genesisRoot, params.BeaconConfig().ZeroHash,
		0, 0, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	genesisSummary := &ethpb.StateSummary{
		Root: genesisStateRoot[:],
		Slot: 0,
	}
	require.NoError(t, beaconDB.SaveStateSummary(ctx, genesisSummary))

	// Get last validated checkpoint
	origCheckpoint, err := service.cfg.BeaconDB.LastValidatedCheckpoint(ctx)
	require.NoError(t, err)
	require.NoError(t, beaconDB.SaveLastValidatedCheckpoint(ctx, origCheckpoint))

	// Optimistic finalized checkpoint
	blk := util.NewBeaconBlock()
	blk.Block.Slot = 320
	blk.Block.ParentRoot = genesisRoot[:]
	wr, err = wrapper.WrappedSignedBeaconBlock(blk)
	require.NoError(t, err)
	require.NoError(t, beaconDB.SaveBlock(ctx, wr))
	opRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	opCheckpoint := &ethpb.Checkpoint{
		Root:  opRoot[:],
		Epoch: 10,
	}
	opStateSummary := &ethpb.StateSummary{
		Root: opRoot[:],
		Slot: 320,
	}
	require.NoError(t, beaconDB.SaveStateSummary(ctx, opStateSummary))
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 320, opRoot, genesisRoot,
		10, 10, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	assert.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, opRoot))
	require.NoError(t, service.updateFinalized(ctx, opCheckpoint))
	cp, err := service.cfg.BeaconDB.LastValidatedCheckpoint(ctx)
	require.NoError(t, err)
	require.DeepEqual(t, origCheckpoint.Root, cp.Root)
	require.Equal(t, origCheckpoint.Epoch, cp.Epoch)

	// Validated finalized checkpoint
	blk = util.NewBeaconBlock()
	blk.Block.Slot = 640
	blk.Block.ParentRoot = opRoot[:]
	wr, err = wrapper.WrappedSignedBeaconBlock(blk)
	require.NoError(t, err)
	require.NoError(t, beaconDB.SaveBlock(ctx, wr))
	validRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	validCheckpoint := &ethpb.Checkpoint{
		Root:  validRoot[:],
		Epoch: 20,
	}
	validSummary := &ethpb.StateSummary{
		Root: validRoot[:],
		Slot: 640,
	}
	require.NoError(t, beaconDB.SaveStateSummary(ctx, validSummary))
	require.NoError(t, fcs.InsertOptimisticBlock(ctx, 640, validRoot, params.BeaconConfig().ZeroHash,
		20, 20, params.BeaconConfig().ZeroHash[:], params.BeaconConfig().ZeroHash[:], nil))
	require.NoError(t, fcs.SetOptimisticToValid(ctx, validRoot))
	assert.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, validRoot))
	require.NoError(t, service.updateFinalized(ctx, validCheckpoint))
	cp, err = service.cfg.BeaconDB.LastValidatedCheckpoint(ctx)
	require.NoError(t, err)

	optimistic, err := service.IsOptimisticForRoot(ctx, validRoot)
	require.NoError(t, err)
	require.Equal(t, false, optimistic)
	require.DeepEqual(t, validCheckpoint.Root, cp.Root)
	require.Equal(t, validCheckpoint.Epoch, cp.Epoch)
}

func TestService_removeInvalidBlockAndState(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	opts := []Option{
		WithDatabase(beaconDB),
		WithStateGen(stategen.New(beaconDB)),
		WithForkChoiceStore(protoarray.New(0, 0)),
	}
	service, err := NewService(ctx, opts...)
	require.NoError(t, err)

	// Deleting unknown block should not error.
	require.NoError(t, service.removeInvalidBlockAndState(ctx, [][32]byte{{'a'}, {'b'}, {'c'}}))

	// Happy case
	b1 := util.NewBeaconBlock()
	b1.Block.Slot = 1
	blk1, err := wrapper.WrappedSignedBeaconBlock(b1)
	require.NoError(t, err)
	r1, err := blk1.Block().HashTreeRoot()
	require.NoError(t, err)
	st, _ := util.DeterministicGenesisStateBellatrix(t, 1)
	require.NoError(t, service.cfg.BeaconDB.SaveBlock(ctx, blk1))
	require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(ctx, &ethpb.StateSummary{
		Slot: 1,
		Root: r1[:],
	}))
	require.NoError(t, service.cfg.BeaconDB.SaveState(ctx, st, r1))

	b2 := util.NewBeaconBlock()
	b2.Block.Slot = 2
	blk2, err := wrapper.WrappedSignedBeaconBlock(b2)
	require.NoError(t, err)
	r2, err := blk2.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, service.cfg.BeaconDB.SaveBlock(ctx, blk2))
	require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(ctx, &ethpb.StateSummary{
		Slot: 2,
		Root: r2[:],
	}))
	require.NoError(t, service.cfg.BeaconDB.SaveState(ctx, st, r2))

	require.NoError(t, service.removeInvalidBlockAndState(ctx, [][32]byte{r1, r2}))

	require.Equal(t, false, service.hasBlock(ctx, r1))
	require.Equal(t, false, service.hasBlock(ctx, r2))
	require.Equal(t, false, service.cfg.BeaconDB.HasStateSummary(ctx, r1))
	require.Equal(t, false, service.cfg.BeaconDB.HasStateSummary(ctx, r2))
	has, err := service.cfg.StateGen.HasState(ctx, r1)
	require.NoError(t, err)
	require.Equal(t, false, has)
	has, err = service.cfg.StateGen.HasState(ctx, r2)
	require.NoError(t, err)
	require.Equal(t, false, has)
}
