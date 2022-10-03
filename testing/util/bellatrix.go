package util

import (
	"context"

	"github.com/waterfall-foundation/coordinator/beacon-chain/core/helpers"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/signing"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/time"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/transition"
	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/crypto/bls"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/wrapper"
)

// BlockSignatureBellatrix calculates the post-state root of the block and returns the signature.
func BlockSignatureBellatrix(
	bState state.BeaconState,
	block *ethpb.BeaconBlockBellatrix,
	privKeys []bls.SecretKey,
) (bls.Signature, error) {
	var err error
	wsb, err := wrapper.WrappedSignedBeaconBlock(&ethpb.SignedBeaconBlockBellatrix{Block: block})
	if err != nil {
		return nil, err
	}
	s, err := transition.CalculateStateRoot(context.Background(), bState, wsb)
	if err != nil {
		return nil, err
	}
	block.StateRoot = s[:]
	domain, err := signing.Domain(bState.Fork(), time.CurrentEpoch(bState), params.BeaconConfig().DomainBeaconProposer, bState.GenesisValidatorsRoot())
	if err != nil {
		return nil, err
	}
	blockRoot, err := signing.ComputeSigningRoot(block, domain)
	if err != nil {
		return nil, err
	}
	// Temporarily increasing the beacon state slot here since BeaconProposerIndex is a
	// function deterministic on beacon state slot.
	currentSlot := bState.Slot()
	if err := bState.SetSlot(block.Slot); err != nil {
		return nil, err
	}
	proposerIdx, err := helpers.BeaconProposerIndex(context.Background(), bState)
	if err != nil {
		return nil, err
	}
	if err := bState.SetSlot(currentSlot); err != nil {
		return nil, err
	}
	return privKeys[proposerIdx].Sign(blockRoot[:]), nil
}
