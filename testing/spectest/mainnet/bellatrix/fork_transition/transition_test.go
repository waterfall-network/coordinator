package fork_transition

import (
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/spectest/shared/bellatrix/fork"
)

func TestMainnet_Bellatrix_Transition(t *testing.T) {
	fork.RunForkTransitionTest(t, "mainnet")
}
