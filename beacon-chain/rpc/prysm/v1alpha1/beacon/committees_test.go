package beacon

import (
	"context"
	"encoding/binary"
	"math"
	"testing"
	"time"

	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	dbTest "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	mockstategen "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen/mock"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
	prysmTime "gitlab.waterfall.network/waterfall/protocol/coordinator/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"google.golang.org/protobuf/proto"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestServer_ListBeaconCommittees_CurrentEpoch(t *testing.T) {
	db := dbTest.SetupDB(t)
	helpers.ClearCache()

	numValidators := 128
	ctx := context.Background()
	headState := setupActiveValidators(t, numValidators)

	offset := int64(headState.Slot().Mul(params.BeaconConfig().SecondsPerSlot))
	m := &mock.ChainService{
		Genesis: prysmTime.Now().Add(time.Duration(-1*offset) * time.Second),
	}
	bs := &Server{
		HeadFetcher:        m,
		GenesisTimeFetcher: m,
		StateGen:           stategen.New(db),
	}
	b := util.NewBeaconBlock()
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, wsb))
	gRoot, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, gRoot))
	require.NoError(t, db.SaveState(ctx, headState, gRoot))

	bs.ReplayerBuilder = mockstategen.NewMockReplayerBuilder(mockstategen.WithMockState(headState))

	activeIndices, err := helpers.ActiveValidatorIndices(ctx, headState, 0)
	require.NoError(t, err)
	attesterSeed, err := helpers.Seed(headState, 0, params.BeaconConfig().DomainBeaconAttester)
	require.NoError(t, err)
	committees, err := computeCommittees(context.Background(), 0, activeIndices, attesterSeed)
	require.NoError(t, err)

	wanted := &ethpb.BeaconCommittees{
		Epoch:                0,
		Committees:           committees.SlotToUint64(),
		ActiveValidatorCount: uint64(numValidators),
	}
	res, err := bs.ListBeaconCommittees(context.Background(), &ethpb.ListCommitteesRequest{
		QueryFilter: &ethpb.ListCommitteesRequest_Genesis{Genesis: true},
	})
	require.NoError(t, err)
	if !proto.Equal(res, wanted) {
		t.Errorf("Expected %v, received %v", wanted, res)
	}
}

func addDefaultReplayerBuilder(s *Server, h stategen.HistoryAccessor) {
	cc := &mockstategen.MockCanonicalChecker{Is: true, Err: nil}
	cs := &mockstategen.MockCurrentSlotter{Slot: math.MaxUint64 - 1}
	s.ReplayerBuilder = stategen.NewCanonicalHistory(h, cc, cs)
}

func TestServer_ListBeaconCommittees_PreviousEpoch(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.OverrideBeaconConfig(params.MainnetConfig())
	ctx := context.Background()

	db := dbTest.SetupDB(t)
	helpers.ClearCache()

	numValidators := 128
	headState := setupActiveValidators(t, numValidators)

	mixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := 0; i < len(mixes); i++ {
		mixes[i] = make([]byte, fieldparams.RootLength)
	}
	require.NoError(t, headState.SetRandaoMixes(mixes))
	require.NoError(t, headState.SetSlot(params.BeaconConfig().SlotsPerEpoch))

	b, err := wrapper.WrappedSignedBeaconBlock(util.NewBeaconBlock())
	require.NoError(t, wrapper.SetBlockSlot(b, headState.Slot()))
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, b))
	gRoot, err := b.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, headState, gRoot))

	offset := int64(headState.Slot().Mul(params.BeaconConfig().SecondsPerSlot))
	m := &mock.ChainService{
		State:   headState,
		Genesis: prysmTime.Now().Add(time.Duration(-1*offset) * time.Second),
	}
	bs := &Server{
		HeadFetcher:        m,
		GenesisTimeFetcher: m,
		StateGen:           stategen.New(db),
	}
	addDefaultReplayerBuilder(bs, db)

	activeIndices, err := helpers.ActiveValidatorIndices(ctx, headState, 1)
	require.NoError(t, err)
	attesterSeed, err := helpers.Seed(headState, 1, params.BeaconConfig().DomainBeaconAttester)
	require.NoError(t, err)
	startSlot, err := slots.EpochStart(1)
	require.NoError(t, err)
	wanted, err := computeCommittees(context.Background(), startSlot, activeIndices, attesterSeed)
	require.NoError(t, err)

	tests := []struct {
		req *ethpb.ListCommitteesRequest
		res *ethpb.BeaconCommittees
	}{
		{
			req: &ethpb.ListCommitteesRequest{
				QueryFilter: &ethpb.ListCommitteesRequest_Epoch{Epoch: 1},
			},
			res: &ethpb.BeaconCommittees{
				Epoch:                1,
				Committees:           wanted.SlotToUint64(),
				ActiveValidatorCount: uint64(numValidators),
			},
		},
	}
	helpers.ClearCache()
	for i, test := range tests {
		res, err := bs.ListBeaconCommittees(context.Background(), test.req)
		require.NoError(t, err)
		if !proto.Equal(res, test.res) {
			diff, _ := messagediff.PrettyDiff(res, test.res)
			t.Errorf("%d/ Diff between responses %s", i, diff)
		}
	}
}

func TestRetrieveCommitteesForRoot(t *testing.T) {

	db := dbTest.SetupDB(t)
	helpers.ClearCache()
	ctx := context.Background()

	numValidators := 128
	headState := setupActiveValidators(t, numValidators)

	offset := int64(headState.Slot().Mul(params.BeaconConfig().SecondsPerSlot))
	m := &mock.ChainService{
		Genesis: prysmTime.Now().Add(time.Duration(-1*offset) * time.Second),
	}
	bs := &Server{
		HeadFetcher:        m,
		GenesisTimeFetcher: m,
		StateGen:           stategen.New(db),
	}
	b := util.NewBeaconBlock()
	wsb, err := wrapper.WrappedSignedBeaconBlock(b)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, wsb))
	gRoot, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, gRoot))
	require.NoError(t, db.SaveState(ctx, headState, gRoot))
	stateSummary := &ethpb.StateSummary{
		Slot: 0,
		Root: gRoot[:],
	}
	require.NoError(t, db.SaveStateSummary(ctx, stateSummary))

	// Store the genesis seed.
	seed, err := helpers.Seed(headState, 0, params.BeaconConfig().DomainBeaconAttester)
	require.NoError(t, err)
	require.NoError(t, headState.SetSlot(params.BeaconConfig().SlotsPerEpoch*10))

	activeIndices, err := helpers.ActiveValidatorIndices(ctx, headState, 0)
	require.NoError(t, err)

	wanted, err := computeCommittees(context.Background(), 0, activeIndices, seed)
	require.NoError(t, err)
	committees, activeIndices, err := bs.retrieveCommitteesForRoot(context.Background(), gRoot[:])
	require.NoError(t, err)

	wantedRes := &ethpb.BeaconCommittees{
		Epoch:                0,
		Committees:           wanted.SlotToUint64(),
		ActiveValidatorCount: uint64(numValidators),
	}
	receivedRes := &ethpb.BeaconCommittees{
		Epoch:                0,
		Committees:           committees.SlotToUint64(),
		ActiveValidatorCount: uint64(len(activeIndices)),
	}
	assert.DeepEqual(t, wantedRes, receivedRes)
}

func setupActiveValidators(t *testing.T, count int) state.BeaconState {
	balances := make([]uint64, count)
	validators := make([]*ethpb.Validator, 0, count)
	for i := 0; i < count; i++ {
		pubKey := make([]byte, params.BeaconConfig().BLSPubkeyLength)
		binary.LittleEndian.PutUint64(pubKey, uint64(i))
		balances[i] = uint64(i)
		validators = append(validators, &ethpb.Validator{
			PublicKey:             pubKey,
			ActivationEpoch:       0,
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
			CreatorAddress:        make([]byte, 20),
			WithdrawalCredentials: make([]byte, 20),
			ActivationHash:        make([]byte, 32),
			ExitHash:              make([]byte, 32),
			WithdrawalOps:         make([]*ethpb.WithdrawalOp, 0),
		})
	}
	s, err := util.NewBeaconState()
	require.NoError(t, err)
	if err := s.SetValidators(validators); err != nil {
		t.Error(err)
		return nil
	}
	if err := s.SetBalances(balances); err != nil {
		t.Error(err)
		return nil
	}
	return s
}
