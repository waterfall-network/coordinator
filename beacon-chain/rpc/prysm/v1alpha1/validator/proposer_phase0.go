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
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
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
}

func (vs *Server) getPhase0BeaconBlock(ctx context.Context, req *ethpb.BlockRequest) (*ethpb.BeaconBlock, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.getPhase0BeaconBlock")
	defer span.End()
	blkData, err := vs.buildPhase0BlockData(ctx, req)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"req": req,
		}).Error("#### build-Phase0-BeaconBlock: could not build block data ###")
		return nil, fmt.Errorf("could not build block data: %v", err)
	}

	log.WithFields(logrus.Fields{
		"req.slot":           req.Slot,
		"blkData.Candidates": gwatCommon.HashArrayFromBytes(blkData.Eth1Data.Candidates),
	}).Info("#### get-Phase0Beacon-Block ###")

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
			"Syncing": vs.SyncChecker.Syncing(),
		}).Warn("Proposing skipped (synchronizing)")
		return nil, fmt.Errorf("syncing to latest head, not ready to respond")
	}
	//if vs.HeadFetcher.IsGwatSynchronizing() {
	//	log.WithError(fmt.Errorf("GWAT synchronization process is running, not ready to respond")).WithFields(logrus.Fields{
	//		"Syncing": vs.HeadFetcher.IsGwatSynchronizing(),
	//	}).Warn("Proposing skipped (synchronizing)")
	//	return nil, fmt.Errorf("GWAT synchronization process is running, not ready to respond")
	//}

	// calculate the parent block by optimistic spine.
	currHead, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		log.WithError(err).Error("build block data: retrieving of head state failed")
		return nil, fmt.Errorf("could not get head state %v", err)
	}

	jCpRoot := bytesutil.ToBytes32(currHead.CurrentJustifiedCheckpoint().Root)
	if currHead.CurrentJustifiedCheckpoint().Epoch == 0 {
		jCpRoot, err = vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			log.WithError(err).Error("build block data: retrieving of genesis root")
			return nil, errors.Wrap(err, "get genesis root failed")
		}
	}

	cpSt, err := vs.StateGen.StateByRoot(ctx, jCpRoot)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"jCpRoot": fmt.Sprintf("%#x", jCpRoot),
		}).Error("build block data: retrieving of cp state failed")
		return nil, err
	}

	//request optimistic spine
	baseSpine := helpers.GetTerminalFinalizedSpine(cpSt)

	optSpines, err := vs.HeadFetcher.GetOptimisticSpines(ctx, baseSpine)
	if err != nil {
		errWrap := fmt.Errorf("could not get gwat candidates: %v", err)
		log.WithError(errWrap).WithFields(logrus.Fields{
			"baseSpine": baseSpine,
		}).Error("build block data: retrieving of parent failed")
		return nil, errWrap
	}

	//prepend current optimistic finalization to optimistic spine to calc parent
	optFinalisation := make([]gwatCommon.HashArray, len(cpSt.SpineData().Finalization)/gwatCommon.HashLength)
	for i, h := range gwatCommon.HashArrayFromBytes(cpSt.SpineData().Finalization) {
		optFinalisation[i] = gwatCommon.HashArray{h}
	}
	optSpines = append(optFinalisation, optSpines...)

	//calculate optimistic parent root
	parentRoot, err := vs.HeadFetcher.ForkChoicer().GetParentByOptimisticSpines(ctx, optSpines)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"extOptSpines": optSpines,
		}).Error("build block data: retrieving of gwat optimistic spines failed")
	}

	log.WithFields(logrus.Fields{
		"1.parentRoot":   fmt.Sprintf("%#x", parentRoot),
		"2.extOptSpines": len(optSpines),
	}).Info("build block data: retrieving of gwat optimistic spines")

	head, err := vs.StateGen.StateByRoot(ctx, parentRoot)
	if err != nil {
		return nil, fmt.Errorf("could not get head state %v", err)
	}

	head, err = transition.ProcessSlotsUsingNextSlotCache(ctx, head, parentRoot[:], req.Slot)
	if err != nil {
		return nil, fmt.Errorf("could not advance slots to calculate proposer index: %v", err)
	}

	eth1Data, err := vs.eth1DataMajorityVote(ctx, head)
	if err != nil {
		return nil, fmt.Errorf("could not get ETH1 data: %v", err)
	}

	candidates := helpers.CalculateCandidates(head, optSpines)
	eth1Data.Candidates = candidates.ToBytes()
	log.WithFields(logrus.Fields{
		"1.req.Slot":   req.Slot,
		"2.candidates": candidates,
	}).Info("build block data: retrieving of gwat candidates")

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
	exits := vs.ExitPool.PendingExits(head, req.Slot, false /*noLimit*/)
	validExits := make([]*ethpb.VoluntaryExit, 0, len(exits))
	for _, exit := range exits {
		val, err := head.ValidatorAtIndexReadOnly(exit.ValidatorIndex)
		if err != nil {
			log.WithError(err).Warn("Proposer: invalid exit")
			continue
		}
		if err := blocks.VerifyExitData(val, head.Slot(), exit); err != nil {
			log.WithError(err).Warn("Proposer: invalid exit")
			continue
		}
		validExits = append(validExits, exit)
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
	}, nil
}
