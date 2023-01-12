package deposit_test

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/contracts/deposit"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestDepositInput_GeneratesPb(t *testing.T) {
	k1, err := bls.RandKey()
	require.NoError(t, err)
	k2, err := bls.RandKey()
	require.NoError(t, err)

	result, _, err := deposit.DepositInput(k1, k2, 0)
	require.NoError(t, err)
	assert.DeepEqual(t, k1.PublicKey().Marshal(), result.PublicKey)

	sig, err := bls.SignatureFromBytes(result.Signature)
	require.NoError(t, err)
	testData := &ethpb.DepositMessage{
		PublicKey:             result.PublicKey,
		WithdrawalCredentials: result.WithdrawalCredentials,
		Amount:                result.Amount,
	}
	sr, err := testData.HashTreeRoot()
	require.NoError(t, err)
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainDeposit,
		nil, /*forkVersion*/
		nil, /*genesisValidatorsRoot*/
	)
	require.NoError(t, err)
	root, err := (&ethpb.SigningData{ObjectRoot: sr[:], Domain: domain}).HashTreeRoot()
	require.NoError(t, err)
	assert.Equal(t, true, sig.Verify(k1.PublicKey(), root[:]))
}

func TestVerifyDepositSignature_ValidSig(t *testing.T) {
	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	dep := deposits[0]
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainDeposit,
		params.BeaconConfig().GenesisForkVersion,
		params.BeaconConfig().ZeroHash[:],
	)
	require.NoError(t, err)
	err = deposit.VerifyDepositSignature(dep.Data, domain)
	require.NoError(t, err)
}

func TestVerifyDepositSignature_InvalidSig(t *testing.T) {
	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	dep := deposits[0]
	domain, err := signing.ComputeDomain(
		params.BeaconConfig().DomainDeposit,
		params.BeaconConfig().GenesisForkVersion,
		params.BeaconConfig().ZeroHash[:],
	)
	require.NoError(t, err)
	dep.Data.Signature = dep.Data.Signature[1:]
	err = deposit.VerifyDepositSignature(dep.Data, domain)
	if err == nil {
		t.Fatal("Deposit Verification succeeds with a invalid signature")
	}
}
