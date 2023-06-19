package validator

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
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

	candidates, err := vs.Eth1BlockFetcher.ExecutionDagGetCandidates(ctx, req.GetSlot())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve prevote candidates: %v", err)
	}
	if candidates == nil {
		return nil, status.Errorf(codes.DataLoss, "A request was in progress and resolved to nil")
	}

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
