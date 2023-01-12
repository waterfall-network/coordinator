package validator

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition/interop"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
)

func (vs *Server) getBellatrixBeaconBlock(ctx context.Context, req *ethpb.BlockRequest) (*ethpb.BeaconBlockBellatrix, error) {
	altairBlk, err := vs.buildAltairBeaconBlock(ctx, req)
	if err != nil {
		return nil, err
	}

	payload, err := vs.getExecutionPayload(ctx, req.Slot, altairBlk.ProposerIndex)
	if err != nil {
		return nil, err
	}

	blk := &ethpb.BeaconBlockBellatrix{
		Slot:          altairBlk.Slot,
		ProposerIndex: altairBlk.ProposerIndex,
		ParentRoot:    altairBlk.ParentRoot,
		StateRoot:     params.BeaconConfig().ZeroHash[:],
		Body: &ethpb.BeaconBlockBodyBellatrix{
			RandaoReveal:      altairBlk.Body.RandaoReveal,
			Eth1Data:          altairBlk.Body.Eth1Data,
			Graffiti:          altairBlk.Body.Graffiti,
			ProposerSlashings: altairBlk.Body.ProposerSlashings,
			AttesterSlashings: altairBlk.Body.AttesterSlashings,
			Attestations:      altairBlk.Body.Attestations,
			Deposits:          altairBlk.Body.Deposits,
			VoluntaryExits:    altairBlk.Body.VoluntaryExits,
			SyncAggregate:     altairBlk.Body.SyncAggregate,
			ExecutionPayload:  payload,
		},
	}
	// Compute state root with the newly constructed block.
	wsb, err := wrapper.WrappedSignedBeaconBlock(
		&ethpb.SignedBeaconBlockBellatrix{Block: blk, Signature: make([]byte, 96)},
	)
	if err != nil {
		return nil, err
	}
	stateRoot, err := vs.computeStateRoot(ctx, wsb)

	log.WithError(err).WithFields(logrus.Fields{
		"block.slot": wsb.Block().Slot(),
	}).Info("<<<< getBellatrixBeaconBlock:computeStateRoot >>>>> 6666666")

	if err != nil {
		interop.WriteBlockToDisk(wsb, true /*failed*/)
		return nil, fmt.Errorf("could not compute state root: %v", err)
	}
	blk.StateRoot = stateRoot
	return blk, nil
}
