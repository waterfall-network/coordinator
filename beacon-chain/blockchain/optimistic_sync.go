package blockchain

import (
	"context"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db/kv"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	enginev1 "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/engine/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

// getPayloadAttributes returns the payload attributes for the given state and slot.
// The attribute is required to initiate a payload build process in the context of an `engine_forkchoiceUpdated` call.
func (s *Service) getPayloadAttribute(ctx context.Context, st state.BeaconState, slot types.Slot) (bool, *enginev1.PayloadAttributes, types.ValidatorIndex, error) {
	proposerID, _, ok := s.cfg.ProposerSlotIndexCache.GetProposerPayloadIDs(slot)
	if !ok { // There's no need to build attribute if there is no proposer for slot.
		return false, nil, 0, nil
	}

	// Get previous randao.
	st = st.Copy()
	st, err := transition.ProcessSlotsIfPossible(ctx, st, slot)
	if err != nil {
		return false, nil, 0, err
	}
	prevRando, err := helpers.RandaoMix(st, time.CurrentEpoch(st))
	if err != nil {
		return false, nil, 0, nil
	}

	// Get fee recipient.
	feeRecipient := params.BeaconConfig().DefaultFeeRecipient
	recipient, err := s.cfg.BeaconDB.FeeRecipientByValidatorID(ctx, proposerID)
	switch {
	case errors.Is(err, kv.ErrNotFoundFeeRecipient):
		if feeRecipient.String() == fieldparams.EthBurnAddressHex {
			logrus.WithFields(logrus.Fields{
				"validatorIndex": proposerID,
				"burnAddress":    fieldparams.EthBurnAddressHex,
			}).Error("Fee recipient not set. Using burn address")
		}
	case err != nil:
		return false, nil, 0, errors.Wrap(err, "could not get fee recipient in db")
	default:
		feeRecipient = recipient
	}

	// Get timestamp.
	t, err := slots.ToTime(uint64(s.genesisTime.Unix()), slot)
	if err != nil {
		return false, nil, 0, err
	}
	attr := &enginev1.PayloadAttributes{
		Timestamp:             uint64(t.Unix()),
		PrevRandao:            prevRando,
		SuggestedFeeRecipient: feeRecipient.Bytes(),
	}
	return true, attr, proposerID, nil
}

// removeInvalidBlockAndState removes the invalid block and its corresponding state from the cache and DB.
func (s *Service) removeInvalidBlockAndState(ctx context.Context, blkRoots [][32]byte) error {
	for _, root := range blkRoots {
		if err := s.cfg.StateGen.DeleteStateFromCaches(ctx, root); err != nil {
			return err
		}

		// Delete block also deletes the state as well.
		if err := s.cfg.BeaconDB.DeleteBlock(ctx, root); err != nil {
			// TODO(10487): If a caller requests to delete a root that's justified and finalized. We should gracefully shutdown.
			// This is an irreparable condition, it would me a justified or finalized block has become invalid.
			return err
		}
	}
	return nil
}
