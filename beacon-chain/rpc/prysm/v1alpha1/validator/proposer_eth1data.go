package validator

import (
	"context"
	"math/big"

	fastssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/blocks"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/crypto/hash"
	"github.com/waterfall-foundation/coordinator/crypto/rand"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/time/slots"
	"github.com/waterfall-foundation/gwat/common"
	gwatCommon "github.com/waterfall-foundation/gwat/common"
)

// eth1DataMajorityVote determines the appropriate eth1data for a block proposal using
// an algorithm called Voting with the Majority. The algorithm works as follows:
//  - Determine the timestamp for the start slot for the eth1 voting period.
//  - Determine the earliest and latest timestamps that a valid block can have.
//  - Determine the first block not before the earliest timestamp. This block is the lower bound.
//  - Determine the last block not after the latest timestamp. This block is the upper bound.
//  - If the last block is too early, use current eth1data from the beacon state.
//  - Filter out votes on unknown blocks and blocks which are outside of the range determined by the lower and upper bounds.
//  - If no blocks are left after filtering votes, use eth1data from the latest valid block.
//  - Otherwise:
//    - Determine the vote with the highest count. Prefer the vote with the highest eth1 block height in the event of a tie.
//    - This vote's block is the eth1 block to use for the block proposal.
func (vs *Server) eth1DataMajorityVote(ctx context.Context, beaconState state.BeaconState) (*ethpb.Eth1Data, error) {
	ctx, cancel := context.WithTimeout(ctx, eth1dataTimeout)
	defer cancel()

	slot := beaconState.Slot()
	votingPeriodStartTime := vs.slotStartTime(slot)

	// if inerOp true
	if vs.MockEth1Votes {
		return vs.mockETH1DataVote(ctx, slot)
	}
	if !vs.Eth1InfoFetcher.IsConnectedToETH1() {
		//return vs.randomETH1DataVote
		prevEth1Data := vs.HeadFetcher.HeadETH1Data()
		candidates := vs.HeadFetcher.GetCacheCandidates()
		finalization := vs.HeadFetcher.GetCacheFinalization()
		return &ethpb.Eth1Data{
			Candidates:   candidates.ToBytes(),
			Finalization: finalization.ToBytes(),
			BlockHash:    prevEth1Data.GetBlockHash(),
			DepositCount: prevEth1Data.GetDepositCount(),
			DepositRoot:  prevEth1Data.GetDepositRoot(),
		}, nil
	}
	eth1DataNotification = false

	eth1FollowDistance := params.BeaconConfig().Eth1FollowDistance
	earliestValidTime := votingPeriodStartTime - 2*params.BeaconConfig().SecondsPerETH1Block*eth1FollowDistance
	latestValidTime := votingPeriodStartTime - params.BeaconConfig().SecondsPerETH1Block*eth1FollowDistance

	//if !features.Get().EnableGetBlockOptimizations {
	//	_, err := vs.Eth1BlockFetcher.BlockByTimestamp(ctx, earliestValidTime)
	//	if err != nil {
	//		log.WithError(err).Error("Could not get last block by earliest valid time")
	//		return vs.randomETH1DataVote(ctx)
	//	}
	//}

	lastBlockByLatestValidTime, err := vs.Eth1BlockFetcher.BlockByTimestamp(ctx, latestValidTime)
	if err != nil {
		log.WithError(err).Error("Could not get last block by latest valid time")
		//return vs.randomETH1DataVote(ctx)
		prevEth1Data := vs.HeadFetcher.HeadETH1Data()
		candidates := vs.HeadFetcher.GetCacheCandidates()
		finalization := vs.HeadFetcher.GetCacheFinalization()
		return &ethpb.Eth1Data{
			Candidates:   candidates.ToBytes(),
			Finalization: finalization.ToBytes(),
			BlockHash:    prevEth1Data.GetBlockHash(),
			DepositCount: prevEth1Data.GetDepositCount(),
			DepositRoot:  prevEth1Data.GetDepositRoot(),
		}, nil
	}
	if lastBlockByLatestValidTime.Time < earliestValidTime {
		prevEth1Data := vs.HeadFetcher.HeadETH1Data()
		candidates := vs.HeadFetcher.GetCacheCandidates()
		finalization := vs.HeadFetcher.GetCacheFinalization()
		return &ethpb.Eth1Data{
			Candidates:   candidates.ToBytes(),
			Finalization: finalization.ToBytes(),
			BlockHash:    prevEth1Data.GetBlockHash(),
			DepositCount: prevEth1Data.GetDepositCount(),
			DepositRoot:  prevEth1Data.GetDepositRoot(),
		}, nil
	}

	lastBlockDepositCount, lastBlockDepositRoot := vs.DepositFetcher.DepositsNumberAndRootAtHeight(ctx, lastBlockByLatestValidTime.Number)
	//if lastBlockDepositCount == 0 {
	//	prevEth1Data := vs.HeadFetcher.HeadETH1Data()
	//	candidates := vs.HeadFetcher.GetCacheCandidates()
	//	finalization := vs.HeadFetcher.GetCacheFinalization()
	//	return &ethpb.Eth1Data{
	//		Candidates:   candidates.ToBytes(),
	//		Finalization: finalization.ToBytes(),
	//		BlockHash:    prevEth1Data.GetBlockHash(),
	//		DepositCount: prevEth1Data.GetDepositCount(),
	//		DepositRoot:  prevEth1Data.GetDepositRoot(),
	//	}, nil
	//}

	if lastBlockDepositCount >= vs.HeadFetcher.HeadETH1Data().DepositCount {
		hash, err := vs.Eth1BlockFetcher.BlockHashByHeight(ctx, lastBlockByLatestValidTime.Number)
		if err != nil {
			log.WithError(err).Error("Could not get hash of last block by latest valid time")
			//return vs.randomETH1DataVote(ctx)
			prevEth1Data := vs.HeadFetcher.HeadETH1Data()
			candidates := vs.HeadFetcher.GetCacheCandidates()
			finalization := vs.HeadFetcher.GetCacheFinalization()
			return &ethpb.Eth1Data{
				Candidates:   candidates.ToBytes(),
				Finalization: finalization.ToBytes(),
				BlockHash:    prevEth1Data.GetBlockHash(),
				DepositCount: prevEth1Data.GetDepositCount(),
				DepositRoot:  prevEth1Data.GetDepositRoot(),
			}, nil
		}
		candidates := vs.HeadFetcher.GetCacheCandidates()
		finalization := vs.HeadFetcher.GetCacheFinalization()
		return &ethpb.Eth1Data{
			Candidates:   candidates.ToBytes(),
			Finalization: finalization.ToBytes(),
			BlockHash:    hash.Bytes(),
			DepositCount: lastBlockDepositCount,
			DepositRoot:  lastBlockDepositRoot[:],
		}, nil
	}

	prevEth1Data := vs.HeadFetcher.HeadETH1Data()
	candidates := vs.HeadFetcher.GetCacheCandidates()
	finalization := vs.HeadFetcher.GetCacheFinalization()
	return &ethpb.Eth1Data{
		Candidates:   candidates.ToBytes(),
		Finalization: finalization.ToBytes(),
		BlockHash:    prevEth1Data.GetBlockHash(),
		DepositCount: prevEth1Data.GetDepositCount(),
		DepositRoot:  prevEth1Data.GetDepositRoot(),
	}, nil
}

func (vs *Server) slotStartTime(slot types.Slot) uint64 {
	startTime, _ := vs.Eth1InfoFetcher.Eth2GenesisPowchainInfo()
	return slots.VotingPeriodStartTime(startTime, slot)
}

// canonicalEth1Data determines the canonical eth1data and eth1 block height to use for determining deposits.
func (vs *Server) canonicalEth1Data(
	ctx context.Context,
	beaconState state.BeaconState,
	currentVote *ethpb.Eth1Data) (*ethpb.Eth1Data, *big.Int, error) {

	var eth1BlockHash [32]byte

	// Add in current vote, to get accurate vote tally
	if err := beaconState.AppendEth1DataVotes(currentVote); err != nil {
		return nil, nil, errors.Wrap(err, "could not append eth1 data votes to state")
	}
	hasSupport, err := blocks.Eth1DataHasEnoughSupport(beaconState, currentVote)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not determine if current eth1data vote has enough support")
	}
	var canonicalEth1Data *ethpb.Eth1Data
	if hasSupport {
		canonicalEth1Data = currentVote
		eth1BlockHash = bytesutil.ToBytes32(currentVote.BlockHash)
	} else {
		canonicalEth1Data = beaconState.Eth1Data()
		eth1BlockHash = bytesutil.ToBytes32(beaconState.Eth1Data().BlockHash)
	}
	_, canonicalEth1DataHeight, err := vs.Eth1BlockFetcher.BlockExists(ctx, eth1BlockHash)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not fetch eth1data height")
	}
	return canonicalEth1Data, canonicalEth1DataHeight, nil
}

func (vs *Server) mockETH1DataVote(ctx context.Context, slot types.Slot) (*ethpb.Eth1Data, error) {
	if !eth1DataNotification {
		log.Warn("Beacon Node is no longer connected to an ETH1 chain, so ETH1 data votes are now mocked.")
		eth1DataNotification = true
	}
	// If a mock eth1 data votes is specified, we use the following for the
	// eth1data we provide to every proposer based on https://github.com/ethereum/eth2.0-pm/issues/62:
	//
	// slot_in_voting_period = current_slot % SLOTS_PER_ETH1_VOTING_PERIOD
	// Eth1Data(
	//   DepositRoot = hash(current_epoch + slot_in_voting_period),
	//   DepositCount = state.eth1_deposit_index,
	//   BlockHash = hash(hash(current_epoch + slot_in_voting_period)),
	//   Candidates = ffinalizer.NrHashMap{ slot: hash(hash(current_epoch + slot_in_voting_period)) }
	// )
	slotInVotingPeriod := slot.ModSlot(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerEth1VotingPeriod)))
	headState, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, err
	}
	var enc []byte
	enc = fastssz.MarshalUint64(enc, uint64(slots.ToEpoch(slot))+uint64(slotInVotingPeriod))
	depRoot := hash.Hash(enc)
	blockHash := hash.Hash(depRoot[:])

	finHash := &common.Hash{}
	finHash.SetBytes(depRoot[:])
	candidates := gwatCommon.HashArray{*finHash}

	return &ethpb.Eth1Data{
		DepositRoot:  depRoot[:],
		DepositCount: headState.Eth1DepositIndex(),
		BlockHash:    blockHash[:],
		Candidates:   candidates.ToBytes(),
	}, nil
}

func (vs *Server) randomETH1DataVote(ctx context.Context) (*ethpb.Eth1Data, error) {
	if !eth1DataNotification {
		log.Warn("Beacon Node is no longer connected to an ETH1 chain, so ETH1 data votes are now random.")
		eth1DataNotification = true
	}
	headState, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, err
	}

	// set random roots and block hashes to prevent a majority from being
	// built if the eth1 node is offline
	randGen := rand.NewGenerator()
	depRoot := hash.Hash(bytesutil.Bytes32(randGen.Uint64()))
	blockHash := hash.Hash(bytesutil.Bytes32(randGen.Uint64()))

	finHash := &common.Hash{}
	finHash.SetBytes(blockHash[:])
	candidates := gwatCommon.HashArray{*finHash}

	return &ethpb.Eth1Data{
		DepositRoot:  depRoot[:],
		DepositCount: headState.Eth1DepositIndex(),
		BlockHash:    blockHash[:],
		Candidates:   candidates.ToBytes(),
	}, nil
}
