package utils

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/waterfall-foundation/coordinator/config/params"
	"github.com/waterfall-foundation/coordinator/testing/require"
)

func TestConfig(t *testing.T) {
	require.NoError(t, SetConfig(t, "minimal"))
	require.Equal(t, types.Slot(8), params.BeaconConfig().SlotsPerEpoch)
	require.NoError(t, SetConfig(t, "mainnet"))
	require.Equal(t, types.Slot(32), params.BeaconConfig().SlotsPerEpoch)
}
