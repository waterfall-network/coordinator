package blocks

import (
	"bytes"
	"context"
	"fmt"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"math"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

var (
	ErrWithdrawalIsNil             = errors.New("nil withdrawal")
	ErrWithdrawalBadValidatorIndex = errors.New("withdrawal bad ValidatorIndex")
	ErrWithdrawalBadPublicKey      = errors.New("withdrawal bad PublicKey")
	ErrWithdrawalBadEpoch          = errors.New("withdrawal bad Epoch")
	ErrWithdrawalLowBalance        = errors.New("withdrawal low balance")
	ErrWithdrawalAlreadyApplied    = errors.New("withdrawal already applied")
)

// ProcessWithdrawal is one of the operations
// performed on each processed beacon block to
// withdraw amount from validator balance.
func ProcessWithdrawal(
	ctx context.Context,
	beaconState state.BeaconState,
	withdrawals []*ethpb.Withdrawal,
) (state.BeaconState, error) {
	for _, withdrawal := range withdrawals {
		// validate withdrawal operation
		if withdrawal == nil {
			return nil, ErrWithdrawalIsNil
		}
		roVal, err := beaconState.ValidatorAtIndexReadOnly(withdrawal.ValidatorIndex)
		if err != nil {
			return nil, err
		}
		availableBalance, err := helpers.AvailableWithdrawalAmount(withdrawal.ValidatorIndex, beaconState)
		if err != nil {
			return nil, err
		}
		if err = VerifyWithdrawalData(withdrawal, roVal, beaconState.Slot(), availableBalance); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"VIndex":           fmt.Sprintf("%d", withdrawal.ValidatorIndex),
				"PublicKey":        fmt.Sprintf("%#x", withdrawal.PublicKey),
				"Epoch":            fmt.Sprintf("%d", withdrawal.Epoch),
				"Amount":           fmt.Sprintf("%d", withdrawal.Amount),
				"InitTxHash":       fmt.Sprintf("%#x", withdrawal.InitTxHash),
				"availableBalance": availableBalance,
			}).Error("Withdrawal transition: validation failed")

			// if withdrawal already applied - skip err
			if err == ErrWithdrawalAlreadyApplied {
				return beaconState, nil
			}
			return nil, errors.Wrapf(err, "withdrawal verify %d", withdrawal.ValidatorIndex)
		}

		// apply withdrawal
		amount := withdrawal.Amount
		// if amount is zero - withdraw the entire available balance
		if amount == 0 {
			amount = availableBalance
		}

		// write Rewards And Penalties log
		if err = helpers.LogBeforeRewardsAndPenalties(
			beaconState,
			withdrawal.ValidatorIndex,
			amount,
			nil,
			helpers.BalanceDecrease,
			helpers.OpWithdrawal,
		); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"Slot":       beaconState.Slot(),
				"Validator":  withdrawal.ValidatorIndex,
				"withdrawal": amount,
			}).Error("Log rewards and penalties failed: ProcessWithdrawal")
		}

		// update balance
		if err = helpers.DecreaseBalance(beaconState, withdrawal.ValidatorIndex, amount); err != nil {
			return nil, err
		}
		// update validator
		upVal, err := beaconState.ValidatorAtIndex(withdrawal.ValidatorIndex)
		if err != nil {
			return nil, err
		}
		if upVal.WithdrawalOps == nil {
			upVal.WithdrawalOps = make([]*ethpb.WithdrawalOp, 1)
		}
		// check WithdrawalOps length limit
		if uint64(len(upVal.WithdrawalOps)) >= params.BeaconConfig().WithdrawalOpsLimit {
			upVal.WithdrawalOps = upVal.WithdrawalOps[1:]
		}

		upVal.WithdrawalOps = append(upVal.WithdrawalOps, &ethpb.WithdrawalOp{
			Amount: amount,
			Hash:   bytesutil.SafeCopyBytes(withdrawal.InitTxHash),
			Slot:   beaconState.Slot(),
		})
		if err = beaconState.UpdateValidatorAtIndex(withdrawal.ValidatorIndex, upVal); err != nil {
			return nil, err
		}
	}
	return beaconState, nil
}

func VerifyWithdrawalData(
	withdrawal *ethpb.Withdrawal,
	validator state.ReadOnlyValidator,
	currentSlot types.Slot,
	availableBalance uint64,
) error {
	if withdrawal == nil {
		return ErrWithdrawalIsNil
	}
	if withdrawal.ValidatorIndex == math.MaxUint64 {
		return ErrWithdrawalBadValidatorIndex
	}

	if bytesutil.ToBytes48(withdrawal.PublicKey) != validator.PublicKey() {
		return ErrWithdrawalBadPublicKey
	}

	currentEpoch := slots.ToEpoch(currentSlot)
	if withdrawal.Epoch > currentEpoch {
		return ErrWithdrawalBadEpoch
	}

	if availableBalance < withdrawal.Amount {
		return ErrWithdrawalLowBalance
	}

	for _, v := range validator.WithdrawalOps() {
		if bytes.Equal(v.Hash, withdrawal.InitTxHash) {
			return ErrWithdrawalAlreadyApplied
		}
	}

	return nil
}
