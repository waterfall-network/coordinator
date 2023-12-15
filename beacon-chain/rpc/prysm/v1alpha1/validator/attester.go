package validator

import (
	"context"
	"errors"
	"fmt"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/cache"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAttestationData requests that the beacon node produce an attestation data object,
// which the validator acting as an attester will then sign.
func (vs *Server) GetAttestationData(ctx context.Context, req *ethpb.AttestationDataRequest) (*ethpb.AttestationData, error) {
	ctx, span := trace.StartSpan(ctx, "AttesterServer.RequestAttestation")
	defer span.End()
	span.AddAttributes(
		trace.Int64Attribute("slot", int64(req.Slot)),
		trace.Int64Attribute("committeeIndex", int64(req.CommitteeIndex)),
	)

	if vs.SyncChecker.Syncing() {
		return nil, status.Errorf(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	//if vs.HeadFetcher.IsGwatSynchronizing() {
	//	log.WithError(fmt.Errorf("GWAT synchronization process is running, not ready to respond")).WithFields(logrus.Fields{
	//		"Syncing": vs.HeadFetcher.IsGwatSynchronizing(),
	//	}).Warn("GetAttestationData: Proposing skipped (synchronizing)")
	//	return nil, status.Errorf(codes.Unavailable, "Syncing to latest head, not ready to respond")
	//}

	// An optimistic validator MUST NOT participate in attestation. (i.e., sign across the DOMAIN_BEACON_ATTESTER, DOMAIN_SELECTION_PROOF or DOMAIN_AGGREGATE_AND_PROOF domains).
	if err := vs.optimisticStatus(ctx); err != nil {
		return nil, err
	}

	if err := helpers.ValidateAttestationTime(req.Slot, vs.TimeFetcher.GenesisTime(),
		params.BeaconNetworkConfig().MaximumGossipClockDisparity); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid request: %v", err))
	}

	res, err := vs.AttestationCache.Get(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve data from attestation cache: %v", err)
	}
	if res != nil {
		res.CommitteeIndex = req.CommitteeIndex
		return res, nil
	}

	if err := vs.AttestationCache.MarkInProgress(req); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			res, err := vs.AttestationCache.Get(ctx, req)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not retrieve data from attestation cache: %v", err)
			}
			if res == nil {
				return nil, status.Error(codes.DataLoss, "A request was in progress and resolved to nil")
			}
			res.CommitteeIndex = req.CommitteeIndex
			return res, nil
		}
		return nil, status.Errorf(codes.Internal, "Could not mark attestation as in-progress: %v", err)
	}
	defer func() {
		if err := vs.AttestationCache.MarkNotInProgress(req); err != nil {
			log.WithError(err).Error("Could not mark cache not in progress")
		}
	}()

	optSpines, err := vs.getOptimisticSpine(ctx)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat optimistic spines: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"slot": req.Slot,
		}).Error("Get attestation data: Could not retrieve of gwat optimistic spines")
		return nil, errWrap
	}

	currHead, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		log.WithError(err).Error("Get attestation data: Could not retrieve head state")
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(err).Error("Get attestation data: retrieving of genesis root")
			return nil, status.Errorf(codes.Internal, "Could not retrieve of genesis root: %v", err)
		}
	}

	//calculate optimistic parent root
	supportedRoot, err := vs.HeadFetcher.ForkChoicer().GetParentByOptimisticSpines(ctx, optSpines, jCpRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"extOptSpines": optSpines,
		}).Error("Get attestation data: Could not retrieve of supported root")
		return nil, status.Errorf(codes.Internal, "Could not retrieve of supported root: %v", err)
	}

	log.WithFields(logrus.Fields{
		"1.supportedRoot": fmt.Sprintf("%#x", supportedRoot),
		"2.extOptSpines":  len(optSpines),
	}).Info("Get attestation data: retrieving of gwat optimistic spines")

	headState, err := vs.StateGen.StateByRoot(ctx, supportedRoot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}

	headRoot := supportedRoot[:]
	// In the case that we receive an attestation request after a newer state/block has been processed.
	if headState.Slot() > req.Slot {
		headRoot, err = helpers.BlockRootAtSlot(headState, req.Slot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get historical head root: %v", err)
		}
		headState, err = vs.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(headRoot))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get historical head state: %v", err)
		}
	}

	log.WithFields(logrus.Fields{
		"slot":     headState.Slot(),
		"headRoot": fmt.Sprintf("%#x", headRoot),
		"stPrefix": gwatCommon.HashArrayFromBytes(headState.SpineData().Prefix),
	}).Info("GetAttestationData: prefix validation success")

	if headState == nil || headState.IsNil() {
		return nil, status.Error(codes.Internal, "Could not lookup parent state from head.")
	}

	if time.CurrentEpoch(headState) < slots.ToEpoch(req.Slot) {
		headState, err = transition.ProcessSlotsUsingNextSlotCache(ctx, headState, headRoot, req.Slot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not process slots up to %d: %v", req.Slot, err)
		}
	}

	targetEpoch := time.CurrentEpoch(headState)
	epochStartSlot, err := slots.EpochStart(targetEpoch)
	if err != nil {
		return nil, err
	}
	var targetRoot []byte
	if epochStartSlot == headState.Slot() {
		targetRoot = headRoot
	} else {
		targetRoot, err = helpers.BlockRootAtSlot(headState, epochStartSlot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get target block for slot %d: %v", epochStartSlot, err)
		}
		if bytesutil.ToBytes32(targetRoot) == params.BeaconConfig().ZeroHash {
			targetRoot = headRoot
		}
	}

	res = &ethpb.AttestationData{
		Slot:            req.Slot,
		CommitteeIndex:  req.CommitteeIndex,
		BeaconBlockRoot: headRoot,
		Source:          headState.CurrentJustifiedCheckpoint(),
		Target: &ethpb.Checkpoint{
			Epoch: targetEpoch,
			Root:  targetRoot,
		},
	}

	if err := vs.AttestationCache.Put(ctx, req, res); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not store attestation data in cache: %v", err)
	}
	return res, nil
}

// ProposeAttestation is a function called by an attester to vote
// on a block via an attestation object as defined in the Ethereum Serenity specification.
func (vs *Server) ProposeAttestation(ctx context.Context, att *ethpb.Attestation) (*ethpb.AttestResponse, error) {
	ctx, span := trace.StartSpan(ctx, "AttesterServer.ProposeAttestation")
	defer span.End()

	if _, err := bls.SignatureFromBytes(att.Signature); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Incorrect attestation signature")
	}

	root, err := att.Data.HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not tree hash attestation: %v", err)
	}

	// Broadcast the unaggregated attestation on a feed to notify other services in the beacon node
	// of a received unaggregated attestation.
	vs.OperationNotifier.OperationFeed().Send(&feed.Event{
		Type: operation.UnaggregatedAttReceived,
		Data: &operation.UnAggregatedAttReceivedData{
			Attestation: att,
		},
	})

	// Determine subnet to broadcast attestation to
	wantedEpoch := slots.ToEpoch(att.Data.Slot)
	vals, err := vs.HeadFetcher.HeadValidatorsIndices(ctx, wantedEpoch)
	if err != nil {
		return nil, err
	}
	subnet := helpers.ComputeSubnetFromCommitteeAndSlot(uint64(len(vals)), att.Data.CommitteeIndex, att.Data.Slot)

	// Broadcast the new attestation to the network.
	if err := vs.P2P.BroadcastAttestation(ctx, subnet, att); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not broadcast attestation: %v", err)
	}

	go func() {
		ctx = trace.NewContext(context.Background(), trace.FromContext(ctx))
		attCopy := ethpb.CopyAttestation(att)
		if err := vs.AttPool.SaveUnaggregatedAttestation(attCopy); err != nil {
			log.WithError(err).Error("Could not handle attestation in operations service")
			return
		}
	}()

	return &ethpb.AttestResponse{
		AttestationDataRoot: root[:],
	}, nil
}

// SubscribeCommitteeSubnets subscribes to the committee ID subnet given subscribe request.
func (vs *Server) SubscribeCommitteeSubnets(ctx context.Context, req *ethpb.CommitteeSubnetsSubscribeRequest) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "AttesterServer.SubscribeCommitteeSubnets")
	defer span.End()

	if len(req.Slots) != len(req.CommitteeIds) || len(req.CommitteeIds) != len(req.IsAggregator) {
		return nil, status.Error(codes.InvalidArgument, "request fields are not the same length")
	}
	if len(req.Slots) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no attester slots provided")
	}

	fetchValsLen := func(slot types.Slot) (uint64, error) {
		wantedEpoch := slots.ToEpoch(slot)
		vals, err := vs.HeadFetcher.HeadValidatorsIndices(ctx, wantedEpoch)
		if err != nil {
			return 0, err
		}
		return uint64(len(vals)), nil
	}

	// Request the head validator indices of epoch represented by the first requested
	// slot.
	currValsLen, err := fetchValsLen(req.Slots[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head validator length: %v", err)
	}
	currEpoch := slots.ToEpoch(req.Slots[0])

	log.WithFields(logrus.Fields{
		"req.slot":          req.Slots,
		"req.ProposerSlots": req.ProposerSlots,
		"req.CommitteeIds":  req.CommitteeIds,
		"req.IsAggregator":  req.IsAggregator,
		"currValsLen":       currValsLen,
	}).Info("Validator subscription: SubscribeCommitteeSubnets: slots")

	for i := 0; i < len(req.Slots); i++ {
		// If epoch has changed, re-request active validators length
		if currEpoch != slots.ToEpoch(req.Slots[i]) {
			currValsLen, err = fetchValsLen(req.Slots[i])
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not retrieve head validator length: %v", err)
			}
			currEpoch = slots.ToEpoch(req.Slots[i])
		}
		subnet := helpers.ComputeSubnetFromCommitteeAndSlot(currValsLen, req.CommitteeIds[i], req.Slots[i])
		cache.SubnetIDs.AddAttesterSubnetID(req.Slots[i], subnet)
		if req.IsAggregator[i] {
			cache.SubnetIDs.AddAggregatorSubnetID(req.Slots[i], subnet)
		}
		// prevoting
		subnetPv := helpers.ComputeSubnetPrevotingBySlot(currValsLen, req.Slots[i])
		cache.SubnetIDs.AddPrevotingSubnetID(req.Slots[i]-1, subnetPv)
	}

	for _, ps := range req.ProposerSlots {
		subnet := helpers.ComputeSubnetPrevotingBySlot(currValsLen, ps)
		cache.SubnetIDs.AddPrevotingSubnetID(ps-1, subnet)
	}

	return &emptypb.Empty{}, nil
}
