package simulator

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	mock "github.com/waterfall-foundation/coordinator/beacon-chain/blockchain/testing"
	dbtest "github.com/waterfall-foundation/coordinator/beacon-chain/db/testing"
	mockstategen "github.com/waterfall-foundation/coordinator/beacon-chain/state/stategen/mock"
	"github.com/waterfall-foundation/coordinator/crypto/bls"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/testing/util"
)

func setupService(t *testing.T, params *Parameters) *Simulator {
	slasherDB := dbtest.SetupSlasherDB(t)
	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)

	// We setup validators in the beacon state along with their
	// private keys used to generate valid signatures in generated objects.
	validators := make([]*ethpb.Validator, params.NumValidators)
	privKeys := make(map[types.ValidatorIndex]bls.SecretKey)
	for valIdx := range validators {
		privKey, err := bls.RandKey()
		require.NoError(t, err)
		privKeys[types.ValidatorIndex(valIdx)] = privKey
		validators[valIdx] = &ethpb.Validator{
			PublicKey:             privKey.PublicKey().Marshal(),
			WithdrawalCredentials: make([]byte, 32),
		}
	}
	err = beaconState.SetValidators(validators)
	require.NoError(t, err)

	gen := mockstategen.NewMockService()
	gen.AddStateForRoot(beaconState, [32]byte{})
	return &Simulator{
		srvConfig: &ServiceConfig{
			Params:                      params,
			Database:                    slasherDB,
			AttestationStateFetcher:     &mock.ChainService{State: beaconState},
			PrivateKeysByValidatorIndex: privKeys,
			StateGen:                    gen,
		},
	}
}
