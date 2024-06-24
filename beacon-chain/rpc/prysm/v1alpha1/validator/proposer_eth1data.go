package validator

import (
	"context"
	"fmt"
	"math/big"

	fastssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/hash"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// eth1DataMajorityVote determines the appropriate eth1data for a block proposal using
// an algorithm called Voting with the Majority. The algorithm works as follows:
//   - Determine the timestamp for the start slot for the eth1 voting period.
//   - Determine the earliest and latest timestamps that a valid block can have.
//   - Determine the first block not before the earliest timestamp. This block is the lower bound.
//   - Determine the last block not after the latest timestamp. This block is the upper bound.
//   - If the last block is too early, use current eth1data from the beacon state.
//   - Filter out votes on unknown blocks and blocks which are outside of the range determined by the lower and upper bounds.
//   - If no blocks are left after filtering votes, use eth1data from the latest valid block.
//   - Otherwise:
//   - Determine the vote with the highest count. Prefer the vote with the highest eth1 block height in the event of a tie.
//   - This vote's block is the eth1 block to use for the block proposal.
func (vs *Server) eth1DataMajorityVote(ctx context.Context, beaconState state.BeaconState) (*ethpb.Eth1Data, error) {
	ctx, cancel := context.WithTimeout(ctx, eth1dataTimeout)
	defer cancel()

	prevEth1Data := beaconState.Eth1Data()
	if !vs.Eth1InfoFetcher.IsConnectedToETH1() || beaconState.Slot() < params.BeaconConfig().SlotsPerEpoch*types.Slot(params.BeaconConfig().EpochsPerEth1VotingPeriod) {
		return &ethpb.Eth1Data{
			BlockHash:    prevEth1Data.GetBlockHash(),
			DepositCount: prevEth1Data.GetDepositCount(),
			DepositRoot:  prevEth1Data.GetDepositRoot(),
		}, nil
	}
	eth1DataNotification = false

	prevEth1BlockHash := gwatCommon.BytesToHash(prevEth1Data.GetBlockHash())
	prevExists, prevEth1BlockNr, err := vs.Eth1BlockFetcher.BlockExists(ctx, prevEth1BlockHash)
	if !prevExists || err != nil {
		log.WithError(err).Warn("eth1DataMajorityVote: eth1 block not found")
		return nil, errors.Wrap(err, "eth1 block not found")
	}

	cpSpine := helpers.GetBaseSpine(beaconState)
	if params.BeaconConfig().IsFinEth1ForkSlot(beaconState.Slot()) {
		log.WithFields(logrus.Fields{
			"slot":              beaconState.Slot(),
			"forkSlot":          params.BeaconConfig().FinEth1ForkSlot,
			"eth1.DepositCount": prevEth1Data.DepositCount,
			"eth1.BlockHash":    fmt.Sprintf("%#x", prevEth1Data.BlockHash),
			"eth1.DepositRoot":  fmt.Sprintf("%#x", prevEth1Data.DepositRoot),
			"prevEth1BlockNr":   prevEth1BlockNr.String(),
		}).Info("eth1DataMajorityVote: finalized eth1 fork active")

		fCpRoot := bytesutil.ToBytes32(beaconState.FinalizedCheckpoint().Root)
		if fCpRoot == params.BeaconConfig().ZeroHash {
			log.Warn("eth1DataMajorityVote: finalized root empty")
			return &ethpb.Eth1Data{
				BlockHash:    prevEth1Data.GetBlockHash(),
				DepositCount: prevEth1Data.GetDepositCount(),
				DepositRoot:  prevEth1Data.GetDepositRoot(),
			}, nil
		}

		finSt, err := vs.StateGen.StateByRoot(ctx, fCpRoot)
		if err != nil {
			log.WithError(err).Warn("eth1DataMajorityVote: get finalized state failed")
			return &ethpb.Eth1Data{
				BlockHash:    prevEth1Data.GetBlockHash(),
				DepositCount: prevEth1Data.GetDepositCount(),
				DepositRoot:  prevEth1Data.GetDepositRoot(),
			}, nil
		}
		cpSpine = helpers.GetBaseSpine(finSt)
	}

	cpSpineExists, cpSpineNum, err := vs.Eth1BlockFetcher.BlockExists(ctx, cpSpine)
	if !cpSpineExists || err != nil {
		log.WithError(err).Warn("eth1DataMajorityVote: could not retrieve checkpoint terminal spine")
		return nil, errors.Wrap(err, "eth1DataMajorityVote: could not retrieve checkpoint terminal spine")
	}

	if cpSpineNum.Cmp(prevEth1BlockNr) < 0 || cpSpineNum.Cmp(prevEth1BlockNr) == 0 {
		log.WithFields(logrus.Fields{
			"slot":              beaconState.Slot(),
			"forkSlot":          params.BeaconConfig().FinEth1ForkSlot,
			"eth1.DepositCount": prevEth1Data.DepositCount,
			"eth1.BlockHash":    fmt.Sprintf("%#x", prevEth1Data.BlockHash),
			"eth1.DepositRoot":  fmt.Sprintf("%#x", prevEth1Data.DepositRoot),
			"prevEth1BlockNr":   prevEth1BlockNr.String(),
			"cpSpineNum":        cpSpineNum.String(),
		}).Warn("eth1DataMajorityVote: low next block number: using current")
		return &ethpb.Eth1Data{
			BlockHash:    prevEth1Data.GetBlockHash(),
			DepositCount: prevEth1Data.GetDepositCount(),
			DepositRoot:  prevEth1Data.GetDepositRoot(),
		}, nil
	}

	cpDepositCount, cpDepositRoot := vs.DepositFetcher.DepositsNumberAndRootAtHeight(ctx, cpSpineNum)

	if cpDepositCount >= vs.HeadFetcher.HeadETH1Data().DepositCount && cpDepositCount > 0 {

		log.WithFields(logrus.Fields{
			" BlockHash":                  fmt.Sprintf("%#x", cpSpine.Bytes()),
			"cpDepositRoot":               fmt.Sprintf("%#x", cpDepositRoot),
			"cpDepositCount":              cpDepositCount,
			"HeadETH1Data().DepositCount": vs.HeadFetcher.HeadETH1Data().DepositCount,
			"condition":                   cpDepositCount >= vs.HeadFetcher.HeadETH1Data().DepositCount,
			"prevEth1BlockNr":             prevEth1BlockNr.String(),
			"cpSpineNum":                  cpSpineNum.String(),
		}).Info("eth1DataMajorityVote: update deposit eth1 data")

		return &ethpb.Eth1Data{
			BlockHash:    cpSpine.Bytes(),
			DepositCount: cpDepositCount,
			DepositRoot:  cpDepositRoot[:],
		}, nil
	}

	return &ethpb.Eth1Data{
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
		return nil, nil, errors.Wrapf(err, "could not fetch eth1data by hash=%#x hasSupport=%v", eth1BlockHash, hasSupport)
	}
	return canonicalEth1Data, canonicalEth1DataHeight, nil
}

func (vs *Server) mockETH1DataVote(ctx context.Context, slot types.Slot) (*ethpb.Eth1Data, error) {
	if !eth1DataNotification {
		log.Warn("Beacon Node is no longer connected to an ETH1 chain, so ETH1 data votes are now mocked.")
		eth1DataNotification = true
	}
	slotInVotingPeriod := slot.ModSlot(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerEth1VotingPeriod)))
	headState, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, err
	}
	var enc []byte
	enc = fastssz.MarshalUint64(enc, uint64(slots.ToEpoch(slot))+uint64(slotInVotingPeriod))
	depRoot := hash.Hash(enc)
	blockHash := hash.Hash(depRoot[:])

	finHash := &gwatCommon.Hash{}
	finHash.SetBytes(depRoot[:])
	candidates := gwatCommon.HashArray{*finHash}

	return &ethpb.Eth1Data{
		DepositRoot:  depRoot[:],
		DepositCount: headState.Eth1DepositIndex(),
		BlockHash:    blockHash[:],
		Candidates:   candidates.ToBytes(),
	}, nil
}
