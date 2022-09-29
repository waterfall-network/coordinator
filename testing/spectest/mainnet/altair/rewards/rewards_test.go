package rewards

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/config/features"
	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/rewards"
)

func TestMain(m *testing.M) {
	resetCfg := features.InitWithReset(&features.Flags{EnableBalanceTrieComputation: true})
	defer resetCfg()
	m.Run()
}

func TestMainnet_Altair_Rewards(t *testing.T) {
	rewards.RunPrecomputeRewardsAndPenaltiesTests(t, "mainnet")
}
