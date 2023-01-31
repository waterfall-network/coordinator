package testing

import (
	"context"
	"math/big"

	"github.com/holiman/uint256"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/engine/v1"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
)

// EngineClient --
type EngineClient struct {
	NewPayloadResp          []byte
	PayloadIDBytes          *pb.PayloadIDBytes
	ForkChoiceUpdatedResp   []byte
	ExecutionPayload        *pb.ExecutionPayload
	ExecutionBlock          *pb.ExecutionBlock
	Err                     error
	ErrLatestExecBlock      error
	ErrExecBlockByHash      error
	ErrForkchoiceUpdated    error
	ErrNewPayload           error
	BlockByHashMap          map[[32]byte]*pb.ExecutionBlock
	TerminalBlockHash       []byte
	TerminalBlockHashExists bool
}

func (e *EngineClient) ExecutionDagSync(ctx context.Context, syncParams *gwatTypes.ConsensusInfo) (gwatCommon.HashArray, error) {
	panic("implement me")
}

func (e *EngineClient) ExecutionDagGetCandidates(ctx context.Context, slot types.Slot) (gwatCommon.HashArray, error) {
	panic("implement me")
}

func (e *EngineClient) ExecutionDagFinalize(ctx context.Context, spines gwatCommon.HashArray, baseSpine *gwatCommon.Hash) (*gwatCommon.Hash, error) {
	panic("implement me")
}

func (e *EngineClient) ExecutionDagHeadSyncReady(ctx context.Context, params *gwatTypes.ConsensusInfo) (bool, error) {
	panic("implement me")
}

func (e *EngineClient) ExecutionDagHeadSync(ctx context.Context, params []gwatTypes.ConsensusInfo) (bool, error) {
	panic("implement me")
}

func (e *EngineClient) ExecutionDagValidateSpines(ctx context.Context, params gwatCommon.HashArray) (bool, error) {
	panic("implement me")
}

func (e *EngineClient) GetHeaderByHash(ctx context.Context, hash gwatCommon.Hash) (*gwatTypes.Header, error) {
	panic("implement me")
}

func (e *EngineClient) GetHeaderByNumber(ctx context.Context, nr *big.Int) (*gwatTypes.Header, error) {
	panic("implement me")
}

// NewPayload --
func (e *EngineClient) NewPayload(_ context.Context, _ *pb.ExecutionPayload) ([]byte, error) {
	return e.NewPayloadResp, e.ErrNewPayload
}

// ForkchoiceUpdated --
func (e *EngineClient) ForkchoiceUpdated(
	_ context.Context, _ *pb.ForkchoiceState, _ *pb.PayloadAttributes,
) (*pb.PayloadIDBytes, []byte, error) {
	return e.PayloadIDBytes, e.ForkChoiceUpdatedResp, e.ErrForkchoiceUpdated
}

// GetPayload --
func (e *EngineClient) GetPayload(_ context.Context, _ [8]byte) (*pb.ExecutionPayload, error) {
	return e.ExecutionPayload, nil
}

// ExchangeTransitionConfiguration --
func (e *EngineClient) ExchangeTransitionConfiguration(_ context.Context, _ *pb.TransitionConfiguration) error {
	return e.Err
}

// LatestExecutionBlock --
func (e *EngineClient) LatestExecutionBlock(_ context.Context) (*pb.ExecutionBlock, error) {
	return e.ExecutionBlock, e.ErrLatestExecBlock
}

// ExecutionBlockByHash --
func (e *EngineClient) ExecutionBlockByHash(_ context.Context, h common.Hash) (*pb.ExecutionBlock, error) {
	b, ok := e.BlockByHashMap[h]
	if !ok {
		return nil, errors.New("block not found")
	}
	return b, e.ErrExecBlockByHash
}

// GetTerminalBlockHash --
func (e *EngineClient) GetTerminalBlockHash(ctx context.Context) ([]byte, bool, error) {
	ttd := new(big.Int)
	ttd.SetString(params.BeaconConfig().TerminalTotalDifficulty, 10)
	terminalTotalDifficulty, overflows := uint256.FromBig(ttd)
	if overflows {
		return nil, false, errors.New("could not convert terminal total difficulty to uint256")
	}
	blk, err := e.LatestExecutionBlock(ctx)
	if err != nil {
		return nil, false, errors.Wrap(err, "could not get latest execution block")
	}
	if blk == nil {
		return nil, false, errors.New("latest execution block is nil")
	}

	for {
		b, err := hexutil.DecodeBig(blk.TotalDifficulty)
		if err != nil {
			return nil, false, errors.Wrap(err, "could not convert total difficulty to uint256")
		}
		currentTotalDifficulty, _ := uint256.FromBig(b)
		blockReachedTTD := currentTotalDifficulty.Cmp(terminalTotalDifficulty) >= 0

		parentHash := bytesutil.ToBytes32(blk.ParentHash)
		if len(blk.ParentHash) == 0 || parentHash == params.BeaconConfig().ZeroHash {
			return nil, false, nil
		}
		parentBlk, err := e.ExecutionBlockByHash(ctx, parentHash)
		if err != nil {
			return nil, false, errors.Wrap(err, "could not get parent execution block")
		}
		if blockReachedTTD {
			b, err := hexutil.DecodeBig(parentBlk.TotalDifficulty)
			if err != nil {
				return nil, false, errors.Wrap(err, "could not convert total difficulty to uint256")
			}
			parentTotalDifficulty, _ := uint256.FromBig(b)
			parentReachedTTD := parentTotalDifficulty.Cmp(terminalTotalDifficulty) >= 0
			if !parentReachedTTD {
				return blk.Hash, true, nil
			}
		} else {
			return nil, false, nil
		}
		blk = parentBlk
	}
}
