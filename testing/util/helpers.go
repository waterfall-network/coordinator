package util

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/rand"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
)

// RandaoReveal returns a signature of the requested epoch using the beacon proposer private key.
func RandaoReveal(beaconState state.ReadOnlyBeaconState, epoch types.Epoch, privKeys []bls.SecretKey) ([]byte, error) {
	// We fetch the proposer's index as that is whom the RANDAO will be verified against.
	proposerIdx, err := helpers.BeaconProposerIndex(context.Background(), beaconState)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not get beacon proposer index")
	}
	buf := make([]byte, 32)
	binary.LittleEndian.PutUint64(buf, uint64(epoch))

	// We make the previous validator's index sign the message instead of the proposer.
	sszEpoch := types.SSZUint64(epoch)
	return signing.ComputeDomainAndSign(beaconState, epoch, &sszEpoch, params.BeaconConfig().DomainRandao, privKeys[proposerIdx])
}

// BlockSignature calculates the post-state root of the block and returns the signature.
func BlockSignature(
	bState state.BeaconState,
	block *ethpb.BeaconBlock,
	privKeys []bls.SecretKey,
) (bls.Signature, error) {
	wsb, err := wrapper.WrappedSignedBeaconBlock(&ethpb.SignedBeaconBlock{Block: block})
	if err != nil {
		return nil, errors.Wrap(err, "could not wrap block")
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

// Random32Bytes generates a random 32 byte slice.
func Random32Bytes(t *testing.T) []byte {
	b := make([]byte, 32)
	_, err := rand.NewDeterministicGenerator().Read(b)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
