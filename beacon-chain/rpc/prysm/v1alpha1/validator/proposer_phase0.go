package validator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition/interop"
	v "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/validators"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// blockData required to create a beacon block.
type blockData struct {
	ParentRoot        []byte
	Graffiti          [32]byte
	ProposerIdx       types.ValidatorIndex
	Eth1Data          *ethpb.Eth1Data
	Deposits          []*ethpb.Deposit
	Attestations      []*ethpb.Attestation
	ProposerSlashings []*ethpb.ProposerSlashing
	AttesterSlashings []*ethpb.AttesterSlashing
	VoluntaryExits    []*ethpb.VoluntaryExit
	Withdrawals       []*ethpb.Withdrawal
}

func (vs *Server) getPhase0BeaconBlock(ctx context.Context, req *ethpb.BlockRequest) (*ethpb.BeaconBlock, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.getPhase0BeaconBlock")
	defer span.End()
	blkData, err := vs.buildPhase0BlockData(ctx, req)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"req.Slot": req.Slot,
		}).Error("Build beacon block Phase0: could not build data")
		return nil, fmt.Errorf("could not build block data: %v", err)
	}

	log.WithFields(logrus.Fields{
		"req.slot":           req.Slot,
		"blkData.Candidates": gwatCommon.HashArrayFromBytes(blkData.Eth1Data.Candidates),
	}).Info("Build beacon block Phase0: start")

	// Use zero hash as stub for state root to compute later.
	stateRoot := params.BeaconConfig().ZeroHash[:]

	blk := &ethpb.BeaconBlock{
		Slot:          req.Slot,
		ParentRoot:    blkData.ParentRoot,
		StateRoot:     stateRoot,
		ProposerIndex: blkData.ProposerIdx,
		Body: &ethpb.BeaconBlockBody{
			Eth1Data:          blkData.Eth1Data,
			Deposits:          blkData.Deposits,
			Attestations:      blkData.Attestations,
			RandaoReveal:      req.RandaoReveal,
			ProposerSlashings: blkData.ProposerSlashings,
			AttesterSlashings: blkData.AttesterSlashings,
			VoluntaryExits:    blkData.VoluntaryExits,
			Graffiti:          blkData.Graffiti[:],
			Withdrawals:       blkData.Withdrawals,
		},
	}

	// Compute state root with the newly constructed block.
	wsb, err := wrapper.WrappedSignedBeaconBlock(&ethpb.SignedBeaconBlock{Block: blk, Signature: make([]byte, 96)})
	if err != nil {
		return nil, err
	}
	stateRoot, err = vs.computeStateRoot(ctx, wsb)

	log.WithError(err).WithFields(logrus.Fields{
		"block.slot": wsb.Block().Slot(),
	}).Info("<<<< getPhase0BeaconBlock:computeStateRoot >>>>> 000000")

	if err != nil {
		interop.WriteBlockToDisk(wsb, true /*failed*/)
		return nil, errors.Wrap(err, "could not compute state root")
	}
	blk.StateRoot = stateRoot
	return blk, nil
}

// Build data required for creating a new beacon block, so this method can be shared across forks.
func (vs *Server) buildPhase0BlockData(ctx context.Context, req *ethpb.BlockRequest) (*blockData, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.buildPhase0BlockData")
	defer span.End()

	if vs.SyncChecker.Syncing() {
		log.WithError(fmt.Errorf("syncing to latest head, not ready to respond")).WithFields(logrus.Fields{
			"Syncing":  vs.SyncChecker.Syncing(),
			"req.slot": req.Slot,
		}).Warn("Proposing skipped (synchronizing)")
		return nil, fmt.Errorf("syncing to latest head, not ready to respond")
	}
	//if vs.HeadFetcher.IsGwatSynchronizing() {
	//	log.WithError(fmt.Errorf("GWAT synchronization process is running, not ready to respond")).WithFields(logrus.Fields{
	//		"Syncing": vs.HeadFetcher.IsGwatSynchronizing(),
	//	}).Warn("Proposing skipped (synchronizing)")
	//	return nil, fmt.Errorf("GWAT synchronization process is running, not ready to respond")
	//}

	optSpines, err := vs.getOptimisticSpine(ctx)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat optimistic spines: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"slot": req.Slot,
		}).Error("Build block data: Could not retrieve of gwat optimistic spines")
		return nil, errWrap
	}
	if len(optSpines) == 0 {
		log.Errorf("Empty list of optimistic spines was retrieved for slot: %v", req.Slot)
	}

	currHead, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		log.WithError(err).Error("Build block data:  Could not retrieve head state")
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(err).Error("Build block data:  retrieving of genesis root")
			return nil, status.Errorf(codes.Internal, "Could not retrieve of genesis root: %v", err)
		}
	}

	//calculate optimistic parent root
	parentRoot, err := vs.HeadFetcher.ForkChoicer().GetParentByOptimisticSpines(ctx, optSpines, jCpRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"req.slot":     req.Slot,
			"extOptSpines": optSpines,
		}).Error("Build block data: retrieving of parent root failed")
		return nil, err
	}
	if parentRoot == ([32]byte{}) {
		log.WithError(err).WithFields(logrus.Fields{
			"req.slot":   req.Slot,
			"parentRoot": fmt.Sprintf("%#x", parentRoot),
		}).Error("Build block data: retrieving empty parent root")
		return nil, fmt.Errorf("empty parent root %#x (slot=%d)", parentRoot, req.Slot)
	}

	log.WithFields(logrus.Fields{
		"req.slot":       req.Slot,
		"1.parentRoot":   fmt.Sprintf("%#x", parentRoot),
		"2.extOptSpines": len(optSpines),
	}).Info("Build block data: get parent root")

	//head, err := vs.StateGen.SyncStateByRoot(ctx, parentRoot)
	head, err := vs.StateGen.StateByRoot(ctx, parentRoot)
	if err != nil {
		return nil, fmt.Errorf("could not get head state %v", err)
	}

	log.WithFields(logrus.Fields{
		"0:req.slot":    req.Slot,
		"1:stSlot":      head.Slot(),
		"2.parentRoot":  fmt.Sprintf("%#x", parentRoot),
		"3:stBlockHash": fmt.Sprintf("%#x", head.Eth1Data().BlockHash),
	}).Info("Build block data: get parent state")

	head, err = transition.ProcessSlotsUsingNextSlotCache(ctx, head, parentRoot[:], req.Slot)
	if err != nil {
		return nil, fmt.Errorf("could not advance slots to calculate proposer index: %v", err)
	}

	eth1Data, err := vs.eth1DataMajorityVote(ctx, head)
	if err != nil {
		return nil, fmt.Errorf("could not get ETH1 data: %v", err)
	}

	prevoteData := vs.PrevotePool.GetPrevoteBySlot(ctx, req.Slot)
	candidates := helpers.CalculateCandidates(head, optSpines)
	log.WithFields(logrus.Fields{
		"0:slot":     req.Slot,
		"1:prevotes": len(prevoteData),
	}).Infof("Build block data: get prevotes")

	if len(prevoteData) == 0 {
		log.Warnf("Build block data: no prevote data was retrieved for slot %v", req.Slot)
	} else {
		prevoteCandidates := vs.prepareAndProcessPrevoteData(candidates.Copy(), prevoteData, head)
		if len(prevoteCandidates) == 0 {
			log.Warn("Build block data: prevote data was processed but returned empty candidates array, fallback to candidates" +
				" retrieved using optimistic spines")
		} else {
			candidates = prevoteCandidates
		}
	}

	eth1Data.Candidates = candidates.ToBytes()
	log.WithFields(logrus.Fields{
		"1.req.Slot":   req.Slot,
		"2.candidates": candidates,
	}).Info("Build block data: candidates which will be added to block")

	deposits, atts, err := vs.packDepositsAndAttestations(ctx, head, eth1Data, parentRoot)
	if err != nil {
		return nil, err
	}

	graffiti := bytesutil.ToBytes32(req.Graffiti)

	// Calculate new proposer index.
	idx, err := helpers.BeaconProposerIndex(ctx, head)
	if err != nil {
		return nil, fmt.Errorf("could not calculate proposer index %v", err)
	}

	proposerSlashings := vs.SlashingsPool.PendingProposerSlashings(ctx, head, false /*noLimit*/)
	validProposerSlashings := make([]*ethpb.ProposerSlashing, 0, len(proposerSlashings))
	for _, slashing := range proposerSlashings {
		_, err := blocks.ProcessProposerSlashing(ctx, head, slashing, v.SlashValidator)
		if err != nil {
			log.WithError(err).Warn("Proposer: invalid proposer slashing")
			continue
		}
		validProposerSlashings = append(validProposerSlashings, slashing)
	}

	attSlashings := vs.SlashingsPool.PendingAttesterSlashings(ctx, head, false /*noLimit*/)
	validAttSlashings := make([]*ethpb.AttesterSlashing, 0, len(attSlashings))
	for _, slashing := range attSlashings {
		_, err := blocks.ProcessAttesterSlashing(ctx, head, slashing, v.SlashValidator)
		if err != nil {
			log.WithError(err).Warn("Proposer: invalid attester slashing")
			continue
		}
		validAttSlashings = append(validAttSlashings, slashing)
	}

	exits := vs.ExitPool.PendingExits(head, req.Slot, false)
	validExits := make([]*ethpb.VoluntaryExit, 0, len(exits))
	for _, exit := range exits {
		val, err := head.ValidatorAtIndexReadOnly(exit.ValidatorIndex)
		if err != nil {
			log.WithError(err).Warn("Proposer: exit op: get validator feiled ")
			continue
		}
		if err := blocks.VerifyExitData(val, head.Slot(), exit); err != nil {
			log.WithError(err).Warn("Proposer: exit op: invalid data")
			continue
		}
		validExits = append(validExits, exit)
	}

	withdrawals := vs.WithdrawalPool.PendingWithdrawals(req.Slot, head, false)

	if len(withdrawals) > 0 {
		log.WithFields(logrus.Fields{
			"req.Slot":    req.Slot,
			"withdrawals": len(withdrawals),
		}).Info("Build block data: add withdrawals")
	}

	return &blockData{
		ParentRoot:        parentRoot[:],
		Graffiti:          graffiti,
		ProposerIdx:       idx,
		Eth1Data:          eth1Data,
		Deposits:          deposits,
		Attestations:      atts,
		ProposerSlashings: validProposerSlashings,
		AttesterSlashings: validAttSlashings,
		VoluntaryExits:    validExits,
		Withdrawals:       withdrawals,
	}, nil
}

func (vs *Server) prepareAndProcessPrevoteData(optCandidates gwatCommon.HashArray, prevoteData []*ethpb.PreVote, head state.BeaconState) gwatCommon.HashArray {
	// Make every prevote candidate hash as a separate hasharray to trim non-relevant spines
	// using head
	for i, pv := range prevoteData {
		prevoteSpines := make([]gwatCommon.HashArray, 0, len(gwatCommon.HashArrayFromBytes(pv.Data.Candidates)))
		for i := 0; i < len(pv.Data.Candidates); i += gwatCommon.HashLength {
			h := gwatCommon.BytesToHash(pv.Data.Candidates[i : i+gwatCommon.HashLength])
			prevoteSpines = append(prevoteSpines, gwatCommon.HashArray{h})
		}
		prevoteCandidates := helpers.CalculateCandidates(head, prevoteSpines)

		// Replace initial prevote candidates with corresponding processed ones to save candidates-votes mapping
		prevoteData[i].Data.Candidates = prevoteCandidates.ToBytes()
	}

	// Process prevote data and calculate longest chain of spines with most of the votes
	return vs.processPrevoteData(prevoteData, optCandidates)
}
