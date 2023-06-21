package validator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetPrevoteData requests that the beacon node produce an prevote data object,
// which the validator acting as an attester will then sign.
func (vs *Server) GetPrevoteData(ctx context.Context, req *ethpb.PreVoteRequest) (*ethpb.PreVoteData, error) {
	ctx, span := trace.StartSpan(ctx, "AttesterServer.RequestPrevote")
	defer span.End()
	span.AddAttributes(
		trace.Int64Attribute("slot", int64(req.Slot)),
		trace.Int64Attribute("committeeIndex", int64(req.CommitteeIndex)),
	)

	if vs.SyncChecker.Syncing() {
		return nil, status.Errorf(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}
	if vs.HeadFetcher.IsGwatSynchronizing() {
		log.WithError(fmt.Errorf("GWAT synchronization process is running, not ready to respond")).WithFields(logrus.Fields{
			"Syncing": vs.HeadFetcher.IsGwatSynchronizing(),
		}).Warn("GetPrevoteData: Proposing skipped (synchronizing)")
		return nil, status.Errorf(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	if err := vs.optimisticStatus(ctx); err != nil {
		return nil, err
	}

	// result is not depending on CommitteeIndex
	req.CommitteeIndex = 0
	res, err := vs.PrevoteCache.Get(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve data from prevote cache: %v", err)
	}
	if res != nil {
		res.Index = req.CommitteeIndex
		return res, nil
	}

	if err := vs.PrevoteCache.MarkInProgress(req); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			res, err := vs.PrevoteCache.Get(ctx, req)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not retrieve data from prevote cache: %v", err)
			}
			if res == nil {
				return nil, status.Error(codes.DataLoss, "A request was in progress and resolved to nil")
			}
			res.Index = req.CommitteeIndex
			return res, nil
		}
		return nil, status.Errorf(codes.Internal, "Could not mark prevote as in-progress: %v", err)
	}
	defer func() {
		if err := vs.PrevoteCache.MarkNotInProgress(req); err != nil {
			log.WithError(err).Error("Could not mark cache not in progress")
		}
	}()

	// todo check
	// calculate the parent block by optimistic spine.
	currHead, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		log.WithError(err).Error("Get prevote data: Could not retrieve head state")
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}

	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(err).Error("Get prevote data: retrieving of genesis root")
			return nil, status.Errorf(codes.Internal, "Could not retrieve of genesis root: %v", err)
		}
	}

	cpSt, err := vs.StateGen.StateByRoot(ctx, jCpRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"jCpRoot": fmt.Sprintf("%#x", jCpRoot),
		}).Error("Could not retrieve of cp state while prevote")
		return nil, status.Errorf(codes.Internal, "Could not retrieve of cp state: %v", err)
	}

	//request optimistic spine
	baseSpine := helpers.GetTerminalFinalizedSpine(cpSt)
	optSpines, err := vs.ExecutionEngineCaller.ExecutionDagGetOptimisticSpines(ctx, baseSpine)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat candidates: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"baseSpine": baseSpine,
		}).Error("Get attestation data: Could not retrieve of gwat optimistic spines while prevote")
		return nil, status.Errorf(codes.Internal, "Could not retrieve of gwat optimistic spines: %v", err)
	}
	candidates := make(gwatCommon.HashArray, 0, len(optSpines))
	for _, spines := range optSpines {
		if len(spines) != 0 {
			candidates = append(candidates, spines[0])
		}
	}
	// todo check ^^^^^^^^^^

	res = &ethpb.PreVoteData{
		Slot:       req.Slot,
		Index:      req.CommitteeIndex,
		Candidates: candidates.ToBytes(),
	}

	if err := vs.PrevoteCache.Put(ctx, req, res); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not store prevote data in cache: %v", err)
	}
	return res, nil
}

func (vs *Server) ProposePrevote(ctx context.Context, pv *ethpb.PreVote) (*ethpb.PrevoteResponse, error) {
	ctx, span := trace.StartSpan(ctx, "PrevoteServer.ProposePrevote")
	defer span.End()

	if _, err := bls.SignatureFromBytes(pv.Signature); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Incorrect prevote signature")
	}

	root, err := pv.Data.HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not tree hash prevote: %v", err)
	}

	// Determine subnet to broadcast prevote to
	wantedEpoch := slots.ToEpoch(pv.Data.Slot)
	vals, err := vs.HeadFetcher.HeadValidatorsIndices(ctx, wantedEpoch)
	if err != nil {
		return nil, err
	}
	subnet := helpers.ComputeSubnetFromCommitteeAndSlot(uint64(len(vals)), pv.Data.Index, pv.Data.Slot)

	// Broadcast the new prevote to the network.
	if err := vs.P2P.BroadcastPrevoting(ctx, subnet, pv); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not broadcast prevote: %v", err)
	}

	return &ethpb.PrevoteResponse{
		PrevoteDataRoot: root[:],
	}, nil
}
