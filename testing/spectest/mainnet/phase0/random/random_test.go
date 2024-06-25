package random

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/sanity"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

func TestMainnet_Phase0_Random(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	sanity.RunBlockProcessingTest(t, "test", "random/random/pyspec_tests")
}
