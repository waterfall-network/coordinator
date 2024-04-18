package fork_transition

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/altair/fork"
)

func TestMainnet_Altair_Transition(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	fork.RunForkTransitionTest(t, "mainnet")
}
