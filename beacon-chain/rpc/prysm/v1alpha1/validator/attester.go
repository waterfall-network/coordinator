package validator

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/beacon-chain/cache"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/feed"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/feed/operation"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/time"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/transition"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/crypto/bls"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/time/slots"
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

	headState, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	headRoot, err := vs.HeadFetcher.HeadRoot(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head root: %v", err)
	}

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

	//validate state's candidates
	checkPoint := headState.CurrentJustifiedCheckpoint()
	cpSlot, err := slots.EpochStart(checkPoint.Epoch)
	if err != nil {
		log.WithError(err).Warn("GetAttestationData: candidates validation err")
		return nil, status.Errorf(codes.Internal, "Could not calculate 1st slot of epoch=%d: %v", checkPoint.Epoch, err)
	}
	//get last valid block info
	validatedRoot, validatedSlot := vs.HeadFetcher.GetValidatedBlockInfo()
	for {

		log.WithFields(logrus.Fields{
			"slot":                                 headState.Slot(),
			"headRoot":                             fmt.Sprintf("%#x", headRoot),
			"validatedRoot":                        fmt.Sprintf("%#x", validatedRoot),
			"cpSlot":                               cpSlot,
			"validatedSlot":                        validatedSlot,
			"bytes.Equal(headRoot, validatedRoot)": bytes.Equal(headRoot, validatedRoot),
		}).Warn("GetAttestationData: candidates validation by cache")

		if bytes.Equal(headRoot, validatedRoot) {
			break
		}

		if headState.Slot() <= cpSlot {
			log.WithError(status.Errorf(codes.Internal, "Not found valid candidates after checkpoint: cp.Slot=%d cp.Slot=%v", cpSlot, checkPoint.Root)).WithFields(logrus.Fields{
				"cp.Slot": cpSlot,
				"cp.Root": checkPoint.Root,
			}).Error("GetAttestationData: candidates validation failed")
			return nil, status.Errorf(codes.Internal, "Not found valid candidates after checkpoint: cp.Slot=%d cp.Slot=%v", cpSlot, checkPoint.Root)
		}

		// gwat validation
		candidates := gwatCommon.HashArrayFromBytes(headState.Eth1Data().Candidates)
		log.WithFields(logrus.Fields{
			"slot":            headState.Slot(),
			"headRoot":        fmt.Sprintf("%#x", headRoot),
			"validatedRoot":   fmt.Sprintf("%#x", validatedRoot),
			"validatedSlot":   validatedSlot,
			"blockCandidates": candidates,
		}).Warn("GetAttestationData: candidates validation by gwat")

		if len(candidates) == 0 {
			break
		}
		isValidCandidates, err := vs.ExecutionEngineCaller.ExecutionDagValidateSpines(ctx, candidates)
		if isValidCandidates {
			vs.HeadFetcher.SetValidatedBlockInfo(headRoot, headState.Slot())
			break
		}
		if err != nil {
			log.WithError(status.Errorf(codes.Internal, "Could not get lastValidRoot state: %v", err)).WithFields(logrus.Fields{
				"slotCandidates": isValidCandidates,
				"candidates":     candidates,
			}).Error("GetAttestationData: candidates validation by gwat failed")
		}
		if ctx.Err() == context.Canceled {
			return nil, status.Errorf(codes.Internal, "%v", ctx.Err())
		}
		// try previous slot
		headRoot, err = helpers.BlockRootAtSlot(headState, headState.Slot()-1)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"prevSlot": headState.Slot() - 1,
			}).Error("GetAttestationData: candidates validation by gwat failed")
			return nil, status.Errorf(codes.Internal, "Could not get historical head root: %v", err)
		}
		headState, err = vs.StateGen.StateByRoot(ctx, bytesutil.ToBytes32(headRoot))
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"headRoot": fmt.Sprintf("%#x", headRoot),
			}).Error("GetAttestationData: candidates validation by gwat failed")
			return nil, status.Errorf(codes.Internal, "Could not get historical head state: %v", err)
		}
	}

	log.WithFields(logrus.Fields{
		"slot":       headState.Slot(),
		"headRoot":   fmt.Sprintf("%#x", headRoot),
		"candidates": gwatCommon.HashArrayFromBytes(headState.Eth1Data().Candidates),
	}).Warn("GetAttestationData: candidates validation success")

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
	}

	return &emptypb.Empty{}, nil
}
