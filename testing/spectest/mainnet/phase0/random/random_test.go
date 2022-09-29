package random

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/phase0/sanity"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

func TestMainnet_Phase0_Random(t *testing.T) {
	sanity.RunBlockProcessingTest(t, "mainnet", "random/random/pyspec_tests")
}
