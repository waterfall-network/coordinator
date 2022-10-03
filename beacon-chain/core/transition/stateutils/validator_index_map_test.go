package stateutils_test

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/beacon-chain/core/transition/stateutils"
	v1 "github.com/waterfall-foundation/coordinator/beacon-chain/state/v1"
	fieldparams "github.com/waterfall-foundation/coordinator/config/fieldparams"
	"github.com/waterfall-foundation/coordinator/encoding/bytesutil"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestValidatorIndexMap_OK(t *testing.T) {
	base := &ethpb.BeaconState{
		Validators: []*ethpb.Validator{
			{
				PublicKey: []byte("zero"),
			},
			{
				PublicKey: []byte("one"),
			},
		},
	}
	state, err := v1.InitializeFromProto(base)
	require.NoError(t, err)

	tests := []struct {
		key [fieldparams.BLSPubkeyLength]byte
		val types.ValidatorIndex
		ok  bool
	}{
		{
			key: bytesutil.ToBytes48([]byte("zero")),
			val: 0,
			ok:  true,
		}, {
			key: bytesutil.ToBytes48([]byte("one")),
			val: 1,
			ok:  true,
		}, {
			key: bytesutil.ToBytes48([]byte("no")),
			val: 0,
			ok:  false,
		},
	}

	m := stateutils.ValidatorIndexMap(state.Validators())
	for _, tt := range tests {
		result, ok := m[tt.key]
		assert.Equal(t, tt.val, result)
		assert.Equal(t, tt.ok, ok)
	}
}
