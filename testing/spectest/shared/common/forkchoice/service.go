package forkchoice

import (
	"context"
	"math/big"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain"
	mock "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/blockchain/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache/depositcache"
	coreTime "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	testDB "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/testing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/forkchoice/protoarray"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/engine/v1"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

func startChainService(t *testing.T, st state.BeaconState, block block.SignedBeaconBlock, engineMock *engineMock) *blockchain.Service {
	db := testDB.SetupDB(t)
	ctx := context.Background()
	require.NoError(t, db.SaveBlock(ctx, block))
	r, err := block.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, r))
	require.NoError(t, db.SaveState(ctx, st, r))
	cp := &ethpb.Checkpoint{
		Epoch: coreTime.CurrentEpoch(st),
		Root:  r[:],
	}

	require.NoError(t, db.SaveJustifiedCheckpoint(ctx, cp))
	require.NoError(t, db.SaveFinalizedCheckpoint(ctx, cp))
	attPool, err := attestations.NewService(ctx, &attestations.Config{
		Pool: attestations.NewPool(),
	})
	require.NoError(t, err)

	depositCache, err := depositcache.New()
	require.NoError(t, err)

	opts := append([]blockchain.Option{},
		blockchain.WithExecutionEngineCaller(engineMock),
		blockchain.WithFinalizedStateAtStartUp(st),
		blockchain.WithDatabase(db),
		blockchain.WithAttestationService(attPool),
		blockchain.WithForkChoiceStore(protoarray.New(0, 0, params.BeaconConfig().ZeroHash)),
		blockchain.WithStateGen(stategen.New(db)),
		blockchain.WithStateNotifier(&mock.MockStateNotifier{}),
		blockchain.WithAttestationPool(attestations.NewPool()),
		blockchain.WithDepositCache(depositCache),
		blockchain.WithProposerIdsCache(cache.NewProposerPayloadIDsCache()),
	)
	service, err := blockchain.NewService(context.Background(), opts...)
	require.NoError(t, err)
	require.NoError(t, service.StartFromSavedState(st))
	return service
}

type engineMock struct {
	powBlocks map[[32]byte]*ethpb.PowBlock
}

func (m *engineMock) ExecutionDagValidateSpines(ctx context.Context, params gwatCommon.HashArray) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) ExecutionDagFinalize(ctx context.Context, finParams *gwatTypes.FinalizationParams) (*gwatTypes.FinalizationResult, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) ExecutionDagCoordinatedState(ctx context.Context) (*gwatTypes.FinalizationResult, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) ExecutionDagGetCandidates(ctx context.Context, slot types.Slot) (gwatCommon.HashArray, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) GetHeaderByHash(ctx context.Context, hash gwatCommon.Hash) (*gwatTypes.Header, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) GetHeaderByNumber(ctx context.Context, nr *big.Int) (*gwatTypes.Header, error) {
	//TODO implement me
	panic("implement me")
}

func (m *engineMock) GetPayload(context.Context, [8]byte) (*pb.ExecutionPayload, error) {
	return nil, nil
}
func (m *engineMock) ForkchoiceUpdated(context.Context, *pb.ForkchoiceState, *pb.PayloadAttributes) (*pb.PayloadIDBytes, []byte, error) {
	return nil, nil, nil
}
func (m *engineMock) NewPayload(context.Context, *pb.ExecutionPayload) ([]byte, error) {
	return nil, nil
}

func (m *engineMock) LatestExecutionBlock(context.Context) (*pb.ExecutionBlock, error) {
	return nil, nil
}

func (m *engineMock) ExchangeTransitionConfiguration(context.Context, *pb.TransitionConfiguration) error {
	return nil
}

func (m *engineMock) ExecutionBlockByHash(_ context.Context, hash gwatCommon.Hash) (*pb.ExecutionBlock, error) {
	b, ok := m.powBlocks[bytesutil.ToBytes32(hash.Bytes())]
	if !ok {
		return nil, nil
	}

	td := new(big.Int).SetBytes(bytesutil.ReverseByteOrder(b.TotalDifficulty))
	tdHex := hexutil.EncodeBig(td)
	return &pb.ExecutionBlock{
		ParentHash:      b.ParentHash,
		TotalDifficulty: tdHex,
		Hash:            b.BlockHash,
	}, nil
}

func (m *engineMock) GetTerminalBlockHash(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}
