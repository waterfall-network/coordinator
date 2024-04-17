package fork_helper

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/fork"
)

func TestMainnet_Altair_UpgradeToAltair(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	fork.RunUpgradeToAltair(t, "mainnet")
}
