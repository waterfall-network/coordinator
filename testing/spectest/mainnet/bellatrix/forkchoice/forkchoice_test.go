package forkchoice

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/version"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/common/forkchoice"
)

func TestMainnet_Bellatrix_Forkchoice(t *testing.T) {
	forkchoice.Run(t, "mainnet", version.Bellatrix)
}

func TestMainnet_Bellatrix_Forkchoice_DoublyLinkTree(t *testing.T) {
	resetCfg := features.InitWithReset(&features.Flags{
		EnableForkChoiceDoublyLinkedTree: true,
	})
	defer resetCfg()
	forkchoice.Run(t, "mainnet", version.Bellatrix)
}
