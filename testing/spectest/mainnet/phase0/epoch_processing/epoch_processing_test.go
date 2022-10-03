package epoch_processing

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
	"github.com/waterfall-foundation/coordinator/config/params"
)

func TestMain(m *testing.M) {
	prevConfig := params.BeaconConfig().Copy()
	defer params.OverrideBeaconConfig(prevConfig)
	c := params.BeaconConfig()
	c.MinGenesisActiveValidatorCount = 16384
	params.OverrideBeaconConfig(c)
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()

	m.Run()
}
