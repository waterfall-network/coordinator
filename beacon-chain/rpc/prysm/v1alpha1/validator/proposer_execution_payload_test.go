package validator

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

func TestServer_activationEpochNotReached(t *testing.T) {
	require.Equal(t, false, activationEpochNotReached(0))

	cfg := params.BeaconConfig()
	cfg.TerminalBlockHash = common.BytesToHash(bytesutil.PadTo([]byte{0x01}, 32))
	cfg.TerminalBlockHashActivationEpoch = 1
	params.OverrideBeaconConfig(cfg)

	require.Equal(t, true, activationEpochNotReached(0))
	require.Equal(t, false, activationEpochNotReached(params.BeaconConfig().SlotsPerEpoch+1))
}
