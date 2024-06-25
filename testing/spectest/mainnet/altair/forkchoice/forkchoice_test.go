package forkchoice

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/version"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/common/forkchoice"
)

func TestMainnet_Altair_Forkchoice(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	forkchoice.Run(t, "test", version.Altair)
}

func TestMainnet_Altair_Forkchoice_DoublyLinkTree(t *testing.T) {
	t.Skip() // Generate test data with pyton tool
	resetCfg := features.InitWithReset(&features.Flags{
		EnableForkChoiceDoublyLinkedTree: true,
	})
	defer resetCfg()
	forkchoice.Run(t, "test", version.Altair)
}
