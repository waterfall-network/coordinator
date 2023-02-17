package v1_test

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/state-native/v1"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestReadOnlyValidator_ReturnsErrorOnNil(t *testing.T) {
	if _, err := v1.NewValidator(nil); err != v1.ErrNilWrappedValidator {
		t.Errorf("Wrong error returned. Got %v, wanted %v", err, v1.ErrNilWrappedValidator)
	}
}

func TestReadOnlyValidator_EffectiveBalance(t *testing.T) {
	bal := uint64(234)
	v, err := v1.NewValidator(&ethpb.Validator{EffectiveBalance: bal})
	require.NoError(t, err)
	assert.Equal(t, bal, v.EffectiveBalance())
}

func TestReadOnlyValidator_ActivationEligibilityEpoch(t *testing.T) {
	epoch := types.Epoch(234)
	v, err := v1.NewValidator(&ethpb.Validator{ActivationEligibilityEpoch: epoch})
	require.NoError(t, err)
	assert.Equal(t, epoch, v.ActivationEligibilityEpoch())
}

func TestReadOnlyValidator_ActivationEpoch(t *testing.T) {
	epoch := types.Epoch(234)
	v, err := v1.NewValidator(&ethpb.Validator{ActivationEpoch: epoch})
	require.NoError(t, err)
	assert.Equal(t, epoch, v.ActivationEpoch())
}

func TestReadOnlyValidator_WithdrawableEpoch(t *testing.T) {
	epoch := types.Epoch(234)
	v, err := v1.NewValidator(&ethpb.Validator{WithdrawableEpoch: epoch})
	require.NoError(t, err)
	assert.Equal(t, epoch, v.WithdrawableEpoch())
}

func TestReadOnlyValidator_ExitEpoch(t *testing.T) {
	epoch := types.Epoch(234)
	v, err := v1.NewValidator(&ethpb.Validator{ExitEpoch: epoch})
	require.NoError(t, err)
	assert.Equal(t, epoch, v.ExitEpoch())
}

func TestReadOnlyValidator_PublicKey(t *testing.T) {
	key := [fieldparams.BLSPubkeyLength]byte{0xFA, 0xCC}
	v, err := v1.NewValidator(&ethpb.Validator{PublicKey: key[:]})
	require.NoError(t, err)
	assert.Equal(t, key, v.PublicKey())
}

func TestReadOnlyValidator_WithdrawalCredentials(t *testing.T) {
	creds := []byte{0xFA, 0xCC}
	v, err := v1.NewValidator(&ethpb.Validator{WithdrawalCredentials: creds})
	require.NoError(t, err)
	assert.DeepEqual(t, creds, v.WithdrawalCredentials())
}

func TestReadOnlyValidator_CreatorAddress(t *testing.T) {
	creds := []byte{0xFA, 0xCC}
	v, err := v1.NewValidator(&ethpb.Validator{CreatorAddress: creds})
	require.NoError(t, err)
	assert.DeepEqual(t, creds, v.CreatorAddress())
}

func TestReadOnlyValidator_Slashed(t *testing.T) {
	slashed := true
	v, err := v1.NewValidator(&ethpb.Validator{Slashed: slashed})
	require.NoError(t, err)
	assert.Equal(t, slashed, v.Slashed())
}
