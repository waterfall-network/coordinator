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
	VoluntaryExits    []*ethpb.SignedVoluntaryExit
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
		"req.slot":             req.Slot,
		"blkData.Finalization": gwatCommon.HashArrayFromBytes(blkData.Eth1Data.Finalization),
		"blkData.Candidates":   gwatCommon.HashArrayFromBytes(blkData.Eth1Data.Candidates),
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

	// Retrieve the parent block as the current head of the canonical chain.
	parentRoot, err := vs.HeadFetcher.HeadRoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve head root: %v", err)
	}

	head, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get head state %v", err)
	}

	head, err = transition.ProcessSlotsUsingNextSlotCache(ctx, head, parentRoot, req.Slot)
	if err != nil {
		return nil, fmt.Errorf("could not advance slots to calculate proposer index: %v", err)
	}

	eth1Data, err := vs.eth1DataMajorityVote(ctx, head)
	if err != nil {
		return nil, fmt.Errorf("could not get ETH1 data: %v", err)
	}

	//retrieving of gwat candidates
	const CandidatesСutoffSlots = 2
	candidates, err := vs.ExecutionEngineCaller.ExecutionDagGetCandidates(ctx, req.Slot-CandidatesСutoffSlots)
	// todo handle err
	if err != nil {
		log.WithError(fmt.Errorf("could not get gwat candidates: %v", err)).WithFields(logrus.Fields{
			"req.Slot":    req.Slot,
			"cutoff.Slot": req.Slot - CandidatesСutoffSlots,
			"candidates":  candidates,
		}).Error("build block data: retrieving of gwat candidates failed")
	} else {
		eth1Data.Candidates = candidates.ToBytes()
		log.WithFields(logrus.Fields{
			"req.Slot":    req.Slot,
			"cutoff.Slot": req.Slot - CandidatesСutoffSlots,
			"candidates":  candidates,
		}).Info("build block data: retrieving of gwat candidates")
	}

	deposits, atts, err := vs.packDepositsAndAttestations(ctx, head, eth1Data)
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
	validExits := make([]*ethpb.SignedVoluntaryExit, 0, len(exits))
	for _, exit := range exits {
		val, err := head.ValidatorAtIndexReadOnly(exit.Exit.ValidatorIndex)
		if err != nil {
			log.WithError(err).Warn("Proposer: invalid exit")
			continue
		}
		if err := blocks.VerifyExitAndSignature(val, head.Slot(), head.Fork(), exit, head.GenesisValidatorsRoot()); err != nil {
			log.WithError(err).Warn("Proposer: invalid exit")
			continue
		}
		validExits = append(validExits, exit)
	}

	return &blockData{
		ParentRoot:        parentRoot,
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
