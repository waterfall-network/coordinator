package rewards

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/phase0/rewards"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

func TestMainnet_Phase0_Rewards(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	rewards.RunPrecomputeRewardsAndPenaltiesTests(t, "mainnet")
}
