package testing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

type getState func() (state.BeaconState, error)

func VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t *testing.T, factory getState) {
	st, err := factory()
	require.NoError(t, err)

	_, err = st.ValidatorAtIndexReadOnly(0)
	assert.Equal(t, state.ErrNilValidatorsInState, err)
}
