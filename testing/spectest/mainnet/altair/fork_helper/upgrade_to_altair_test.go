package fork_helper

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/altair/fork"
)

func TestMainnet_Altair_UpgradeToAltair(t *testing.T) {
	fork.RunUpgradeToAltair(t, "mainnet")
}
